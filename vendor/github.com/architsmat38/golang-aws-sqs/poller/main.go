package poller

import (
    "log"
    "sync"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"
    SqsService "github.com/architsmat38/golang-aws-sqs/sqs"
)

// HandlerFunc is used to define the Handler that is run on for each message
type HandlerFunc func(msg *sqs.Message) error

func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
    return f(msg)
}

// Handler interface
type Handler interface {
    HandleMessage(msg *sqs.Message) error
}


/**
 * Starts polling the queue and fetches messages once they are queued
 */
func Start(h Handler) {
    for {
        log.Println("Polling ..!!")
        resp, err := SqsService.Receive()

        if err != nil {
            log.Println(err)
            continue
        }

        if len(resp.Messages) > 0 {
            process(h, resp.Messages)
        }
    }
}


/**
 * Process each message and run specific go routine for each message processing
 */
func process(h Handler, messages []*sqs.Message) {
    var wg sync.WaitGroup
    numMessages := len(messages)
    log.Printf("Received %d messages\n", numMessages)

    wg.Add(numMessages)
    for i := range messages {
        go func(m *sqs.Message) {
            defer wg.Done()
            if err := handleMessage(m, h); err != nil {
                log.Println(err.Error())
            }
        }(messages[i])
    }

    wg.Wait()
}


/**
 * Handle each message and delete once it is processed
 */
func handleMessage(m *sqs.Message, h Handler) error {
    var err error
    err = h.HandleMessage(m)
    if err != nil {
        return err
    }

    _, err = SqsService.Delete(m.ReceiptHandle)
    if err != nil {
        return err
    }

    log.Printf("Deleted message from queue: %s\n", aws.StringValue(m.ReceiptHandle))
    return nil
}
