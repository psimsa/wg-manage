package main

import (
	"os"

	"github.com/ofcoursedude/wg-manage/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
