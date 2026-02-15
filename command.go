package wgmanage

type Command interface {
	PrintHelp()
	Run()
	ShortCommand() string
	LongCommand() string
}
