# rebuild

`rebuild` is a command line tool to make it easy to automatically rebuild
projects when files change in a directory. e.g., go backend services or tools.

While this project was written to automatically rebuild my go projects,
`rebuild` doesn't make too many assumptions about commands. You can run this
with any commandline tool you want.

## Install

```sh
go get github.com/paydro/rebuild
```

## Usage

Rebuild and run a simple go tool:

```sh
rebuild --build 'go build -o mybinary .' -- mybinary
```

The `--build` flag is optional if you don't need a build step. For instance, if
you'd like to run tests after file changes:

```sh
rebuild -- go test ./...
```

`rebuild` also works well with projects that run continuously like HTTP servers.
In the following example, when a file changes in the current directory, the
`http` binary is built, and executed. When a new change occurs, `rebuild` will
properly kill the old `http` binary, rebuild the binary, and execute `http`
again.

```sh
rebuild --build 'go build -o http .' -- http
```

Exclude directories:

```sh
rebuild --exclude dir1 --exclude dir2 --exclude=dir3,dir4 -- go build .
```

Watch a different directory than CWD:

```sh
rebuild --watch path/to/files -- mybinary
```

### Stop `rebuild`

Type `CTRL+c`.

If your command handles termination signals, this will also send SIGTERM to the process.

