package report

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
)

type Worker struct {
	appConfig   *config.Config
	builder     *ReportBuilder
	logger      *slog.Logger
	sqsClient   *sqs.Client
	channel     chan types.Message
	concurrency int32
}

func NewWorker(
	appConfig *config.Config,
	builder *ReportBuilder,
	logger *slog.Logger,
	sqsClient *sqs.Client,
	maxCucurrency int32,
) *Worker {
	return &Worker{
		appConfig:   appConfig,
		builder:     builder,
		logger:      logger,
		sqsClient:   sqsClient,
		channel:     make(chan types.Message, maxCucurrency),
		concurrency: maxCucurrency,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	queueURLOutput, err := w.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(w.appConfig.SQSQueue),
	})
	if err != nil {
		return fmt.Errorf("failed to get url for queue %s: %w", w.appConfig.SQSQueue, err)
	}
	w.logger.InfoContext(ctx, "starting worker", slog.String(
		"queue", w.appConfig.SQSQueue,
	), slog.String("queue_url", *queueURLOutput.QueueUrl))

	for i := 0; i < int(w.concurrency); i++ {
		go func(id int) {
			w.logger.Info(fmt.Sprintf("starting goroutine #%d", id))
			for {
				select {
				case <-ctx.Done():
					w.logger.Error("worker stopped", slog.Int("goroutine_id", id),
						slog.Any("error", ctx.Err()),
					)
					return
				case message := <-w.channel:
					if err := w.processMessage(ctx, message); err != nil {
						w.logger.Error("failed to process message", slog.Any("error", err),
							slog.Int("goroutine_id", id),
						)
						continue
					}
					if _, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
						QueueUrl:      queueURLOutput.QueueUrl,
						ReceiptHandle: message.ReceiptHandle,
					}); err != nil {
						w.logger.Error("failed to delete message", slog.Any("error", err),
							slog.Int("goroutine_id", id))
					}
				}
			}
		}(i)
	}

	for {
		output, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            queueURLOutput.QueueUrl,
			MaxNumberOfMessages: w.concurrency + 1,
		})
		if err != nil {
			w.logger.Error("failed to receive message", slog.Any("error", err))
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}

		if len(output.Messages) == 0 {
			continue
		}
		for _, message := range output.Messages {
			w.channel <- message
		}

	}
}

func (w *Worker) processMessage(ctx context.Context, message types.Message) error {
	w.logger.InfoContext(ctx, "processing message", slog.String("message_id", *message.MessageId))
	if message.Body == nil || *message.Body == "" {
		w.logger.ErrorContext(ctx, "message body is empty", slog.String("message_id", *message.MessageId))
		return nil
	}

	var msg SQSMessage
	if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
		w.logger.Warn("message body is invalid", slog.String("message_id", *message.MessageId), slog.String("body", *message.Body))
		return nil
	}

	builderCtx, builderCancel := context.WithTimeout(ctx, time.Second*10)
	defer builderCancel()
	_, err := w.builder.Build(builderCtx, msg.UserID, msg.ReportID)
	if err != nil {
		return fmt.Errorf("failed to build report: %w", err)
	}

	return nil
}
