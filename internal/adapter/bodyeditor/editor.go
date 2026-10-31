// Package bodyeditor manages temporary JSON files and external editor processes.
package bodyeditor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mvdan.cc/sh/v3/shell"

	"github.com/jzes/reqman/internal/config"
)

// Session also implements Bubble Tea's ExecCommand for terminal handoff.
type Session struct {
	ctx            context.Context
	mode           config.EditorMode
	dir, Path      string
	args           []string
	stdin          io.Reader
	stdout, stderr io.Writer
	Body           string
}

func New(ctx context.Context, mode config.EditorMode, editor, name, body string) (*Session, error) {
	// Match request command parsing: support quoted arguments, never execute
	// shell substitutions from EDITOR.
	args, err := shell.Fields(editor, func(string) string { return "" })
	if err != nil || len(args) == 0 || args[0] == "" {
		return nil, fmt.Errorf("EDITOR must contain an executable and optional arguments")
	}
	switch mode {
	case config.InLine, config.GUI, config.SidePanelTmux, config.SidePanelHerdr:
	default:
		return nil, fmt.Errorf("unsupported external editor mode %q", mode)
	}
	dir, err := os.MkdirTemp("", "reqman-*")
	if err != nil {
		return nil, err
	}
	name = filepath.Base(name)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if name == "" || name == "." {
		name = "request"
	}
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	return &Session{ctx: ctx, mode: mode, dir: dir, Path: path, args: append(args, path)}, nil
}

func (s *Session) SetStdin(r io.Reader)  { s.stdin = r }
func (s *Session) SetStdout(w io.Writer) { s.stdout = w }
func (s *Session) SetStderr(w io.Writer) { s.stderr = w }

func (s *Session) Run() error {
	defer os.RemoveAll(s.dir)
	var err error
	if s.mode == config.SidePanelTmux || s.mode == config.SidePanelHerdr {
		err = s.runPane()
	} else {
		cmd := exec.CommandContext(s.ctx, s.args[0], s.args[1:]...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = s.stdin, s.stdout, s.stderr
		err = cmd.Run()
	}
	if err != nil {
		return fmt.Errorf("external editor: %w", err)
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return fmt.Errorf("read edited body: %w", err)
	}
	if strings.TrimSpace(string(data)) != "" && !json.Valid(data) {
		return errors.New("edited body is not valid JSON; previous body preserved")
	}
	s.Body = string(data)
	return nil
}

func quote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }

func (s *Session) runPane() error {
	statusPath := filepath.Join(s.dir, "exit-status")
	parts := make([]string, len(s.args))
	for i, arg := range s.args {
		parts[i] = quote(arg)
	}
	// A marker waits for the editor, not for the multiplexer CLI to return.
	script := strings.Join(parts, " ") + "; result=$?; printf '%s' \"$result\" > " + quote(statusPath)
	var pane string
	if s.mode == config.SidePanelTmux {
		out, err := exec.CommandContext(s.ctx, "tmux", "split-window", "-h", "-P", "-F", "#{pane_id}", "sh -c "+quote(script)).Output()
		if err != nil {
			return fmt.Errorf("open tmux pane: %w", err)
		}
		pane = strings.TrimSpace(string(out))
		if pane == "" {
			return errors.New("tmux returned no pane ID")
		}
		defer exec.Command("tmux", "kill-pane", "-t", pane).Run()
	} else {
		out, err := exec.CommandContext(s.ctx, "herdr", "pane", "split", "--current", "--direction", "right", "--focus").Output()
		if err != nil {
			return fmt.Errorf("open herdr pane: %w", err)
		}
		var reply struct {
			Result struct {
				Pane struct {
					ID string `json:"pane_id"`
				} `json:"pane"`
			} `json:"result"`
		}
		if err := json.Unmarshal(out, &reply); err != nil {
			return err
		}
		pane = reply.Result.Pane.ID
		if pane == "" {
			return errors.New("herdr returned no pane ID")
		}
		defer exec.Command("herdr", "pane", "close", pane).Run()
		if out, err := exec.CommandContext(s.ctx, "herdr", "pane", "run", pane, "sh -c "+quote(script)).CombinedOutput(); err != nil {
			return fmt.Errorf("run editor in herdr: %w: %s", err, out)
		}
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	paneCheck := time.NewTicker(time.Second)
	defer paneCheck.Stop()
	for {
		data, err := os.ReadFile(statusPath)
		if err == nil && len(data) > 0 {
			code, err := strconv.Atoi(string(data))
			if err != nil {
				return err
			}
			if code != 0 {
				return fmt.Errorf("editor exited with status %d", code)
			}
			return nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case <-ticker.C:
		case <-paneCheck.C:
			var cmd *exec.Cmd
			if s.mode == config.SidePanelTmux {
				cmd = exec.CommandContext(s.ctx, "tmux", "display-message", "-p", "-t", pane, "#{pane_id}")
			} else {
				cmd = exec.CommandContext(s.ctx, "herdr", "pane", "get", pane)
			}
			if err := cmd.Run(); err != nil {
				// The editor may have finished between the marker read and
				// this check (tmux normally closes its pane on process exit).
				if data, _ := os.ReadFile(statusPath); len(data) > 0 {
					continue
				}
				return fmt.Errorf("editor pane closed before editing completed: %w", err)
			}
		}
	}
}
