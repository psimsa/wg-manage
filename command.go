package main

type Command interface {
	PrintHelp()
	Run()
	ShortCommand() string
	LongCommand() string
}
