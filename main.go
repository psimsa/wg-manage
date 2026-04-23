package wgmanage

import (
	"fmt"
	"strings"

	"github.com/ofcoursedude/wg-manage/cmd/add"
	"github.com/ofcoursedude/wg-manage/cmd/bootstrap"
	"github.com/ofcoursedude/wg-manage/cmd/format"
	"github.com/ofcoursedude/wg-manage/cmd/generate"
	"github.com/ofcoursedude/wg-manage/cmd/initialize"
	"github.com/ofcoursedude/wg-manage/cmd/recreate"
	"github.com/ofcoursedude/wg-manage/cmd/remove"
)

func AvailableCommands() []Command {
	return []Command{
		add.Add{},
		bootstrap.Bootstrap{},
		generate.Generate{},
		initialize.Initialize{},
		remove.Remove{},
		format.Format{},
		recreate.Recreate{},
	}
}

func Run(args []string) int {
	commands := AvailableCommands()
	if len(args) == 0 {
		printHelp(commands)
		return 0
	}

	cmd := strings.ToLower(args[0])
	for _, command := range commands {
		if cmd == command.ShortCommand() || cmd == command.LongCommand() {
			command.Run()
			return 0
		}
	}
	printHelp(commands)
	return 1
}

func printHelp(commands []Command) {
	fmt.Println("wg-manage is a command line tool to centrally organize your wireguard configuration.")
	for _, command := range commands {
		command.PrintHelp()
	}
}
