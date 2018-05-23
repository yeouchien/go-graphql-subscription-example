package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	mutation := `
		mutation createData($input: CreateDataInput!) {
			createData(input: $input) {
				timestamp
				value
				deviceId
			}
		}
	`
	variables := `{
		"input": {
			"timestamp": %v,
			"value": %v,
			"deviceId": "device-id"
		}
	}`

	// wait opentsdb startup
	<-time.After(1 * time.Minute)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		value := 5.0 + rand.Float64()*(100.0-5.0)
		query := fmt.Sprintf(`{
			"query": %v,
			"variables": %v
		}`,
			strconv.QuoteToASCII(mutation),
			fmt.Sprintf(variables, time.Now().Unix(), value),
		)

		if err := makeRequest(query); err != nil {
			log.Printf("error publishing: %v", err)
		}
	}
}

func makeRequest(requestString string) error {
	var str = []byte(requestString)
	url := os.Getenv("SERVER_URL")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(str))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Status: %v, Body: %v", resp.Status, string(body))

	return nil
}
