package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ofcoursedude/wg-manage/cmd/add"
	"github.com/ofcoursedude/wg-manage/cmd/bootstrap"
	"github.com/ofcoursedude/wg-manage/cmd/format"
	"github.com/ofcoursedude/wg-manage/cmd/generate"
	"github.com/ofcoursedude/wg-manage/cmd/initialize"
	"github.com/ofcoursedude/wg-manage/cmd/pool"
	"github.com/ofcoursedude/wg-manage/cmd/psk"
	"github.com/ofcoursedude/wg-manage/cmd/recreate"
	"github.com/ofcoursedude/wg-manage/cmd/remove"
	"github.com/ofcoursedude/wg-manage/cmd/tui"
	"github.com/ofcoursedude/wg-manage/cmd/version"
	"github.com/ofcoursedude/wg-manage/cmd/viz"
	"github.com/ofcoursedude/wg-manage/internal/buildinfo"
)

var availableCommands []Command

func main() {
	availableCommands = []Command{
		add.Add{},
		bootstrap.Bootstrap{},
		generate.Generate{},
		initialize.Initialize{},
		remove.Remove{},
		format.Format{},
		recreate.Recreate{},
		pool.Pool{},
		psk.Psk{},
		tui.Tui{},
		viz.Viz{},
		version.Version{},
	}

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	cmd := strings.ToLower(os.Args[1])
	for _, command := range availableCommands {
		if cmd == command.ShortCommand() || cmd == command.LongCommand() {
			command.Run()
			os.Exit(0)
		}
	}
	printHelp()
}

func printHelp() {
	fmt.Printf("wg-manage %s is a command line tool to centrally organize your wireguard configuration.\n", buildinfo.Version)
	for _, command := range availableCommands {
		command.PrintHelp()
	}
}
