package main

import (
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/user/hello/api"
	"log"
	_ "net"
	"os"
)

type Volume struct {
	ContainerPath string
	HostPath      string
}

type Parameter struct {
	Key   string
	Value string
}

type DockerJob struct {
	JobName    string
	Cpus       float32
	Mem        int
	Image      string
	Cmd        string
	Parameters []Parameter
	Volumes    []Volume
	TargetIP   string
	Status     string
	ParentList []string
}

type Node struct {
	IP        string
	HostName  string
	TotalCpus float32
	TotalMem  int
	UsingCpus float32
	UsingMem  int
	Status    string
	JobList   []DockerJob
}

type NodeMsgRequest struct {
	Cmd   string
	Nodes []Node
}

func sendSigToMaster() {
	masterIP := os.Getenv("masterIP")
	myIP := os.Getenv("myIP")

	hostName, err := os.Hostname()
	if err != nil {
		log.Println("hostName error")
	}

	v, _ := mem.VirtualMemory()
	totalMem := v.Total / 1024 / 1024
	totalCpus, _ := cpu.Counts(true)

	node := Node{
		IP:        myIP,
		HostName:  hostName,
		TotalCpus: float32(totalCpus),
		TotalMem:  int(totalMem),
		Status:    "Healthy",
		JobList:   []DockerJob{},
	}

	nodes := []Node{node}

	nodeMsgRequest := NodeMsgRequest{
		Cmd:   "UPDATE",
		Nodes: nodes,
	}

	endPoint := fmt.Sprintf("http://%s/node/validate", masterIP)

	status, _ := api.SendData(endPoint, nodeMsgRequest)
	log.Println(status)

}

func main() {
	masterIP := flag.String("master_ip", "", "Master IP")
	myIP := flag.String("my_ip", "", "my ip")
	flag.Parse()

	os.Setenv("masterIP", *masterIP)
	os.Setenv("myIP", *myIP)

	go sendSigToMaster()

	api.HttpServerRun()
}
