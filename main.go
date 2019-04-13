// Copyright (c) Keiran Smith 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"log"

	"github.com/architsmat38/golang-aws-sqs/poller"
	SqsService "github.com/architsmat38/golang-aws-sqs/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/openfaas-incubator/connector-sdk/types"
	"os"
	"strconv"
	"time"
)

var (
	accessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey   = os.Getenv("AWS_ACCESS_ACCESS_KEY_ID")
	region      = os.Getenv("AWS_REGION")
	queueName   = os.Getenv("AWS_SQS_QUEUE_NAME")
)

func InitializePollerSQS() {
	printResponse, _ := strconv.ParseBool(os.Getenv("PRINT_RESPONSE"))
	printResponseBody, _ := strconv.ParseBool(os.Getenv("PRINT_RESPONSE_BODY"))

	creds := types.GetCredentials()
	config := &types.ControllerConfig{
		RebuildInterval:   time.Millisecond * 1000,
		GatewayURL:        os.Getenv("GATEWAY_URL"),
		PrintResponse:     printResponse,
		PrintResponseBody: printResponseBody,
	}

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	go poller.Start(poller.HandlerFunc(func(msg *sqs.Message) error {
		var queueMessage = aws.StringValue(msg.Body)
		decoded, err := SqsService.Decode([]byte(queueMessage))
		if err != nil {
			return err
		}

		controller.Invoke("", &decoded)
		return nil
	}))
}

func main() {

	SqsService.Initialize(
		SqsService.New(queueName, region, accessKeyId, secretKey, ""),
		SqsService.SetWaitSeconds(20),
	)

	InitializePollerSQS()
}

type ResponseReceiver struct {
}

func (ResponseReceiver) Response(res types.InvokerResponse) {
	if res.Error != nil {
		log.Printf("tester got error: %s", res.Error.Error())
	} else {
		log.Printf("tester got result: [%d] %s => %s (%d) bytes", res.Status, res.Topic, res.Function, len(*res.Body))
	}
}
