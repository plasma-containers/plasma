package container

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"slices"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	dcr "github.com/docker/docker/client"
)

var Docker *dcr.Client

func Init() error {
	var err error
	Docker, err = dcr.NewClientWithOpts(dcr.FromEnv, dcr.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func ParseCompose(projName string, cmps string) (*types.Project, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(cmps)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tmp, err := os.CreateTemp("", "compose-*.txt")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(decoded)
	tmp.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{tmp.Name()},
		cli.WithName(projName),
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return project, nil
}

func Get(name string) (*container.InspectResponse, error) {
	ctx := context.Background()
	ctrs, err := Docker.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var id string
	for _, ctr := range ctrs {
		if slices.Contains(ctr.Names, "/"+name) {
			id = ctr.ID
		}
	}
	if id == "" {
		return nil, nil
	}
	container, err := Docker.ContainerInspect(ctx, id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &container, nil
}
