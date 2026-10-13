package buildinfo

var Version = "dev" //nolint:gochecknoglobals // set by `make build` through -ldflags -X, the linker only sets globals
