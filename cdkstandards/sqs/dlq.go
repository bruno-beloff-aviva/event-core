package sqs

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatchactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
)

// DeadLetterQueueConfig defines the configuration for the dead-letter queue.
type DeadLetterQueueConfig struct {
	SQSKey awskms.IKey
	Topics []awssns.ITopic
	// Exclude the Daily and Rate alarms when initialising the DLQ
	ExcludeAlarms     bool
	ExcludeRateAlarm  bool
	ExcludeDailyAlarm bool
}

// NewDeadletterQueue creates a new dead-letter queue.
func NewDeadletterQueue(stack awscdk.Stack, lambdaName string, props DeadLetterQueueConfig) awssqs.Queue {
	dlq := awssqs.NewQueue(stack, aws.String(lambdaName+"DLQ"), &awssqs.QueueProps{
		RetentionPeriod:     awscdk.Duration_Days(aws.Float64(14)),
		Encryption:          awssqs.QueueEncryption_KMS,
		EncryptionMasterKey: props.SQSKey,
	})

	dlqMetric := dlq.MetricApproximateNumberOfMessagesVisible(&awscloudwatch.MetricOptions{
		Period:    awscdk.Duration_Minutes(aws.Float64(1)),
		Statistic: aws.String("max"),
	})

	dlqMetricAlarm := dlqMetric.CreateAlarm(stack, aws.String(lambdaName+"DLQAlarm"), &awscloudwatch.CreateAlarmOptions{
		DatapointsToAlarm:  aws.Float64(1),
		EvaluationPeriods:  aws.Float64(1),
		Threshold:          aws.Float64(1),
		ActionsEnabled:     aws.Bool(true),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_IGNORE,
	})

	// if !props.ExcludeAlarms {
	// 	alerts.CreateDailyDLQAlarm(alerts.DailyDLQAlarmProps{
	// 		Stack:  stack,
	// 		ID:     lambdaName + "DailyDLQAlarm",
	// 		DLQ:    dlq,
	// 		Topics: props.Topics,
	// 	})

	// 	alerts.CreateDLQRateAlarm(alerts.DLQRateAlarmProps{
	// 		Stack:  stack,
	// 		ID:     lambdaName + "DLQRateAlarm",
	// 		DLQ:    dlq,
	// 		Topics: props.Topics,
	// 	})
	// }

	for _, topic := range props.Topics {
		dlqMetricAlarm.AddAlarmAction(awscloudwatchactions.NewSnsAction(topic))
	}

	return dlq
}

// DeadLetterQueueConfig defines the configuration for the dead-letter queue.
type DeadLetterQueueV2Config struct {
	SQSKey awskms.IKey
	Topics []awssns.ITopic
	// Exclude the Daily and Rate alarms when initialising the DLQ
	ExcludeRateAlarm  bool
	ExcludeDailyAlarm bool
}

// NewDeadletterQueue creates a new dead-letter queue.
func NewDeadLetterQueueV2(stack awscdk.Stack, lambdaName string, props DeadLetterQueueV2Config) awssqs.Queue {
	err := validateDeadLetterQueueConfig(props)
	if err != nil {
		fmt.Printf("error creating DLQ: %v\n", err)
		os.Exit(1)
	}

	dlq := awssqs.NewQueue(stack, aws.String(lambdaName+"DLQ"), &awssqs.QueueProps{
		RetentionPeriod:     awscdk.Duration_Days(aws.Float64(14)),
		Encryption:          awssqs.QueueEncryption_KMS,
		EncryptionMasterKey: props.SQSKey,
	})

	dlqMetric := dlq.MetricApproximateNumberOfMessagesVisible(&awscloudwatch.MetricOptions{
		Period:    awscdk.Duration_Minutes(aws.Float64(1)),
		Statistic: aws.String("max"),
	})

	dlqMetricAlarm := dlqMetric.CreateAlarm(stack, aws.String(lambdaName+"DLQAlarm"), &awscloudwatch.CreateAlarmOptions{
		DatapointsToAlarm:  aws.Float64(1),
		EvaluationPeriods:  aws.Float64(1),
		Threshold:          aws.Float64(1),
		ActionsEnabled:     aws.Bool(true),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		TreatMissingData:   awscloudwatch.TreatMissingData_IGNORE,
	})

	// if !props.ExcludeDailyAlarm {
	// 	alerts.CreateDailyDLQAlarm(alerts.DailyDLQAlarmProps{
	// 		Stack:  stack,
	// 		ID:     lambdaName + "DailyDLQAlarm",
	// 		DLQ:    dlq,
	// 		Topics: props.Topics,
	// 	})
	// }

	// if !props.ExcludeRateAlarm {
	// 	alerts.CreateDLQRateAlarm(alerts.DLQRateAlarmProps{
	// 		Stack:  stack,
	// 		ID:     lambdaName + "DLQRateAlarm",
	// 		DLQ:    dlq,
	// 		Topics: props.Topics,
	// 	})
	// }

	for _, topic := range props.Topics {
		dlqMetricAlarm.AddAlarmAction(awscloudwatchactions.NewSnsAction(topic))
	}

	return dlq
}

func validateDeadLetterQueueConfig(cfg DeadLetterQueueV2Config) error {
	if cfg.SQSKey == nil {
		return fmt.Errorf("missing DLQ SQSKey")
	}
	if len(cfg.Topics) == 0 {
		return fmt.Errorf("missing DLQ SNS Topics")
	}
	return nil
}
