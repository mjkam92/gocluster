package main

import (
	_ "encoding/json"
	"flag"
	_ "fmt"
	"github.com/user/hello/api"
	"log"
	_ "net/http"
	"os"
)

func main() {
	masterIP := flag.String("master_ip", "", "Master IP")
	fileDir := flag.String("file_dir", "", "install file directory")
	pingInterval := flag.String("interval", "10", "ping interval")

	flag.Parse()

	if *masterIP == "" || *fileDir == "" {
		log.Println("Need Master IP")
		os.Exit(1)
	}

	os.Setenv("masterIP", *masterIP)
	os.Setenv("fileDir", *fileDir)
	os.Setenv("pingInterval", *pingInterval)

	//writeService(*masterIP, *fileDir)

	api.HttpServerRun()
}
