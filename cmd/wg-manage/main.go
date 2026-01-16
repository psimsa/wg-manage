package main

import (
	"os"

	wgmanage "github.com/ofcoursedude/wg-manage"
)

func main() {
	os.Exit(wgmanage.Run(os.Args[1:]))
}
