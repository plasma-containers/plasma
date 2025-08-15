package controller

import (
	"log"
	"os"
	"time"

	"github.com/pgulb/plasma/container"
	"github.com/pgulb/plasma/db"
)

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
		var projects []db.Project
		err := db.DB.Find(&projects).Error
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Found", len(projects), "projects in db.")
		var services []db.Service
		err = db.DB.Find(&services).Error
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Found", len(services), "services in db.")
		for _, svc := range services {
			ctr, err := container.Get(svc.Name)
			if err != nil {
				log.Println(err)
			} else {
				present, alive, healthy := container.IsPresentAliveAndHealthy(&svc, ctr)
			}
		}
		time.Sleep(parsedInterval)
	}
}
