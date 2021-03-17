package main

import (
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/cmd/add"
	"github.com/ofcoursedude/wg-manage/cmd/bootstrap"
	"github.com/ofcoursedude/wg-manage/cmd/generate"
	"github.com/ofcoursedude/wg-manage/cmd/initialize"
	"github.com/ofcoursedude/wg-manage/cmd/remove"
)

func main() {

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}
	switch os.Args[1] {
	case "help", "h":
		printHelp()
		os.Exit(0)
	case "init", "i":
		initialize.Run()
	case "add", "a":
		add.Run()
	case "remove", "r":
		remove.Run()
	case "generate", "g":
		generate.Run()
	case "bootstrap", "b":
		bootstrap.Run()
	}
}

func printHelp() {
	fmt.Println("wg-manage is a command line tool to centrally organize your wireguard configuration.")
	bootstrap.PrintHelp()
	initialize.PrintHelp()
	generate.PrintHelp()
	add.PrintHelp()
	remove.PrintHelp()
}
