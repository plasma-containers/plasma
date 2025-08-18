package cli

import (
	"bytes"
	"context"
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
	"text/template"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/pgulb/plasma/db"
	logsv1 "github.com/pgulb/plasma/gen/logs/v1"
	"github.com/pgulb/plasma/gen/logs/v1/logsv1connect"
	"github.com/pgulb/plasma/server"
	"github.com/pgulb/plasma/version"
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

  plasma logs <container-name>
  - streams logs from plasma-server for <container-name> through gRPC

  plasma logs
  - streams logs for plasma-server itself
`

const wrongOrMissingParameters = "\nWrong or missing command parameters, check usage"

var baseURL string
var client *http.Client

//go:embed plasma-compose.yml
var plasmaCompose string

//go:embed plasma-compose.dev.yml
var plasmaComposeDev string

type QueryParams struct {
	Compose *string `json:"compose"`
	Project *string `json:"project"`
}

type verTpl struct {
	Version string
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

func composeDevOrTaggedVer() string {
	if version.Version == "develop" {
		color.Yellow("Using development compose file with build: .\n")
		return plasmaComposeDev
	}
	ver := verTpl{Version: version.Version}
	tpl, err := template.New("compose").Parse(plasmaCompose)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	var b bytes.Buffer
	err = tpl.Execute(&b, ver)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	return b.String()
}

func checkServerVer() {
	msg, status, err := reqDo("GET", "/version", nil)
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	if status != 200 {
		color.Red(msg.Msg)
		os.Exit(1)
	}
	if msg.Msg != version.Version {
		color.Red(
			fmt.Sprintf(
				"CLI version (%s) and server version (%s) must match.\n",
				version.Version,
				msg.Msg,
			),
		)
		os.Exit(1)
	}
}

func Run() {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	initHttpClient()
	initBaseURL()
	if len(os.Args) < 2 {
		color.Magenta(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		checkServerVer()
		projName := createCmd.String("n", "", "project name to create")
		composeFile := createCmd.String("c", "docker-compose.yml", "compose file to upload")
		createCmd.Parse(os.Args[2:])
		if projName == nil {
			color.Magenta(usage)
			color.Red(wrongOrMissingParameters)
			os.Exit(1)
		}
		if *projName == "" {
			color.Magenta(usage)
			color.Red(wrongOrMissingParameters)
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
		checkServerVer()
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
		fmt.Fprintln(w, "project\t|\tcreated_at\t")
		fmt.Fprintln(w, "---\t|\t---\t")
		for _, proj := range psResp.Projects {
			fmt.Fprintf(
				w,
				"%s\t|\t%s\t\n",
				proj.Name,
				proj.CreatedAt.Format(time.RFC3339),
			)
		}
		fmt.Fprintf(w, "\n")
		fmt.Fprintln(w, "svc\t|\tproj\t|\timg\t|\tstatus\t|\tports\t|\tmounts\t|\trestarts\t")
		fmt.Fprintln(w, "---\t|\t---\t|\t---\t|\t---\t|\t---\t|\t---\t|\t---\t")
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
			var ports string
			vols := 0
			var portsFromDB []db.PortInDB
			var volsFromDB []db.VolumeInDB
			if svc.Ports != nil {
				err := json.Unmarshal([]byte(*svc.Ports), &portsFromDB)
				if err != nil {
					color.Red(err.Error())
					os.Exit(1)
				}
				portsBuilder := strings.Builder{}
				for i, p := range portsFromDB {
					if p.HostIP == "" {
						portsBuilder.WriteString(fmt.Sprintf("%s:%v", p.Published, p.Target))
					} else {
						portsBuilder.WriteString(fmt.Sprintf("%v:%s:%v", p.HostIP, p.Published, p.Target))
					}
					if i < len(portsFromDB)-1 {
						portsBuilder.WriteString(", ")
					}
				}
				ports = portsBuilder.String()
			} else {
				ports = "-"
			}
			if svc.Volumes != nil {
				err := json.Unmarshal([]byte(*svc.Volumes), &volsFromDB)
				if err != nil {
					color.Red(err.Error())
					os.Exit(1)
				}
				vols = len(volsFromDB)
			}
			fmt.Fprintf(
				w,
				"%s\t|\t%s\t|\t%s\t|\t%s\t|\t%s\t|\t%v\t|\t%v\t\n",
				svc.Name,
				projName,
				svc.Image,
				ctrStatus,
				ports,
				vols,
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
		tempFile, err := os.CreateTemp("", "docker-compose.plasma.*.yml")
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		err = os.WriteFile(tempFile.Name(), []byte(composeDevOrTaggedVer()), 0644)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		defer func() {
			err = os.Remove(tempFile.Name())
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
		}()
		err = exec.Command("docker", "compose", "-f", tempFile.Name(), "--project-directory", ".", "up", "--build", "-d").
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
			"Then use 'plasma ps' to see if it's working.",
		)
	case "destroy":
		color.Magenta("Destroying local plasma...\n")
		tempFile, err := os.CreateTemp("", "docker-compose.plasma.*.yml")
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		err = os.WriteFile(tempFile.Name(), []byte(composeDevOrTaggedVer()), 0644)
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		defer func() {
			err = os.Remove(tempFile.Name())
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
		}()
		err = exec.Command("docker", "compose", "-f", tempFile.Name(), "--project-directory", ".", "down", "-v").
			Run()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Magenta("Plasma removed.")
	case "logs":
		checkServerVer()
		var ctrName string
		if len(os.Args) > 2 {
			ctrName = os.Args[2]
			if ctrName == "" {
				color.Magenta(usage)
				color.Red(wrongOrMissingParameters)
				os.Exit(1)
			}
		} else {
			ctrName = "plasma-server"
		}
		grpcClient := logsv1connect.NewLoggerServiceClient(
			client,
			"http://localhost:8081", // TODO: read from env or some config file
		)
		ctx := context.Background()
		stream, err := grpcClient.LogStream(ctx, connect.NewRequest(&logsv1.LogStreamRequest{
			Name: ctrName,
		}))
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
		for {
			more := stream.Receive()
			if !more {
				color.Magenta("log stream finished")
				if stream.Err() != nil {
					color.Red(stream.Err().Error())
					os.Exit(1)
				}
				break
			}

			// [8:] to remove stream metadata
			// https://pkg.go.dev/github.com/docker/docker/client#Client.ContainerLogs
			fmt.Println(string(stream.Msg().Message[8:]))
		}
	case "help":
		color.Magenta(usage)
	case "--help":
		color.Magenta(usage)
	case "-h":
		color.Magenta(usage)
	default:
		color.Magenta(usage)
		color.Red(fmt.Sprintf("unknown command '%s'", os.Args[1]))
		os.Exit(1)
	}
}
