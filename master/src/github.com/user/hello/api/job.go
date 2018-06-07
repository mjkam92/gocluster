package api

import (
	"fmt"
	"github.com/user/hello/godag"
	"log"
	"time"
	_ "time"
)

type Volume struct {
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
}

type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DockerJob struct {
	JobName    string      `json:"jobName"`
	Cpus       float32     `json:"cpus"`
	Mem        int         `json:"mem"`
	Image      string      `json:"image"`
	Cmd        string      `json:"cmd"`
	Parameters []Parameter `json:"parameters"`
	Volumes    []Volume    `json:"volumes"`
	TargetIP   string      `json:"targetIp"`
	Status     string      `json:"status"`
	ParentList []string    `json:"parentList"`
}

type JobMsg struct {
	Cmd        string
	DockerJobs []DockerJob
}

type GDataContainer struct {
	JobNames      []string
	JobsContainer chan []DockerJob
}

type GMsg struct {
	Cmd           string
	Jobs          []DockerJob
	DataContainer GDataContainer
	From          string
	To            string
}

type RunnerMsg struct {
}

func graphManager(gMsgsChan <-chan GMsg) {
	dag := godag.NewGraph()

	for {
		select {
		case gMsg := <-gMsgsChan:
			graphMsgHandler(gMsg, dag)
		}
	}
}

func graphMsgHandler(gMsg GMsg, dag *godag.GoGraph) {
	if gMsg.Cmd == "AddV" {
		jobs := gMsg.Jobs

		for _, job := range jobs {
			dag.AddVertex(job.JobName, job)

			parentList := job.ParentList

			for _, parentJobName := range parentList {
				parentJob := dag.GetVertex(parentJobName)
				pJob, _ := parentJob.(DockerJob)
				dag.AddEdge(pJob.JobName, job.JobName)
			}
		}
	} else if gMsg.Cmd == "GetV" {
		dataContainer := gMsg.DataContainer
		jobNames := dataContainer.JobNames
		dockerJobsChan := dataContainer.JobsContainer
		dockerJobs := []DockerJob{}
		for _, jobName := range jobNames {
			job := dag.GetVertex(jobName)
			dockerJobs = append(dockerJobs, job.(DockerJob))
		}
		dockerJobsChan <- dockerJobs

	} else if gMsg.Cmd == "GetParents" {
		dataContainer := gMsg.DataContainer
		jobNames := dataContainer.JobNames
		dockerJobsChan := dataContainer.JobsContainer
		dockerJobs := []DockerJob{}
		for _, jobName := range jobNames {
			parents := dag.GetParents(jobName)

			for _, parent := range parents {
				castedParent := parent.(DockerJob)
				dockerJobs = append(dockerJobs, castedParent)
			}
		}
		dockerJobsChan <- dockerJobs
	}
}

func jobManager(jobMsgChan <-chan JobMsg, jobQueue chan string, gMsgChan chan<- GMsg) {
	for {
		select {
		case jobMsg := <-jobMsgChan:
			jobMsgHandler(jobQueue, jobMsg, gMsgChan)
		}
	}
}

func jobMsgHandler(jobQueue chan string, jobMsg JobMsg, gMsgChan chan<- GMsg) {

	if jobMsg.Cmd == "ADD" {
		jobList := jobMsg.DockerJobs
		logStr := ""
		for _, job := range jobList {
			jobQueue <- job.JobName
			logStr += (job.JobName + " ")
		}
		gMsg := GMsg{Cmd: "AddV", Jobs: jobMsg.DockerJobs}
		gMsgChan <- gMsg

		log.Printf("Job ADDED %s", logStr)

	} else {

	}

}

func jobRunner(jobQueue chan string, nodeMsgChan chan<- NodeMsg, gMsgChan chan<- GMsg) {
	//nodeList := []Node{}
	//notRunnableJobQueue := []DockerJob{}

	//delay := time.Second * 5
	//timer := time.NewTimer(delay)
	//case <-timer.C:
	//runnerChan <- jobQueue

	for {
		getNodeListChan := make(chan []Node)
		nodeMsg := NodeMsg{Cmd: "GetNodeList", NodeListChan: getNodeListChan}
		nodeMsgChan <- nodeMsg
		nodeList := <-getNodeListChan

		jobMap := make(map[string][]DockerJob)

		notRunnableJobQueue := make(chan string, 1000)

		jobNames := make(chan string, 1000)

		for {
			if len(jobQueue) == 0 {
				break
			}
			jobNames <- <-jobQueue
		}
		close(jobNames)

		for jobName := range jobNames {
			mainJobs := getJobFromGraphByJobName([]string{jobName}, gMsgChan)
			mainJob := mainJobs[0]
			runCheck := false

			parentJobList := getJobFromGraphByJobName(mainJob.ParentList, gMsgChan)
			if checkJobsFinished(parentJobList) == false {
				continue
			}

			if mainJob.TargetIP == "0.0.0.0" {
				for i, node := range nodeList {
					if checkResource(mainJob, node) == true {
						nodeList[i].UsingCpus += mainJob.Cpus
						nodeList[i].UsingMem += mainJob.Mem
						mainJob.TargetIP = node.IP
						jobList, exist := jobMap[node.IP]
						if !exist {
							jobMap[node.IP] = []DockerJob{mainJob}
						} else {
							jobList = append(jobList, mainJob)
						}
						runCheck = true
						break
					}
				}
				if runCheck == false {
					notRunnableJobQueue <- mainJob.JobName
				}

			} else {
				for i, node := range nodeList {
					if node.IP == mainJob.TargetIP {
						if checkResource(mainJob, node) == true {
							nodeList[i].UsingCpus += mainJob.Cpus
							nodeList[i].UsingMem += mainJob.Mem
							jobList, exist := jobMap[node.IP]
							if !exist {
								jobMap[node.IP] = []DockerJob{mainJob}
							} else {
								jobList = append(jobList, mainJob)
							}
							runCheck = true
							break
						}
					}
				}
				if runCheck == false {
					notRunnableJobQueue <- mainJob.JobName
				}
			}

		}

		resultDockerJobs := []DockerJob{}

		for ip, dockerJobs := range jobMap {
			endPoint := fmt.Sprintf("http://%s:5001/job", ip)
			for _, dockerJob := range dockerJobs {
				status, _ := sendNewJobRequest(endPoint, dockerJob)
				if status != "200 OK" {
					log.Println("job dispatch failed")
				} else {
					resultDockerJobs = append(resultDockerJobs, dockerJob)
				}
			}
		}

		if len(resultDockerJobs) > 0 {
			nodeMsgChan <- NodeMsg{Cmd: "RESOURCE_UPDATE", DockerJobs: resultDockerJobs}
		}

		for {
			if len(notRunnableJobQueue) == 0 {
				break
			}
			jobQueue <- <-notRunnableJobQueue
		}
		close(notRunnableJobQueue)

		//sendJob
		//job sucess fail, jobMap modify
		time.Sleep(5 * time.Second)
	}
}

func getJobFromGraphByJobName(jobNames []string, gMsgChan chan<- GMsg) []DockerJob {
	dockerJobs := []DockerJob{}

	dockerJobsChan := make(chan []DockerJob)
	gDataContainer := GDataContainer{JobNames: jobNames, JobsContainer: dockerJobsChan}
	gMsg := GMsg{Cmd: "GetV", DataContainer: gDataContainer}
	gMsgChan <- gMsg
	jobs := <-dockerJobsChan
	dockerJobs = append(dockerJobs, jobs...)

	return dockerJobs
}

func checkJobsFinished(parentJobList []DockerJob) bool {
	checkFlag := true
	for _, job := range parentJobList {
		if job.Status != "SUCCESS" {
			checkFlag = false
			break
		}
	}

	return checkFlag
}

func checkParents(job DockerJob) {

}

func checkResource(job DockerJob, node Node) bool {
	if (node.UsingCpus+job.Cpus <= node.TotalCpus) && (node.UsingMem+job.Mem < node.TotalMem) {
		return true
	}

	return false
}

/*
	if nodeMsg.CMD == "ADD" {
		*agentNodeList = append(*agentNodeList, nodeMsg.NODE)
	} else if nodeMsg.CMD == "UPDATE" {
		for i, node := range *agentNodeList {
			if node.IP == nodeMsg.NODE.IP {
				(*agentNodeList)[i].HostName = nodeMsg.NODE.HostName
				(*agentNodeList)[i].TotalCpus = nodeMsg.NODE.TotalCpus
				(*agentNodeList)[i].TotalMem = nodeMsg.NODE.TotalMem
				(*agentNodeList)[i].Status = nodeMsg.NODE.Status
			}
		}
	} else if nodeMsg.CMD == "STATUS_UPDATE" {
		for i, node := range *agentNodeList {
			if node.IP == nodeMsg.NODE.IP {
				(*agentNodeList)[i].Status = nodeMsg.NODE.Status
			}
		}
	} else if nodeMsg.CMD == "DEL" {
		*agentNodeList = delNode(*agentNodeList, nodeMsg.NODE)
	} else {
		log.Println("INVALID NODE CMD")
	}*/
