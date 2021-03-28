package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
)

type Payload struct {
	Endpoint string   `json:"end_point"`
	Body     string   `json:"body"`
	Headers  []string `json:"headers"`
}

var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
var AWS_ACCESS_KEY_ID = os.Getenv("AWS_ACCESS_KEY_ID")
var AWS_REGION = os.Getenv("AWS_REGION")

func handleRequests(port int) {
	log.Infof("Sender started on port %d...", port)
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/{queue_name}/{end_point}", sendToSqs)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), myRouter))
}

func sendToSqs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queueName := vars["queue_name"]
	endPoint := vars["end_point"]
	w.Header().Set("Content-Type", "application/json")
	body, _ := ioutil.ReadAll(r.Body)
	_sendToSqs(queueName, endPoint, string(body))

}

func _sendToSqs(queueName string, endPoint string, body string) {
	log.Infof("Receiving payload with queueName: %s, endPoint: %s, body: %s", queueName, endPoint, body)
	var incoming = Payload{Endpoint: endPoint, Body: body}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(AWS_REGION)},
	}))
	svc := sqs.New(sess)
	queue, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: &queueName,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	toJson, _ := json.Marshal(incoming)
	_, err2 := svc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(toJson)),
		QueueUrl:    queue.QueueUrl,
	})
	if err2 != nil {
		fmt.Println(err2)
		return
	}
}

func main() {
	queueName := flag.String("queue_name", "", "the name of the sqs queue")
	endPoint := flag.String("end_point", "", "webhook endpoint on the target")
	body := flag.String("body", "", "payload body to send to webhook(if any)")
	port := flag.Int("port", 10000, "port to start the relay proxy")
	flag.Parse()
	if len(*queueName) > 0 && len(*endPoint) > 0 {
		_sendToSqs(*queueName, *endPoint, *body)
	} else {
		handleRequests(*port)
	}
}
