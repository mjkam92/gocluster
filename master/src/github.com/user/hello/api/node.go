package api

import (
	_ "encoding/json"
	"fmt"
	"log"
	_ "os"
	_ "strconv"
	"time"
)

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

type NodeMsg struct {
	Cmd          string      `json:"cmd"`
	Nodes        []Node      `json:"nodes"`
	NodeListChan chan []Node `json:"nodeListChan"`
	DockerJobs   []DockerJob `json:"dockerJobs"`
}

func delNode(nodeList []Node, delNode Node) []Node {
	delIndex := -1

	//sList := sockList
	for i, node := range nodeList {
		if node.IP == delNode.IP {
			delIndex = i
		}
	}

	if delIndex != -1 {
		nodeList = append(nodeList[:delIndex], nodeList[delIndex+1:]...)
	}

	return nodeList
}

func nodeManager(nodeMsgChan <-chan NodeMsg, sockMsgChan chan<- SockMsg) {
	var nodeList = []Node{}

	for {
		select {
		case msg := <-nodeMsgChan:
			log.Printf("NodeManager MSG RECEIVE: %v\n", msg)
			nodeMsgHandler(&nodeList, msg)
			copyNodeList := make([]Node, len(nodeList))
			copy(copyNodeList, nodeList)
			if msg.Cmd != "GetNodeList" {
				sockMsgChan <- SockMsg{Cmd: "BROADCAST", NodeList: copyNodeList}
			}

			//nodeListChan2 <- nodeList
		}
	}

}

func nodeMsgHandler(agentNodeList *[]Node, nodeMsg NodeMsg) {
	if nodeMsg.Cmd == "ADD" {
		for _, node := range nodeMsg.Nodes {
			*agentNodeList = append(*agentNodeList, node)
		}
	} else if nodeMsg.Cmd == "UPDATE" {
		for i, node := range *agentNodeList {
			for _, newNode := range nodeMsg.Nodes {
				if node.IP == newNode.IP {
					(*agentNodeList)[i].HostName = newNode.HostName
					(*agentNodeList)[i].TotalCpus = newNode.TotalCpus
					(*agentNodeList)[i].TotalMem = newNode.TotalMem
					(*agentNodeList)[i].Status = newNode.Status
				}
			}
		}
	} else if nodeMsg.Cmd == "STATUS_UPDATE" {
		for i, node := range *agentNodeList {
			for _, newNode := range nodeMsg.Nodes {
				if node.IP == newNode.IP {
					(*agentNodeList)[i].Status = newNode.Status
				}
			}
		}
	} else if nodeMsg.Cmd == "DEL" {
		for _, node := range nodeMsg.Nodes {
			*agentNodeList = delNode(*agentNodeList, node)
		}
	} else if nodeMsg.Cmd == "GetNodeList" {
		copyNodeList := make([]Node, len(*agentNodeList))
		copy(copyNodeList, *agentNodeList)
		nodeMsg.NodeListChan <- copyNodeList
	} else if nodeMsg.Cmd == "RESOURCE_UPDATE" {
		for _, dockerJob := range nodeMsg.DockerJobs {
			for _, node := range *agentNodeList {
				if node.IP == dockerJob.TargetIP {
					node.UsingCpus += dockerJob.Cpus
					node.UsingMem += dockerJob.Mem
					node.JobList = append(node.JobList, dockerJob)
				}
			}
		}
	} else {
		log.Printf("INVALID MSG: %v\n", nodeMsg)
	}

}

func healthMonitor(nodes []Node, nodeMsgChan chan<- NodeMsg) {
	//pingIntervalInt, err := strconv.Atoi(os.Getenv("pingInterval"))
	pingIntervalTime := time.Duration(5 * time.Second)
	retryNum := 5

	/*
		if err != nil {
			log.Println("interval error")
		}*/

	node := nodes[0]

	endPoint := fmt.Sprintf("http://%s:5001/ping", node.IP)

	for {
		status := pingRequest(endPoint)

		if status != "200 OK" {
			log.Printf("Ping Failed to %s, Remain Retry Num: %d", node.IP, retryNum)

			if retryNum == 0 {
				unhealthyNode := []Node{Node{IP: node.IP, Status: "unhealthy"}}
				nodeMsg := NodeMsg{Cmd: "STATUS_UPDATE", Nodes: unhealthyNode}
				nodeMsgChan <- nodeMsg
				break
			} else {
				retryNum--
			}
		} else {
			retryNum = 5
		}
		log.Printf("Ping %s", node.IP)
		time.Sleep(pingIntervalTime)
	}
}
