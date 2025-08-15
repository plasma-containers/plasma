package cli

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const usage = `Usage:
  plasma create -n <project-name> -c [optional] <compose-file>
	<compose-file> - default: docker-compose.yml
`

var baseURL string
var client *http.Client

func IsCLI() bool {
	splitted := strings.Split(os.Args[0], "/")
	executable := splitted[len(splitted)-1]
	if executable == "main" || executable == "plasma-server" {
		return false
	}
	return true
}

func initHttpClient() {
	client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func initBaseURL() {
	// TODO: read from env or some config file and provide localhost as default
	baseURL = "http://localhost:8080"
}

func Run() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	initHttpClient()
	initBaseURL()
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
