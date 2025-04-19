package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

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

type SMSRequestFormat struct {
	Message string `json:"message"`
	Target  string `json:"target"` // destination phone number
}
type ResponseFormat map[string]string

const SMSWebhook string = "https://sms.org"

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
func sendSMS(annotations map[string]string) ResponseFormat {
	var responseData ResponseFormat

	smsInstance := SMSRequestFormat{
		Message: annotations["description"],
		Target:  annotations["target"],
	}

	postBody, _ := json.Marshal(smsInstance)
	requestBody := bytes.NewBuffer(postBody)

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
	var response ResponseFormat
	c.BindJSON(&request)
	annotations := request.Alerts[0].Annotations

	switch sendingMethod := annotations["sending-method"]; sendingMethod {
	case "sms":
		// response = sendSMS(annotations)
		// c.IndentedJSON(http.StatusOK, response)
		c.IndentedJSON(http.StatusOK, map[string]string{"response": "every this is okey"}) // test api!
	case "call":
		// call user will implement soon ...
	default:
		response = ResponseFormat{
			"message": "sending method not defined",
		}
		c.IndentedJSON(http.StatusBadRequest, response)
	}
}

func main() {
	router := gin.Default()
	router.GET("/fetch-action", fetchResponseSample)
	router.POST("/request-action", sendRequestToTargetUser)
	router.Run("localhost:8080")
}
