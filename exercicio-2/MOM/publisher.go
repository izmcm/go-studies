package main

import (
	"encoding/json"
	// "github.com/masnun/gopher-and-rabbit"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var amqpURL string = "amqp://guest:guest@localhost:5672/"
var wg sync.WaitGroup

// 0 - add
// 1 - sub
// 2 - mul
// 3 - div
type AddTask struct {
	Number1   int
	Number2   int
	Operation int
	Opid      int
}

type Response struct {
	Number int
	Opid   int
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func createQueue(name string, channel *amqp.Channel) amqp.Queue {
	queue, err := channel.QueueDeclare(name, true, false, false, false, nil)
	handleError(err, fmt.Sprintf("Could not declare %s queue", name))
	return queue
}

func receiveData(receiveChannel <-chan amqp.Delivery) {
	stopChan := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		for d := range receiveChannel {
			log.Printf("Received a message: %s", d.Body)

			res := &Response{}

			err := json.Unmarshal(d.Body, res)

			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}

			// return the data
			log.Printf("Result of operation %d -> %d", res.Opid, res.Number)

			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}

		}
	}()

	// Stop for program termination
	<-stopChan
}

func main() {
	conn, err := amqp.Dial(amqpURL)
	handleError(err, "Can't connect to AMQP")
	defer conn.Close()

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer amqpChannel.Close()

	calculatorQueue := createQueue("calculator", amqpChannel)
	// ansQueue := createQueue("ans", amqpChannel)
	createQueue("ans", amqpChannel)

	// Random pra gerar os numeros pra colocar na fila
	rand.Seed(time.Now().UnixNano())

	addTask := AddTask{Number1: rand.Intn(999), Number2: rand.Intn(999), Operation: 0, Opid: rand.Intn(99999)}
	body, err := json.Marshal(addTask)
	if err != nil {
		handleError(err, "Error encoding JSON")
	}

	// TODO: Fazer o uso de json no outro middleware pra poder evitar problemas
	// na comparação.

	// Zona de publish
	err = amqpChannel.Publish("", calculatorQueue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	if err != nil {
		log.Fatalf("Error publishing message: %s", err)
	}

	// Agora precisa ver a resposta - AKA mete o code de consumer aqui

	log.Printf("AddTask: %d+%d", addTask.Number1, addTask.Number2)

}
