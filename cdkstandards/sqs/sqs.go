// Package sqs provides a function that creates an SQS queue with a dead-letter queue
// and alarms for both the queue and the dead-letter queue.
package sqs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatchactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// DeadLetterQueueConfig defines the configuration for the dead-letter queue.
type SqsQueueWithDLQProps struct {
	Stack                  awscdk.Stack
	QueueName              string
	SQSKey                 awskms.IKey
	Fifo                   bool
	QMaxReceiveCount       int
	QAlarmPeriod           int
	QAlarmThreshold        int
	QAlarmEvaluationPeriod int
	QDeliveryDelay         int
	DLQAlertTopics         []awssns.ITopic
	// Timeout of processing a single message.
	//
	// After dequeuing, the processor has this much time to handle the message
	// and delete it from the queue before it becomes visible again for dequeueing
	// by another processor.
	//
	// Values must be from 0 to 43200 seconds (12 hours). If you don't specify
	// a value, AWS CloudFormation uses the default value of 30 seconds.
	// Default: 30
	//
	QVisibilityTimeout         *float64
	QContentBasedDeduplication *bool
	DLQAlarmPeriod             int
	DLQAlarmThreshold          int
	DLQAlarmEvaluationPeriod   int
	// Number of days the DLQ will retain a message
	DLQRetentionPeriodDays *float64
}

// NewSqsQueueWithDLQ creates a new SQS queue with a dead-letter queue and
// alarms for both the queue and the dead-letter queue.
func NewSqsQueueWithDLQ(props SqsQueueWithDLQProps) awssqs.Queue { // BB: was IQueue
	// Set default value for DLQRetentionPeriodDays if not provided
	var dlqRetentionPeriodDays awscdk.Duration
	if props.DLQRetentionPeriodDays != nil {
		dlqRetentionPeriodDays = awscdk.Duration_Days(props.DLQRetentionPeriodDays)
	} else {
		dlqRetentionPeriodDays = awscdk.Duration_Days(aws.Float64(14))
	}

	// DLQ
	dlq := awssqs.NewQueue(props.Stack, aws.String(props.QueueName+"DLQ"), &awssqs.QueueProps{
		Encryption:                awssqs.QueueEncryption_KMS,
		EncryptionMasterKey:       props.SQSKey,
		ContentBasedDeduplication: props.QContentBasedDeduplication,
		RetentionPeriod:           dlqRetentionPeriodDays,
		Fifo:                      aws.Bool(props.Fifo),
	})

	var visibilityTimeout awscdk.Duration
	if props.QVisibilityTimeout != nil {
		visibilityTimeout = awscdk.Duration_Seconds(props.QVisibilityTimeout)
	} else {
		visibilityTimeout = awscdk.Duration_Seconds(aws.Float64(30))
	}

	// Queue
	q := awssqs.NewQueue(props.Stack, aws.String(props.QueueName), &awssqs.QueueProps{
		Encryption:          awssqs.QueueEncryption_KMS,
		EncryptionMasterKey: props.SQSKey,
		DeliveryDelay:       awscdk.Duration_Seconds(aws.Float64(float64(props.QDeliveryDelay))),
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			MaxReceiveCount: aws.Float64(float64(props.QMaxReceiveCount)),
			Queue:           dlq,
		},
		Fifo:                      aws.Bool(props.Fifo),
		VisibilityTimeout:         visibilityTimeout,
		ContentBasedDeduplication: props.QContentBasedDeduplication,
	})

	// DLQ Alarm
	a := awscloudwatch.NewAlarm(props.Stack, aws.String(props.QueueName+"DLQAlarm"), &awscloudwatch.AlarmProps{
		AlarmDescription: aws.String("Alarm for " + props.QueueName + " DLQ"),
		Metric: dlq.MetricApproximateNumberOfMessagesVisible(&awscloudwatch.MetricOptions{
			Period:    awscdk.Duration_Minutes(aws.Float64(float64(props.DLQAlarmPeriod))),
			Statistic: aws.String("max"),
		}),
		Threshold:          aws.Float64(float64(props.DLQAlarmThreshold)),
		EvaluationPeriods:  aws.Float64(float64(props.DLQAlarmEvaluationPeriod)),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_IGNORE,
	})

	for _, topic := range props.DLQAlertTopics {
		a.AddAlarmAction(awscloudwatchactions.NewSnsAction(topic))
	}

	// alerts.CreateDLQRateAlarm(alerts.DLQRateAlarmProps{
	// 	Stack:  props.Stack,
	// 	ID:     props.QueueName + "DLQRateAlarm",
	// 	DLQ:    dlq,
	// 	Topics: props.DLQAlertTopics,
	// })

	// alerts.CreateDailyDLQAlarm(alerts.DailyDLQAlarmProps{
	// 	Stack:  props.Stack,
	// 	ID:     props.QueueName + "DailyDLQAlarm",
	// 	DLQ:    dlq,
	// 	Topics: props.DLQAlertTopics,
	// })

	// Queue Alarm
	awscloudwatch.NewAlarm(props.Stack, aws.String(props.QueueName+"Alarm"), &awscloudwatch.AlarmProps{
		AlarmDescription: aws.String("Alarm for " + props.QueueName + " queue"),
		Metric: q.MetricApproximateAgeOfOldestMessage(&awscloudwatch.MetricOptions{
			Period:    awscdk.Duration_Minutes(aws.Float64(float64(props.QAlarmPeriod))),
			Statistic: aws.String("max"),
		}),
		Threshold:          aws.Float64(float64(props.QAlarmThreshold)),
		EvaluationPeriods:  aws.Float64(float64(props.QAlarmEvaluationPeriod)),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_IGNORE,
	})

	return q
}
