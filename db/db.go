package db

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Volume struct {
	gorm.Model
	Name      string
	ProjectId uint
}

type Service struct {
	gorm.Model
	Name                     string
	ProjectId                uint
	Command                  *string // originally []string, marshal as json string
	ContainerName            *string
	DependsOn                *string
	Entrypoint               *string // originally []string, marshal as json string
	Environment              *string // originally map[string]string, marshal as json string
	Expose                   *string // originally []string, marshal as json string
	Hostname                 *string
	HealthCheckCmd           *string // originally []string, marshal as json string
	HealthCheckTimeout       *time.Duration
	HealthCheckInterval      *time.Duration
	HealthCheckRetries       *uint64
	HealthCheckStartPeriod   *time.Duration
	HealthCheckStartInterval *time.Duration
	HealthCheckDisable       *bool
	Image                    string
	PullPolicy               *string
	ControllerKillCount      uint
}

type Project struct {
	gorm.Model
	Name string `gorm:"unique"`
}

func Init() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Migrating table projects...")
	err = DB.AutoMigrate(&Project{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Migrating table volumes...")
	err = DB.AutoMigrate(&Volume{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Migrating table services...")
	err = DB.AutoMigrate(&Service{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("SQLite database automigrated.")
	return nil
}

func servicesFromCompose(input *types.Project, projID uint) ([]*Service, error) {
	svcs := []*Service{}
	for _, svc := range input.Services {
		newSvc := Service{
			Name:                     svc.Name, // never empty
			ProjectId:                projID,   // never empty
			Command:                  nil,
			ContainerName:            nil,
			DependsOn:                nil,
			Entrypoint:               nil,
			Environment:              nil,
			Expose:                   nil,
			Hostname:                 nil,
			HealthCheckCmd:           nil,
			HealthCheckTimeout:       nil,
			HealthCheckInterval:      nil,
			HealthCheckRetries:       nil,
			HealthCheckStartPeriod:   nil,
			HealthCheckStartInterval: nil,
			HealthCheckDisable:       nil,
			Image:                    svc.Image, // probably never empty
			PullPolicy:               nil,
		}
		if svc.Command != nil {
			cmdBytes, err := json.Marshal(svc.Command)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			cmd := string(cmdBytes)
			newSvc.Command = &cmd
		}
		if svc.ContainerName != "" {
			newSvc.ContainerName = &svc.ContainerName
		}
		if svc.DependsOn != nil {
			depsKeys := make([]string, len(svc.DependsOn))
			i := 0
			for k := range svc.DependsOn {
				depsKeys[i] = k
				i++
			}
			depsBytes, err := json.Marshal(depsKeys)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			deps := string(depsBytes)
			newSvc.DependsOn = &deps
		}
		if svc.Entrypoint != nil {
			entrBytes, err := json.Marshal(svc.Entrypoint)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			entr := string(entrBytes)
			newSvc.Entrypoint = &entr
		}
		if svc.Environment != nil {
			envsBytes, err := json.Marshal(svc.Environment)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			envs := string(envsBytes)
			newSvc.Environment = &envs
		}
		if svc.Expose != nil {
			exposeBytes, err := json.Marshal(svc.Expose)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			expose := string(exposeBytes)
			newSvc.Expose = &expose
		}
		if svc.Hostname != "" {
			newSvc.Hostname = &svc.Hostname
		}
		if svc.HealthCheck != nil {
			if svc.HealthCheck.Test != nil {
				healthcheckCmd := strings.Join(svc.HealthCheck.Test, " ")
				newSvc.HealthCheckCmd = &healthcheckCmd
			}
			if svc.HealthCheck.Timeout != nil {
				timeout := time.Duration(*svc.HealthCheck.Timeout)
				newSvc.HealthCheckTimeout = &timeout
			}
			if svc.HealthCheck.Interval != nil {
				interval := time.Duration(*svc.HealthCheck.Interval)
				newSvc.HealthCheckInterval = &interval
			}
			if svc.HealthCheck.Retries != nil {
				retries := *svc.HealthCheck.Retries
				newSvc.HealthCheckRetries = &retries
			}
			if svc.HealthCheck.StartPeriod != nil {
				startPeriod := time.Duration(*svc.HealthCheck.StartPeriod)
				newSvc.HealthCheckStartPeriod = &startPeriod
			}
			if svc.HealthCheck.StartInterval != nil {
				startInterval := time.Duration(*svc.HealthCheck.StartInterval)
				newSvc.HealthCheckStartInterval = &startInterval
			}
			newSvc.HealthCheckDisable = &svc.HealthCheck.Disable
		}
		if svc.PullPolicy != "" {
			newSvc.PullPolicy = &svc.PullPolicy
		}
		svcs = append(svcs, &newSvc)
	}
	return svcs, nil
}

func volumesFromCompose(input *types.Project, projID uint) ([]*Volume, error) {
	vols := []*Volume{}
	for _, vol := range input.Volumes {
		vols = append(vols, &Volume{
			Name:      vol.Name,
			ProjectId: projID,
		})
	}
	return vols, nil
}

func splitCompose(input *types.Project, proj *Project) ([]*Service, []*Volume, error) {
	svcs, err := servicesFromCompose(input, proj.ID)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	vols, err := volumesFromCompose(input, proj.ID)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return svcs, vols, nil
}

func NewProjectToDB(input *types.Project) error {
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&Project{Name: input.Name}).Error; err != nil {
			return err
		}
		var proj *Project
		err := tx.Select("id", "name").Where("name = ?", input.Name).First(&proj).Error
		if err != nil {
			log.Println(err)
			return err
		}
		svcs, vols, err := splitCompose(input, proj)
		if err != nil {
			log.Println(err)
			return err
		}
		for _, svc := range svcs {
			if err := tx.Create(svc).Error; err != nil {
				return err
			}
		}
		for _, vol := range vols {
			if err := tx.Create(vol).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func UpKillCount(svc *Service) error {
	err := DB.Transaction(func(tx *gorm.DB) error {
		var svcFromDB *Service
		err := tx.Find(&svcFromDB, "Name = ?", svc.Name).Error
		if err != nil {
			log.Println(err)
			return err
		}
		svcFromDB.ControllerKillCount++
		if err := tx.Save(&svcFromDB).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}
