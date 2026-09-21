<p align="center">
  <img src="https://github.com/user-attachments/assets/b50dc84f-f418-4d58-9e1f-25f030c8509f" alt="Reqman cover" width="900">
</p>

# reqman

Reqman is a terminal-based HTTP client for managing and sending requests from files. It aims to be simple, fast, file based, and friendly to vim-style navigation.

## Status

This project is in the early stages of development. It is not feature-complete yet and may have bugs or missing functionality.

## Installation

Download the release archive for your platform from GitHub Releases, extract it, and move the `reqman` binary to a directory in your `PATH`.

Release artifacts are generated for:

- Linux x86_64
- macOS ARM64 (Apple Silicon / M1+)

You can also install from source with Go:

```sh
go install github.com/jzes/reqman@latest
```

Or run it directly from a cloned repository:

```sh
go run . [requests-dir]
```

## Usage

Reqman loads requests from the directory passed as the first argument. If no directory is provided, it loads requests from the current directory.

```sh
reqman ./requests
```

Only files with the `.curl` extension are loaded. Each file represents one request, and the file name is used as the request name in the TUI.

Basic flow:

1. Create one or more `.curl` files in a requests directory.
2. Start reqman with `reqman ./requests`.
3. Select a request from the left panel.
4. Edit the method, URL, headers, or body.
5. Run the request with the Do button or with `:r`.
6. Save changes with `:w`.

## TUI Controls

Normal mode:

- `h`, `j`, `k`, `l`: move between panels or list items.
- `i`: edit the focused URL, headers, body, or response panel.
- `m`: open the method selector when the method panel is focused.
- `Enter` or `Space`: open the method selector when method is focused, or run the request when Do is focused.
- `a`: create a new request when the requests list is focused.
- `[` and `]`: switch response tabs when the response panel is focused.
- `:`: open the command prompt.
- `Esc`: leave insert mode, close selectors, or cancel an in-flight request.
- `Ctrl+C`: quit.

Command prompt:

- `:w`: save the current request.
- `:r`: run the current request.
- `:wr`: save and run the current request.
- `:wq`: save and quit.
- `:q`: quit.
- `:?`: open help.

## `.curl` Format

A request file must contain a shell-style `curl` command. Reqman parses the command and maps it to method, URL, headers, and body.

Supported fields:

- URL: the non-flag curl argument.
- Method: `-X`, `-XPOST`, `--request`, or `--request=POST`.
- Headers: `-H`, `-H'Name: Value'`, `--header`, or `--header='Name: Value'`.
- Body: `-d`, `--data`, `--data-raw`, `--data-binary`, `--data-ascii`, or `--json`.
- Redirect flag: `-L` or `--location` is accepted.

If a request has a body and no explicit method, Reqman uses `POST`. When sending a request, URLs without a scheme are sent with `http://` prepended.

When Reqman saves a request, it writes a single shell-quoted `curl` command and sorts headers before serializing.

## Examples

GET request:

```sh
curl -X GET -H 'Accept: application/json' https://api.example.com/books
```

POST request with JSON body:

```sh
curl -X POST -H 'Content-Type: application/json' -d '{"title":"Dune"}' https://api.example.com/books
```

Local request without scheme:

```sh
curl -X GET localhost:8080/health
```

## Known Limitations

- Reqman only loads `.curl` files.
- The `.curl` parser supports common curl flags, but it is not a full curl implementation.
- Unknown flags are ignored when possible.
- Environment variable expansion in `.curl` files is not supported.
- The app is still early-stage and may change between releases.

## Development

Run all tests:

```sh
go test ./...
```

Run the app locally:

```sh
go run . [requests-dir]
```

## Releasing

Releases are created automatically when a semver tag is pushed:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs GoReleaser and uploads the generated archives and checksums to the GitHub release.
