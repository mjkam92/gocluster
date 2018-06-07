package api

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

type JobResult struct {
	IP      string
	JobName string
	Result  string
}

func runJob(job DockerJob) {

	dockerCmd := fmt.Sprintf("docker run --memory=%dM --cpus=%f", job.Mem, job.Cpus)

	for _, volume := range job.Volumes {
		volumeCmd := fmt.Sprintf(" -v %s:%s", volume.HostPath, volume.ContainerPath)
		dockerCmd += volumeCmd
	}

	for _, parameter := range job.Parameters {
		parameterCmd := fmt.Sprintf(" --%s=%s", parameter.Key, parameter.Value)
		dockerCmd += parameterCmd
	}

	dockerCmd += fmt.Sprintf(" %s sh -c %s", job.Image, job.Cmd)

	log.Println(dockerCmd)

	_, err := exec.Command("/bin/sh", "-c", dockerCmd).Output()

	masterIP := os.Getenv("masterIP")
	myIP := os.Getenv("myIP")
	endPoint := fmt.Sprintf("http://%s/job/result", masterIP)

	var jobResult JobResult

	if err != nil {
		jobResult = JobResult{IP: myIP, JobName: job.JobName, Result: "FAIL"}
	} else {
		jobResult = JobResult{IP: myIP, JobName: job.JobName, Result: "SUCCESS"}
	}
	SendData(endPoint, jobResult)
}
