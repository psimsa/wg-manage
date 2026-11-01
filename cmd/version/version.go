package version

import (
	"fmt"

	"github.com/ofcoursedude/wg-manage/internal/buildinfo"
)

type Version struct{}

func (v Version) ShortCommand() string {
	return "v"
}

func (v Version) LongCommand() string {
	return "version"
}

func (v Version) Run() {
	fmt.Printf("wg-manage %s\n", buildinfo.Version)
}

func (v Version) PrintHelp() {
	fmt.Println("[version | v]")
	fmt.Println("\tPrints the wg-manage version")
}
