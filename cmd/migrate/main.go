package main

import (
	"os"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func main() {
	os.Exit(bootstrap.RunMigrate(os.Args[1:]))
}
