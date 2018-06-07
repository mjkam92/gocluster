package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

const (
	dockerServiceTemplate = `[Unit]
Description=Docker-Engine
[Service]
Restart=always
StartLimitInterval=0
RestartSec=15
ExecStartPre=-/sbin/ip link del docker0
ExecStart=/usr/bin/dockerd -H unix:///var/run/docker.sock`

	goServiceTemplate = `[Unit]
Description=Go-Main-Server
[Service]
Restart=always
StartLimitInterval=0
RestartSec=15
ExecStartPre=-/sbin/ip link del docker0
ExecStart=/usr/bin/GoAgent -master_ip=%s -my_ip=%s`

	installShTemplate = `#!/bin/bash
sudo su
tar -xf docker-18.03.1-ce.tgz
cp -rf docker/* /usr/bin

mv dockerd.service /etc/systemd/system
mv goagent.service /etc/systemd/system
mv GoAgent /usr/bin

systemctl enable dockerd.service
systemctl enable goagent.service

systemctl start dockerd
systemctl start goagent`
)

func readyToCopy(targetIP string, tempPath string) {

	//msg := fmt.Sprintf(`The operation failed on the %d iteration.
	//	Resumed on the %d iteration.`, 2, 3)

	masterIP := os.Getenv("masterIP")
	fileDir := os.Getenv("fileDir")

	dockerPath := path.Join(fileDir, "docker-18.03.1-ce.tgz")
	goAgentPath := path.Join(fileDir, "GoAgent")
	tarFileName := path.Base(tempPath) + ".tar.gz"

	dockerService := fmt.Sprintf(dockerServiceTemplate)
	goService := fmt.Sprintf(goServiceTemplate, masterIP, targetIP)
	installScript := fmt.Sprintf(installShTemplate)

	err := ioutil.WriteFile(path.Join(tempPath, "dockerd.service"), []byte(dockerService), 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(path.Join(tempPath, "goagent.service"), []byte(goService), 0644)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(path.Join(tempPath, "install.sh"), []byte(installScript), 0777)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	unTarCmd := fmt.Sprintf("tar -xf %s -C %s && cp %s %s && cd /tmp && tar -zcf %s %s", dockerPath, tempPath, goAgentPath, tempPath, tarFileName, path.Base(tempPath))

	log.Println(unTarCmd)
	_, err = exec.Command("sh", "-c", unTarCmd).Output()
	if err != nil {
		log.Println("untar failed")
	}
}
