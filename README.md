# reqman

Postman TUI

Reqman is a terminal-based application that allows you to manage and send HTTP requests, similar to Postman, but in a text-based interface.
It ahims to be simple, fast, file based, and vim motions compatible.

## Status

This project is currently in the early stages of development. It is not yet feature-complete and may have bugs or missing functionality. Contributions are welcome!

## Installation

Download the release archive for your platform from GitHub Releases, extract it, and move the `reqman` binary to a directory in your `PATH`.

Release artifacts are generated for:

- Linux x86_64
- macOS ARM64 (Apple Silicon / M1+)

## Releasing

Releases are created automatically when a semver tag is pushed:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs GoReleaser and uploads the generated archives and checksums to the GitHub release.

## V1 Features

- Send HTTP requests (GET, POST, PUT, DELETE, etc.)
- View response status, headers, and body
- Be file based, allowing you to save and load requests from files
- Use .curl files to define requests
- Allow for easy navigation and editing of requests using vim motions
- Support editing and managing multiple requests in a single session
- Support editing from a TUI interface with Postman like experience
