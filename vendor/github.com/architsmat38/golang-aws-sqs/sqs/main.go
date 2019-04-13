package sqs

import (
    "log"
    "runtime"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws/credentials"
)


type SQSStruct struct {
    SQS *sqs.SQS
    URL string
}

var (
    SQSClient           SQSStruct
    MaxCountMessage     int64 = 10
    VisibilitySeconds   int64 = 10
    WaitSeconds         int64 = 10
)


/**
 * Initialize SQS and corresponding configuration
 * Uses "Functional Options Pattern"
 * Reference: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
 */
func Initialize(options ...func() error) error {
    for _, option := range options {
        err := option()

        if err != nil {
            return err
        }
    }

    runtime.GOMAXPROCS(runtime.NumCPU())
    return nil
}


/**
 * Get AWS config to initialize SQS queue
 */
func getAWSConfig(region, id, secret, token string) *aws.Config {
    return &aws.Config{
        Region: aws.String(region),
        Credentials: credentials.NewStaticCredentials(
            id,
            secret,
            token,
        ),
    }
}


/**
 * Create SQS client using AWS config
 */
func NewUsingConfig(queue string, config *aws.Config) func() error {
    return func() error {
        sess, err := session.NewSession()
        if err != nil {
            log.Println("Failed to create session")
            return err
        }
        SQSClient.SQS = sqs.New(sess, config)

        params := &sqs.GetQueueUrlInput{
            QueueName: aws.String(queue),
        }

        resp, err := SQSClient.SQS.GetQueueUrl(params)
        if err != nil {
            log.Println("Failed to get queue url")
            return err
        }

        SQSClient.URL = aws.StringValue(resp.QueueUrl)
        return nil
    }
}


/**
 * Create SQS client using required AWS params (including creds)
 */
func New(queue, region, id, secret, token string) func() error {
    var config *aws.Config = getAWSConfig(region, id, secret, token)
    return NewUsingConfig(queue, config)
}


/**
 * Sets the max count of message that can be returned by Amazon SQS
 * The value can range between 1 to 10
 */
func SetMaxCountMessage(maxCount int64) func() error {
    return func() error {
        MaxCountMessage = maxCount
        return nil
    }
}


/**
 * Sets the duration for which the received messages are hidden from subsequent retrieve requests
 */
func SetVisibilitySeconds(timeout int64) func() error {
    return func() error {
        VisibilitySeconds = timeout
        return nil
    }
}


/**
 * Sets the duration for which the call waits for a message to appear in queue
 */
func SetWaitSeconds(waitTime int64) func() error {
    return func() error {
        WaitSeconds = waitTime
        return nil
    }
}

