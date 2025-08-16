package container

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	dcr "github.com/docker/docker/client"
	"github.com/pgulb/plasma/db"
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
	ctrs, err := Docker.ContainerList(ctx, container.ListOptions{All: true})
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

func Run(svc *db.Service) error {
	ctx := context.Background()
	imgPresent, err := imagePresent(svc)
	if err != nil {
		log.Println(err)
		return err
	}
	if !imgPresent {
		err := imgPull(svc)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	var volsFromDB []db.VolumeInDB
	var binds []string
	if svc.Volumes != nil {
		err := json.Unmarshal([]byte(*svc.Volumes), &volsFromDB)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, v := range volsFromDB {
			binds = append(binds, v.Source+":"+v.Target)
		}
	}
	created, err := Docker.ContainerCreate(
		ctx,
		&container.Config{Image: svc.Image},
		&container.HostConfig{Binds: binds},
		&network.NetworkingConfig{},
		nil,
		svc.Name,
	)
	if err != nil {
		log.Println(err)
		return err
	}
	err = Docker.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func imgPull(svc *db.Service) error {
	ctx := context.Background()
	log.Println("Pulling image", svc.Image)
	closer, err := Docker.ImagePull(ctx, svc.Image, image.PullOptions{})
	if err != nil {
		log.Println(err)
		return err
	}
	defer closer.Close()
	importCloser, err := Docker.ImageImport(
		ctx,
		image.ImportSource{
			SourceName: "-",
			Source:     closer,
		},
		svc.Image,
		image.ImportOptions{},
	)
	if err != nil {
		// for some reason this is returned but it works
		if !strings.Contains(err.Error(), "invalid tar header") {
			log.Println(err)
			return err
		}
	}
	if importCloser != nil {
		defer importCloser.Close()
	}
	return nil
}

func imagePresent(svc *db.Service) (bool, error) {
	ctx := context.Background()
	images, err := Docker.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		log.Println(err)
		return false, err
	}
	for _, img := range images {
		if slices.Contains(img.RepoTags, svc.Image) {
			return true, nil
		}
	}
	return false, nil
}

func IsPresentAliveAndHealthy(
	svc *db.Service,
	ctr *container.InspectResponse,
) (bool, bool, bool) {
	if ctr == nil {
		//log.Println("Service", svc.Name, "is not present!")
		return false, false, false
	} else {
		if ctr.State.Paused || !ctr.State.Running || ctr.State.OOMKilled || ctr.State.Dead || ctr.State.Restarting {
			//log.Println("Service", svc.Name, "is stopped!")
			return true, false, false
		}
		if ctr.State.Health != nil {
			//log.Println("Service", svc.Name, "is running, status %s.", ctr.State.Health.Status)
			return true, true, ctr.State.Health.Status == "healthy"
		} else {
			//log.Println("Service", svc.Name, "is running, no healthcheck defined.")
			return true, true, true
		}
	}
}

func Volume(volName string) (bool, error) {
	ctx := context.Background()
	vols, err := Docker.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		log.Println(err)
		return false, err
	}
	for _, vol := range vols.Volumes {
		if vol.Name == volName {
			return true, nil
		}
	}
	return false, nil
}

func VolumeCreate(volName string) error {
	ctx := context.Background()
	_, err := Docker.VolumeCreate(ctx, volume.CreateOptions{Name: volName, Driver: "local"})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func Kill(ctrID string) error {
	ctx := context.Background()
	err := Docker.ContainerRemove(ctx, ctrID, container.RemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}
