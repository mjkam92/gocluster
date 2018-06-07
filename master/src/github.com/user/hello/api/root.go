package api

import (
	"encoding/json"
	_ "github.com/starwander/goraph"
	_ "io/ioutil"
	"log"
	"net/http"
	_ "os"
	_ "path"
)

type HttpHandler struct {
	NodeMsgChan chan NodeMsg
	JobMsgChan  chan JobMsg
}

type NewNodeRequest struct {
	IP         string `json:"ip"`
	Port       string `json:"port"`
	AccessType string `json:"accessType"`
	PemPath    string `json:"pemPath"`
	Password   string `json:"password"`
	HostName   string `json:"hostName"`
}

type JobResult struct {
	IP      string
	JobName string
	Result  string
}

type NodeMsgRequest struct {
	Cmd   string
	Nodes []Node
}

func (httpHandler *HttpHandler) handleNodeAddRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newNodeRequest NewNodeRequest
	err := decoder.Decode(&newNodeRequest)
	if err != nil {
		log.Println("err")
	}

	defer r.Body.Close()

	//tempPath, err := ioutil.TempDir("/tmp", "ichthys")
	if err != nil {
		log.Fatal(err)
	}

	newNode := []Node{Node{IP: newNodeRequest.IP, Status: "installing"}}
	newNodeMsg := NodeMsg{Cmd: "ADD", Nodes: newNode}
	httpHandler.NodeMsgChan <- newNodeMsg

	go func() {
		//readyToCopy(newNodeRequest.IP, tempPath)
		//doSCP(newNodeRequest, tempPath)
		//doSSH(newNodeRequest, path.Base(tempPath))
		//defer os.RemoveAll(tempPath)
	}()
	w.WriteHeader(http.StatusOK)
}

//////////////////////////
func (httpHandler *HttpHandler) handleNodeValidate(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var nodeMsgRequest NodeMsgRequest
	err := decoder.Decode(&nodeMsgRequest)
	if err != nil {
		log.Println("aerr")
	}

	defer r.Body.Close()

	nodeMsg := NodeMsg{
		Cmd:          nodeMsgRequest.Cmd,
		Nodes:        nodeMsgRequest.Nodes,
		NodeListChan: nil,
		DockerJobs:   nil,
	}

	httpHandler.NodeMsgChan <- nodeMsg

	go healthMonitor(nodeMsg.Nodes, httpHandler.NodeMsgChan)

	w.WriteHeader(http.StatusOK)
}

////////////////////////
func (httpHandler *HttpHandler) handleNewJobRequest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var dockerJobs []DockerJob
	err := decoder.Decode(&dockerJobs)
	if err != nil {
		log.Println("aaerr")
	}

	log.Printf("%v", dockerJobs)

	defer r.Body.Close()

	jobMsg := JobMsg{Cmd: "ADD", DockerJobs: dockerJobs}

	httpHandler.JobMsgChan <- jobMsg

	w.WriteHeader(http.StatusOK)
}

func (httpHandler *HttpHandler) handleJobResult(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var jobResult JobResult
	err := decoder.Decode(&jobResult)
	if err != nil {
		log.Println("err")
	}

	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
}

func HttpServerRun() {

	sockMsgChan := make(chan SockMsg)
	nodeMsgChan := make(chan NodeMsg)
	jobMsgChan := make(chan JobMsg)
	gMsgChan := make(chan GMsg)

	jobQueue := make(chan string, 1000)

	server := getSocketIOServer(sockMsgChan)

	go nodeManager(nodeMsgChan, sockMsgChan)
	go webSocketManager(sockMsgChan)
	go jobManager(jobMsgChan, jobQueue, gMsgChan)
	go graphManager(gMsgChan)
	go jobRunner(jobQueue, nodeMsgChan, gMsgChan)

	httpHandler := HttpHandler{NodeMsgChan: nodeMsgChan, JobMsgChan: jobMsgChan}

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/node", httpHandler.handleNodeAddRequest)
	http.HandleFunc("/job", httpHandler.handleNewJobRequest)
	http.HandleFunc("/node/validate", httpHandler.handleNodeValidate)
	http.HandleFunc("/job/result", httpHandler.handleJobResult)
	http.ListenAndServe(":5000", nil)
}
