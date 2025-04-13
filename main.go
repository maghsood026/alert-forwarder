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

type WebhookPayload struct {
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
	Webhook string `json:"webhook"`
	Message string `json:"message"`
	Target  string `json:"target"` // destination phone number
}

func formatToCustomPayload(alert Alert) SMSRequestFormat {
	alertLabels := alert.Labels
	alertAnnotations := alert.Annotations

	alertName := alertLabels["alertname"]
	description := alertAnnotations["description"]
	responseMsg := SMSRequestFormat{
		Message: alertName + description,
		Target:  "+989333333333",
	}
	return responseMsg
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
func sendRequestToSMSPanel(c *gin.Context) {

	smsInstance := SMSRequestFormat{
		Webhook: "https://postman-echo.com/post",
		Message: "you have alert",
		Target:  "mgh.esmaeili",
	}
	postBody, _ := json.Marshal(smsInstance)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(smsInstance.Webhook, "application/json", responseBody)
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

func main() {
	router := gin.Default()
	router.GET("/fetch-action", fetchResponseSample)
	router.POST("/request-action", sendRequestToSMSPanel)
	router.Run("localhost:8080")
}
