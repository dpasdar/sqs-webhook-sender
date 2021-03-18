package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
)

type Sample struct {
	Endpoint string   `json:"end_point"`
	Payload  string   `json:"payload"`
	Headers  []string `json:"headers"`
}

var AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")
var AWS_ACCESS_KEY_ID = os.Getenv("AWS_ACCESS_KEY_ID")
var AWS_REGION = os.Getenv("AWS_REGION")

func handleRequests() {
	// creates a new instance of a mux router
	myRouter := mux.NewRouter().StrictSlash(true)
	// replace http.HandleFunc with myRouter.HandleFunc
	myRouter.HandleFunc("/{queue_name}/{end_point}", sendToSqs)
	// finally, instead of passing in nil, we want
	// to pass in our newly created router as the second
	// argument
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func sendToSqs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queueName := vars["queue_name"]
	endPoint := vars["end_point"]
	var sample1 = Sample{Endpoint: endPoint}
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(sample1)
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
	toJson, _ := json.Marshal(sample1)
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
	fmt.Println("Publisher started...")
	handleRequests()
}
