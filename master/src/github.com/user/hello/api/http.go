package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func pingRequest(host string) string {

	resp, err := http.Get(host)
	if err != nil {
		log.Println("http Get request Failed")
		return "404"
	}

	defer resp.Body.Close()
	return resp.Status

}

func sendNewJobRequest(host string, data interface{}) (string, []byte) {
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
