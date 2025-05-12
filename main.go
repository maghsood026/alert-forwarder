package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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

type ResponseFormat map[string]string

const SMSWebhook string = "http://sms-panel/notification"
const Originator string = "originator"

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

func checkResponseOfSmsPanel(responseData interface{}, phonenumber string) (map[string]string, int) {
	resbody, ok := responseData.(map[string]string)
	if !ok { // because the successful response has different format /:
		return map[string]string{
			"message": fmt.Sprintf("message sent to user %s successfully", phonenumber),
		}, 200
	}
	val, nok := resbody["error"]
	if nok {
		return map[string]string{
			"message": val,
		}, 400
	} else {
		return map[string]string{
			"message": fmt.Sprintf("message sent to user %s successfully", phonenumber),
		}, 200
	}

}
func requestToSmsPanel(data map[string]interface{}) (*http.Response, error) {
	postBody, _ := json.Marshal(data)
	requestBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(SMSWebhook, "application/json", requestBody)
	return resp, err
}
func failedResponse() (map[string]string, int) {
	return map[string]string{
		"message": "there was error when sending alert",
	}, 500
}

func convertListToArray(StringNumber string) []string {
	return strings.Split(StringNumber, ",")
}

func sendSMS(annotations map[string]string) (map[string]string, int) {
	var phonenumber string = fmt.Sprintf(annotations["number"])
	var description string = fmt.Sprintf(annotations["description"])
	var responseData interface{}
	var enqueues []map[string]interface{}

	numbers := convertListToArray(phonenumber)
	for _, number := range numbers {
		enqueues = append(enqueues,
			map[string]interface{}{
				"destination": number,
				"message":     description,
				"originator":  Originator,
			},
		)

	}
	data := map[string]interface{}{
		"enqueue": enqueues,
	}
	resp, err := requestToSmsPanel(data)

	if err != nil {
		return failedResponse()
	} else {
		body, err1 := io.ReadAll(resp.Body)
		jsonErr := json.Unmarshal(body, &responseData)
		if jsonErr == nil && err1 == nil {
			return checkResponseOfSmsPanel(responseData, phonenumber)
		} else {
			return failedResponse()

		}
	}
}
func sendRequestToTargetUser(c *gin.Context) {
	var request AlertManagerPayload
	var response map[string]string
	var statuscode int
	c.BindJSON(&request)
	annotations := request.Alerts[0].Annotations

	switch sendingMethod := annotations["sending-method"]; sendingMethod {
	case "sms":

		response, statuscode = sendSMS(annotations)

		c.IndentedJSON(statuscode, response) // test api!
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
	router := gin.Default()
	router.POST("/request-action", sendRequestToTargetUser)
	router.Run(":8086")
}
