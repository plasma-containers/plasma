package cli

import (
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/pgulb/plasma/server"
)

const usage = `Usage:
  plasma create -n <project-name> -c [optional] <compose-file>
	<compose-file> - default: docker-compose.yml
  - creates a new project from a docker compose file
  - fails if project with this name already exists

  plasma ps
  - lists all plasma-managed resources

  plasma serve
  - deploys plasma-server to local docker

  plasma destroy
  - destroys plasma-server ran using 'plasma serve'
`

var baseURL string
var client *http.Client

//go:embed plasma-compose.yml
var plasmaCompose string

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
	case "ps":
		msg, status, err := reqDo("GET", "/ps", &QueryParams{})
		if err != nil {
			color.Magenta(msg.Msg)
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Magenta(fmt.Sprintf("HTTP status code %v", status))
		if status != 200 {
			color.Red(msg.Msg)
			os.Exit(1)
		}
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		var psResp server.PsResp
		err = json.Unmarshal([]byte(msg.Msg), &psResp)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		fmt.Fprintln(w, "project\t|\tcreated_at\t|\tupdated_at\t")
		fmt.Fprintln(w, "---\t|\t---\t|\t---\t")
		for _, proj := range psResp.Projects {
			fmt.Fprintf(
				w,
				"%s\t|\t%s\t|\t%s\t\n",
				proj.Name,
				proj.CreatedAt.Format(time.RFC3339),
				proj.UpdatedAt.Format(time.RFC3339),
			)
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintln(w, "svc\t|\tproject\t|\timage\t|\tstatus\t|\ttimes_killed\t")
		fmt.Fprintln(w, "---\t|\t---\t|\t---\t|\t---\t|\t---\t")
		for _, svc := range psResp.Services {
			var ctrStatus string
			for _, s := range psResp.Statuses {
				if s.Name == svc.Name {
					ctrStatus = s.Status
				}
			}
			projName := ""
			for _, p := range psResp.Projects {
				if p.ID == svc.ProjectId {
					projName = p.Name
				}
			}
			fmt.Fprintf(
				w,
				"%s\t|\t%s\t|\t%s\t|\t%s\t|\t%v\t\n",
				svc.Name,
				projName,
				svc.Image,
				ctrStatus,
				svc.ControllerKillCount,
			)
		}
		err = w.Flush()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
	case "serve":
		color.Magenta("Deploying plasma...\n")
		err := os.WriteFile("docker-compose.plasma.yml", []byte(plasmaCompose), 0644)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		err = exec.Command("docker", "compose", "-f", "docker-compose.plasma.yml", "up", "--build", "-d").
			Run()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Magenta("Plasma deployed on local docker.")
		color.Magenta(
			"Use 'plasma create -n <project-name> -c <compose-file>' to create a new project.",
		)
		color.Magenta(
			"Then 'use plasma ps' to see if it's working.",
		)
		err = os.Remove("docker-compose.plasma.yml")
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
	case "destroy":
		color.Magenta("Destroying local plasma...\n")
		err := os.WriteFile("docker-compose.plasma.yml", []byte(plasmaCompose), 0644)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		err = exec.Command("docker", "compose", "-f", "docker-compose.plasma.yml", "down", "-v").
			Run()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Magenta("Plasma removed.")
		err = os.Remove("docker-compose.plasma.yml")
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
	}
}
