package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pgulb/plasma/container"
	"github.com/pgulb/plasma/db"
)

type RespMsg struct {
	Msg string `json:"msg"`
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write(Msg(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(Msg(fmt.Sprintf("Project '%s' created", projName)))
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

	log.Fatal(http.ListenAndServe(":8080", mux))
}
