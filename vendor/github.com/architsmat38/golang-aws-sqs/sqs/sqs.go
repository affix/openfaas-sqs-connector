package sqs

import (
    "sync"
    "runtime"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"
)

// SQS message batch output
type SQSMessageBatchOutput struct {
    Output *sqs.SendMessageBatchOutput
    Error  error
}

// Max batch size of messages that can be sent to SQS
// This is excluding the metadata that is sent along with the message
var MAX_SQS_BATCH_SIZE_KB = 150
var MAX_BATCH_MESSAGE_COUNT = 10


/**
 * Send message to SQS queue
 */
 func Send(message string) (*sqs.SendMessageOutput, error) {
    var encodedMessage []byte = Encode([]byte(message))

    return SQSClient.SQS.SendMessage(&sqs.SendMessageInput{
        QueueUrl:    &SQSClient.URL,
        MessageBody: aws.String(string(encodedMessage)),
    })
 }


/**
 * Send messages in batch to SQS
 * This reduces the number of calls to SQS, also it reduces the cost as well :)
 */
func SendBatch(messages []string) (*sqs.SendMessageBatchOutput, error) {
    var entries []*sqs.SendMessageBatchRequestEntry = make([]*sqs.SendMessageBatchRequestEntry, len(messages))

    for index, message := range messages {
        var encodedMessage []byte = Encode([]byte(message))

        entries[index] = &sqs.SendMessageBatchRequestEntry{
            Id:          aws.String(string(97 + index)),
            MessageBody: aws.String(string(encodedMessage)),
        }
    }

    return SQSClient.SQS.SendMessageBatch(&sqs.SendMessageBatchInput{
        QueueUrl:   &SQSClient.URL,
        Entries:    entries,
    })
}


/**
 * Receive messages from SQS queue
 */
func Receive() (*sqs.ReceiveMessageOutput, error) {
    return SQSClient.SQS.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:               &SQSClient.URL,
        MaxNumberOfMessages:    aws.Int64(MaxCountMessage),
        VisibilityTimeout:      aws.Int64(VisibilitySeconds),
        WaitTimeSeconds:        aws.Int64(WaitSeconds),
        MessageAttributeNames:  []*string{aws.String("All")},
    })
}

/**
 * Delete a message from SQS queue
 */
func Delete(ReceiptHandle *string) (*sqs.DeleteMessageOutput, error) {
    return SQSClient.SQS.DeleteMessage(&sqs.DeleteMessageInput{
        QueueUrl:      &SQSClient.URL,
        ReceiptHandle: ReceiptHandle,
    })
}


/**
 * Purge all messages from SQS queue
 */
func Purge() (*sqs.PurgeQueueOutput, error) {
    return SQSClient.SQS.PurgeQueue(&sqs.PurgeQueueInput{
        QueueUrl: &SQSClient.URL,
    })
}


/**
 * Group all messages into chunks of 256 KB each and then send these chunks to SQS
 */
func ProcessAndSendBatch(messages []string) []*SQSMessageBatchOutput {
    var (
        wg                  sync.WaitGroup
        chunkMessages       []string
        chunkSize           int = 0
        countMessage        int = 0
        output              []*SQSMessageBatchOutput
    )

    // Send in batch
    result := make(chan *SQSMessageBatchOutput)
    send := func(messages []string) {
        runtime.Gosched()
        var output = &SQSMessageBatchOutput{}
        output.Output, output.Error = SendBatch(messages)
        result <- output
    }

    // Queue messages in batches
    for _, message := range messages {
        messageBytes := []byte(message)

        var messageLength int = len(messageBytes)
        if (chunkSize + messageLength) / 1000 < MAX_SQS_BATCH_SIZE_KB && countMessage < MAX_BATCH_MESSAGE_COUNT {
            // Creates message batch
            chunkSize = chunkSize + messageLength
            countMessage++

        } else {
            // Send batch
            wg.Add(1)
            go send(chunkMessages)
            chunkSize = messageLength
            countMessage = 1
            chunkMessages = nil
        }

        chunkMessages = append(chunkMessages, string(messageBytes))
    }

    if len(chunkMessages) > 0 {
        wg.Add(1)
        go send(chunkMessages)
    }

    // Gathering output
    output = make([]*SQSMessageBatchOutput, 0)
    go func() {
        for {
            select {
            case value, ok := <-result:
                if ok {
                    output = append(output, value)
                    wg.Done()
                }
            }
        }
    }()
    wg.Wait()

    return output
}
