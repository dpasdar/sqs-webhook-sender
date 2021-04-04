package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gorilla/mux"
)

type Payload struct {
	Endpoint string              `json:"end_point"`
	Body     string              `json:"body"`
	Headers  map[string][]string `json:"headers"`
}

func handleRequests(port int) {
	log.Infof("Sender started on port %d...", port)
	// setup signal catching
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Infof("RECEIVED SIGNAL: %s", s)
		os.Exit(0)
	}()
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
	_sendToSqs(queueName, endPoint, string(body), r.Header)

}

func _sendToSqs(queueName string, endPoint string, body string, headers map[string][]string) {
	log.Infof("Receiving payload with queueName: %s, endPoint: %s, body: %s", queueName, endPoint, body)
	var incoming = Payload{Endpoint: endPoint, Body: body, Headers: headers}
	sess := session.Must(session.NewSession())
	svc := sqs.New(sess)
	queue, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: &queueName,
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	toJson, _ := json.Marshal(incoming)
	_, err2 := svc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(toJson)),
		QueueUrl:    queue.QueueUrl,
	})
	if err2 != nil {
		log.Warn(err2)
		return
	}
}
func parseHeaders(headers string) map[string][]string {
	result := make(map[string][]string)
	entries := strings.Split(headers, ";")
	for _, e := range entries {
		parts := strings.Split(strings.TrimSpace(e), ":")
		if len(parts) == 2 {
			log.Debugf("Sending Header: {%s: %s}", parts[0], parts[1])
			r := make([]string, 1)
			r[0] = parts[1]
			result[parts[0]] = r
		}
	}
	return result
}
func main() {
	queueName := flag.String("queue_name", "", "the name of the sqs queue")
	debug := flag.Bool("debug", false, "Enable debug level messages")
	endPoint := flag.String("end_point", "", "webhook endpoint on the target")
	body := flag.String("body", "", "payload body to send to webhook(if any)")
	headers := flag.String("headers", "", "HTTP headers to pass to webhook, e.g. key1:value1;key2:value2;...")
	port := flag.Int("port", 10000, "port to start the relay proxy")

	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if len(*queueName) > 0 && len(*endPoint) > 0 {
		_sendToSqs(*queueName, *endPoint, *body, parseHeaders(*headers))
	} else {
		handleRequests(*port)
	}
}
