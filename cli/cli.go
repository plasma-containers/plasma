package cli

import (
	"log"
	"os"
	"strings"
)

func IsCLI() bool {
	splitted := strings.Split(os.Args[0], "/")
	executable := splitted[len(splitted)-1]
	if executable == "main" || executable == "plasma-server" {
		return false
	}
	return true
}

func Run() {
	log.Println("oOoOo Starting plasma-cli oOoOo")
}
