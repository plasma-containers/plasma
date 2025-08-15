package controller

import (
	"log"
	"os"
	"time"

	"github.com/pgulb/plasma/container"
	"github.com/pgulb/plasma/db"
)

func svcLoop(services []db.Service) {
	for _, svc := range services {
		log.Println("-")
		log.Printf("Checking service '%s'\n", svc.Name)
		if svc.Image == "" {
			log.Println("Plasma does not handle 'build' image services.")
			log.Println("Service", svc.Name, "has no image, skipping.")
			continue
		}
		ctr, err := container.Get(svc.Name)
		if err != nil {
			log.Println(err)
			log.Println("Going to next service.")
			continue
		} else {
			present, alive, healthy := container.IsPresentAliveAndHealthy(&svc, ctr)
			var wasJustRan bool
			if !present {
				log.Println("Service", svc.Name, "is not present!")
				log.Println("Trying to run it...")
				err := container.Run(&svc)
				if err != nil {
					log.Println(err)
					log.Println("Going to next service.")
					continue
				}
				wasJustRan = true
			} else {
				log.Println("Service", svc.Name, "is present.")
			}
			if wasJustRan {
				log.Println("Service", svc.Name, "started, going to next.")
				continue
			}
			if !alive {
				log.Println("Service", svc.Name, "is not running!")
			} else {
				log.Println("Service", svc.Name, "is running.")
			}
			if !healthy {
				log.Println("Service", svc.Name, "is not healthy!")
			} else {
				log.Println("Service", svc.Name, "is healthy.")
			}
		}
	}
}

func GetResources() ([]db.Service, error) {
	var projects []db.Project
	err := db.DB.Find(&projects).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Found", len(projects), "projects in db.")
	var services []db.Service
	err = db.DB.Find(&services).Error
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Found", len(services), "services in db.")
	return services, nil
}

func Run() {
	log.Println("oOoOo Starting plasma-controller oOoOo")
	interval := os.Getenv("PLASMA_CONTROLLER_INTERVAL")
	if interval == "" {
		log.Println("PLASMA_CONTROLLER_INTERVAL is not set, defaulting to 10s.")
		interval = "10s"
	}
	parsedInterval, err := time.ParseDuration(interval)
	if err != nil {
		log.Println("PLASMA_CONTROLLER_INTERVAL is not valid duration")
		log.Fatal(err)
	}
	log.Println("Initializing docker client...")
	err = container.Init()
	if err != nil {
		log.Println("An error occurred while initializing docker client.")
		log.Fatal(err)
	}
	for {
		log.Println("---")
		services, err := GetResources()
		if err != nil {
			continue
		}
		svcLoop(services)
		time.Sleep(parsedInterval)
	}
}
