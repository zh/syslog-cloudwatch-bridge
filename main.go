package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	uuid "github.com/satori/go.uuid"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

var port = os.Getenv("LOG_PORT")
var logGroupName = os.Getenv("LOG_GROUP_NAME")
var streamName = os.Getenv("LOG_STREAM_NAME")
var obj, _err = uuid.NewV4()
var streamNameSuffix = obj.String()
var sequenceToken = ""

var (
	client *http.Client
	pool   *x509.CertPool
)

func init() {
	pool = x509.NewCertPool()
	pool.AppendCertsFromPEM(pemCerts)
}

func main() {
	if logGroupName == "" {
		log.Fatal("LOG_GROUP_NAME must be specified")
	}

	if streamName == "" {
		log.Fatal("LOG_STREAM_NAME must be specified")
	}
	streamName = streamName + "-" + streamNameSuffix

	if port == "" {
		port = "5014"
	}

	address := fmt.Sprintf("127.0.0.1:%v", port)
	log.Println("Starting syslog server on", address)
	log.Println("Logging to group:", logGroupName)
	log.Println("Logging to stream:", streamName)
	initCloudWatchStream()

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	server.ListenUDP(address)
	server.ListenTCP(address)

	server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			sendToCloudWatch(logParts)
		}
	}(channel)

	server.Wait()
}

func sendToCloudWatch(logPart format.LogParts) {
	if logPart == nil || logPart["content"] == nil || logPart["content"] == "" {
		log.Println("invalid log format")
		spew.Dump(logPart)
		return
	}

	// service is defined at run time to avoid session expiry in long running processes
	var svc = cloudwatchlogs.New(session.New())
	// set the AWS SDK to use our bundled certs for the minimal container (certs from CoreOS linux)
	svc.Config.HTTPClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}

	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{
				Message:   aws.String(logPart["content"].(string)),
				Timestamp: aws.Int64(makeMilliTimestamp(logPart["timestamp"].(time.Time))),
			},
		},
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(streamName),
	}

	// first request has no SequenceToken - in all subsequent request we set it
	if sequenceToken != "" {
		params.SequenceToken = aws.String(sequenceToken)
	}

	resp, err := svc.PutLogEvents(params)
	if err != nil {
		log.Println(err)
	}

	sequenceToken = *resp.NextSequenceToken
}

func initCloudWatchStream() {
	// service is defined at run time to avoid session expiry in long running processes
	var svc = cloudwatchlogs.New(session.New())
	// set the AWS SDK to use our bundled certs for the minimal container (certs from CoreOS linux)
	svc.Config.HTTPClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}}

	_, err := svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(streamName),
	})

	if err != nil {
		log.Println("Created CloudWatch Logs stream (err):", err)
		// 	 log.Fatal(err)
	}

	log.Println("Created CloudWatch Logs stream:", streamName)
}

func makeMilliTimestamp(input time.Time) int64 {
	return input.UnixNano() / int64(time.Millisecond)
}
