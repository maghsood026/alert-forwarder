package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

type AlertManagerPayload struct {
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
	Alerts            []Alert           `json:"alerts"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
}

type EnQueue map[string]interface{}
type SMSRequestFormat struct {
	requestBody map[string][]EnQueue
}
type ResponseFormat map[string]string

const SMSWebhook string = "https://sms.org"

func setEnv() {

	envFile, err := os.Open("./.env")
	if err != nil {
		log.Fatalln(err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)

	for scanner.Scan() {
		envVar := strings.Split(scanner.Text(), "=")
		os.Setenv(envVar[0], envVar[1])
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

func fetchResponseSample(c *gin.Context) {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var data map[string]interface{}
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		log.Fatalln(err)
	}
	c.IndentedJSON(http.StatusOK, data)
}
func sendSMS(annotations map[string]string) map[string]string {
	data := map[string]interface{}{
		"enqueue": []map[string]interface{}{
			{
				"Destination": "9333333333",
				"Message":     "salam jigar",
				"Originator":  os.Getenv("ORG_NUM"),
			},
		},
	}
	postBody, _ := json.Marshal(data)
	requestBody := bytes.NewBuffer(postBody)

	var responseData map[string]string
	resp, err := http.Post(SMSWebhook, "application/json", requestBody)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	jsonErr := json.Unmarshal(body, &responseData)
	if jsonErr != nil {
		log.Fatalln(err)
	}
	return responseData
}
func sendRequestToTargetUser(c *gin.Context) {
	var request AlertManagerPayload
	var response map[string]string
	c.BindJSON(&request)
	annotations := request.Alerts[0].Annotations

	switch sendingMethod := annotations["sending-method"]; sendingMethod {
	case "sms":
		response = sendSMS(annotations)
		c.IndentedJSON(http.StatusOK, response) // test api!
	case "call":
		// call user will implement soon ...
	default:
		DefaultResponse := ResponseFormat{
			"message": "sending method not defined",
		}
		c.IndentedJSON(http.StatusBadRequest, DefaultResponse)
	}
}

func main() {
	setEnv()
	router := gin.Default()
	router.GET("/fetch-action", fetchResponseSample)
	router.POST("/request-action", sendRequestToTargetUser)
	router.Run(":8080")
}
