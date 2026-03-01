package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Queue interface {
	Enqueue(ctx context.Context, payload []byte) error
}

type SQSQueue struct {
	client *sqs.Client
	url    string
}

func NewSQSQueue(client *sqs.Client, url string) *SQSQueue {
	return &SQSQueue{
		client: client,
		url:    url,
	}
}

func (q *SQSQueue) Enqueue(ctx context.Context, payload []byte) error {
	_, err := q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &q.url,
		MessageBody: aws.String(string(payload)),
	})
	if err != nil {
		return fmt.Errorf("enqueue to sqs: %w", err)
	}
	return nil
}
