package cli

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pgulb/plasma/server"
)

const usage = `Usage:
  plasma create -n <project-name> -c [optional] <compose-file>
	<compose-file> - default: docker-compose.yml
`

var baseURL string
var client *http.Client

type QueryParams struct {
	Compose *string `json:"compose"`
	Project *string `json:"project"`
}

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

func reqDo(method string, url string, qp *QueryParams) (*server.RespMsg, int, error) {
	req, err := http.NewRequest(method, baseURL+url, nil)
	if err != nil {
		return &server.RespMsg{}, 0, err
	}
	q := req.URL.Query()
	if qp.Compose != nil {
		q.Add("compose", *qp.Compose)
	}
	if qp.Project != nil {
		q.Add("project", *qp.Project)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return &server.RespMsg{}, 0, err
	}
	defer resp.Body.Close()
	var msg server.RespMsg
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return &server.RespMsg{}, 0, err
	}
	err = json.Unmarshal(content, &msg)
	if err != nil {
		return &server.RespMsg{}, 0, err
	}
	return &msg, resp.StatusCode, nil
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
		projName := createCmd.String("n", "", "project name to create")
		composeFile := createCmd.String("c", "docker-compose.yml", "compose file to upload")
		createCmd.Parse(os.Args[2:])
		if projName == nil {
			fmt.Print(usage)
			os.Exit(1)
		}
		if *projName == "" {
			fmt.Print(usage)
			os.Exit(1)
		}
		color.Magenta(fmt.Sprintf("Creating project %s...\n\n", *projName))
		compoBytes, err := os.ReadFile(*composeFile)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		composeB64 := base64.RawURLEncoding.EncodeToString(compoBytes)
		msg, status, err := reqDo(
			"POST",
			"/create",
			&QueryParams{Compose: &composeB64, Project: projName},
		)
		if err != nil {
			color.Magenta(msg.Msg)
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Magenta(fmt.Sprintf("HTTP status code %v", status))
		if status != 201 {
			color.Red(msg.Msg)
			os.Exit(1)
		}
		color.Magenta(msg.Msg)
	}
}
