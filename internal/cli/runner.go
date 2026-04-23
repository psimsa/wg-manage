package cli

import (
	"fmt"
	"strings"

	wgmanage "github.com/ofcoursedude/wg-manage"
)

func Run(args []string) int {
	commands := wgmanage.AvailableCommands()
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

func printHelp(commands []wgmanage.Command) {
	fmt.Println("wg-manage is a command line tool to centrally organize your wireguard configuration.")
	for _, command := range commands {
		command.PrintHelp()
	}
}
