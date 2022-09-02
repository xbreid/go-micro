package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/rpc"
	"time"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "hit the broker",
	}

	_ = app.WriteJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJson(w, r, &requestPayload)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
	case "log":
		app.LogItemViaRPC(w, requestPayload.Log)
		// app.LogEventViaRabbit(w, requestPayload.Log)
		// app.LogItem(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		app.ErrorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) LogItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	loggerUrl := "http://logger-service:8082/log"
	request, err := http.NewRequest("POST", loggerUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling logger service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {
	// create payload to send to auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call service
	authUrl := "http://authentication-service:8081/authenticate"
	request, err := http.NewRequest("POST", authUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	// verify correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJson(w, errors.New("unauthorized"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling auth service"))
		return
	}

	// read response.Body
	var jsonFromService jsonResponse

	// decode response from auth service
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.ErrorJson(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	mailUrl := "http://mail-service:8083/send"
	request, err := http.NewRequest("POST", mailUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.ErrorJson(w, errors.New("error calling mail service"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.PushToQueue(l.Name, l.Data)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) PushToQueue(name string, message string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) LogItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = result

	app.WriteJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogItemViaGRPC(w http.ResponseWriter, r *http.Request) {
	var reqestPayload RequestPayload

	err := app.ReadJson(w, r, &reqestPayload)
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{LogEntry: &logs.Log{
		Name: reqestPayload.Log.Name,
		Data: reqestPayload.Log.Data,
	}})
	if err != nil {
		app.ErrorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.WriteJson(w, http.StatusAccepted, payload)
}
