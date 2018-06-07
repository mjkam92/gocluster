package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	_ "log"
	_ "net"
	"net/http"
)

func SendData(host string, data interface{}) (string, []byte) {
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest("POST", host, buff)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return resp.Status, body
}

/*
func GetMyIP(ni string) string {
	ifaces, err := net.Interfaces()

	if err != nil {
		log.Println("netinterface error")
	}

	// handle err
	for _, i := range ifaces {
		if ni == i.Name {
			addrs, _ := i.Addrs()
			return addrs[0].String()
		}
	}
	return "good"
}*/
