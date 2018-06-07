package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type NewNodeRequest struct {
	IP         string `json:"ip"`
	AccessType string `json:"accessType"`
	PemPath    string `json:"pemPath"`
	Password   string `json:"password"`
	HostName   string `json:"hostName"`
	FileSource string `json:"fileSource"`
	FileDest   string `json:"fileDest"`
}

func handleNodeAddRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newNodeRequest NewNodeRequest
	err := decoder.Decode(&newNodeRequest)
	if err != nil {
		log.Println("err")
	}

	defer r.Body.Close()

}

func handleNewJobRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newDockerJob DockerJob
	err := decoder.Decode(&newDockerJob)
	if err != nil {
		log.Println("err")
	}

	log.Printf("Run Job %v\n", newDockerJob)

	go runJob(newDockerJob)

	defer r.Body.Close()
	w.WriteHeader(http.StatusOK)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HttpServerRun() {

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/node", handleNodeAddRequest)
	http.HandleFunc("/job", handleNewJobRequest)
	http.HandleFunc("/ping", handlePing)
	http.ListenAndServe(":5001", nil)
}
