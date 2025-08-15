package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

const usage = `Usage:
  plasma create -n <project-name> -c [optional] <compose-file>
	<compose-file> - default: docker-compose.yml
`

func IsCLI() bool {
	splitted := strings.Split(os.Args[0], "/")
	executable := splitted[len(splitted)-1]
	if executable == "main" || executable == "plasma-server" {
		return false
	}
	return true
}

func Run() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])
		color.Magenta("Creating project...")
	}
}
