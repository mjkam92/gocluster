package api

import (
	"bytes"
	"fmt"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
)

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func getSession(accessType string, pemPath string, password string, hostName string, ip string) *ssh.Session {

	var config *ssh.ClientConfig

	if accessType == "pem" {
		myPem := publicKeyFile(pemPath)
		config = &ssh.ClientConfig{
			User: hostName,
			Auth: []ssh.AuthMethod{
				myPem,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else if accessType == "pwd" {
		pwd := ssh.Password(password)
		config = &ssh.ClientConfig{
			User: hostName,
			Auth: []ssh.AuthMethod{
				pwd,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {

	}

	client, err := ssh.Dial("tcp", ip, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	return session
}

func doSCP(reqData NewNodeRequest, tempPath string) {
	addr := reqData.IP + ":" + reqData.Port
	session := getSession(reqData.AccessType, reqData.PemPath, reqData.Password, reqData.HostName, addr)

	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	tarFileName := path.Base(tempPath) + ".tar.gz"
	srcPath := fmt.Sprintf("/tmp/%s", tarFileName)
	var destPath string
	if reqData.HostName == "root" {
		destPath = path.Join("/root", tarFileName)
	} else {
		destPath = path.Join("/home", reqData.HostName, tarFileName)
	}

	err := scp.CopyPath(srcPath, destPath, session)
	if err != nil {
		log.Fatal("Failed to scp for ")
	}

	delCmd := fmt.Sprintf("rm %s", srcPath)

	_, err = exec.Command("sh", "-c", delCmd).Output()
	if err != nil {
		log.Println("Del Tar Failed")
	}
	log.Println("Finish SCP : " + reqData.IP)

}

func doSSH(reqData NewNodeRequest, dirName string) {
	addr := reqData.IP + ":" + reqData.Port
	session := getSession(reqData.AccessType, reqData.PemPath, reqData.Password, reqData.HostName, addr)
	defer session.Close()

	fileName := dirName + ".tar.gz"

	cmd := fmt.Sprintf("sudo su && tar -zxf %s && cd %s && bash install.sh", fileName, dirName)
	log.Println(cmd)

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		panic("Failed to run: " + err.Error())
	}

	log.Println("SSH Install FINISH")
}
