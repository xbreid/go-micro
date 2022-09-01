package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

type Consumer struct {
	Conn      *amqp.Connection
	QueueName string
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		Conn: conn,
	}

	err := consumer.Setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) Setup() error {
	channel, err := consumer.Conn.Channel()
	if err != nil {
		return err
	}

	return DeclareExchange(channel)
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := DeclareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go HandlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)

	<-forever

	return nil
}

func HandlePayload(payload Payload) {
	// you can have as many cases as desired, just needs logic
	switch payload.Name {
	case "log", "event":
		// log whatever we get
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		// authenticate
	default:
		err := LogEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func LogEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	loggerUrl := "http://logger-service:8082/log"
	request, err := http.NewRequest("POST", loggerUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
