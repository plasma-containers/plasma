package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pgulb/plasma/container"
	"github.com/pgulb/plasma/db"
)

type RespMsg struct {
	Msg string `json:"msg"`
}

type CtrStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PsResp struct {
	Projects []db.Project `json:"projects"`
	Services []db.Service `json:"services"`
	Volumes  []db.Volume  `json:"volumes"`
	Statuses []CtrStatus  `json:"statuses"`
}

func Msg(msg string) []byte {
	b, _ := json.Marshal(RespMsg{Msg: msg})
	return b
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Create(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cmps := q.Get("compose")
	if cmps == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(Msg("compose param is required"))
		return
	}
	projName := q.Get("project")
	if projName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(Msg("project param is required"))
		return
	}

	project, err := container.ParseCompose(projName, cmps)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(Msg(err.Error()))
		return
	}

	err = db.NewProjectToDB(project)
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			w.WriteHeader(http.StatusConflict)
			w.Write(Msg(fmt.Sprintf("Project '%s' already exists", projName)))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(Msg(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(Msg(fmt.Sprintf("Project '%s' created", projName)))
}

func Ps(w http.ResponseWriter, r *http.Request) {
	projs, svcs, vols, err := db.Ps()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(Msg(err.Error()))
		return
	}
	statuses := []CtrStatus{}
	for _, svc := range svcs {
		// TODO: change to ContainerList, will be probably faster and less prone to
		// listing containers that were just killed by controller
		ctr, err := container.Get(svc.Name)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(Msg(err.Error()))
			return
		}
		if ctr == nil {
			statuses = append(statuses, CtrStatus{Name: svc.Name, Status: "unknown"})
		}
		if ctr.State == nil {
			statuses = append(statuses, CtrStatus{Name: svc.Name, Status: "unknown"})
		} else {
			statuses = append(statuses, CtrStatus{Name: svc.Name, Status: ctr.State.Status})
		}
	}
	psResp := PsResp{Projects: projs, Services: svcs, Volumes: vols, Statuses: statuses}
	b, err := json.Marshal(psResp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(Msg(err.Error()))
		return
	}
	w.Write(Msg(string(b)))
}

func Run() {
	log.Println("oOoOo Starting plasma-server oOoOo")

	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.Handle("GET /healthz", LoggerMiddleware(http.HandlerFunc(Health)))
	mux.Handle("POST /create", LoggerMiddleware(http.HandlerFunc(Create)))
	mux.Handle("GET /ps", LoggerMiddleware(http.HandlerFunc(Ps)))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
