package eventhandler

import (
	"event-core/cdk/dashboard"
	"event-core/cdkstandards/sqs"
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go/aws"
)

type EventHandlerCommonProps struct {
	QueueKey        awskms.IKey
	QueueMaxRetries int
	MessageTable    awsdynamodb.ITable
	Dashboard       dashboard.Dashboard
}

type EventHandlerBuilder struct {
	QueueName   string
	HandlerId   string
	Entry       string
	Environment map[string]*string
}

type EventHandlerConstruct struct {
	Builder   EventHandlerBuilder
	Queue     awssqs.Queue
	Handler   awslambdago.GoFunction
	Dashboard dashboard.Dashboard
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (b EventHandlerBuilder) Setup(stack awscdk.Stack, commonProps EventHandlerCommonProps) EventHandlerConstruct {
	var c EventHandlerConstruct

	c.Builder = b
	c.Dashboard = commonProps.Dashboard
	c.Queue = b.setupQueue(stack, commonProps)

	if b.HandlerId == "" {
		return c
	}

	c.Handler = b.setupSubHandler(stack, c.Queue)
	c.Queue.GrantConsumeMessages(c.Handler)
	commonProps.MessageTable.GrantReadWriteData(c.Handler)

	return c
}

func (b EventHandlerBuilder) setupQueue(stack awscdk.Stack, commonProps EventHandlerCommonProps) awssqs.Queue {
	queueProps := sqs.SqsQueueWithDLQProps{
		Stack:                      stack,
		QueueName:                  b.QueueName,
		Fifo:                       true,
		QContentBasedDeduplication: aws.Bool(true),
		SQSKey:                     commonProps.QueueKey,
		QMaxReceiveCount:           commonProps.QueueMaxRetries,
		QAlarmPeriod:               1,
		QAlarmThreshold:            1,
		QAlarmEvaluationPeriod:     1,
		DLQAlarmPeriod:             1,
		DLQAlarmThreshold:          1,
		DLQAlarmEvaluationPeriod:   1,
	}

	return sqs.NewSqsQueueWithDLQ(queueProps)
}

func (b EventHandlerBuilder) setupSubHandler(stack awscdk.Stack, queue awssqs.IQueue) awslambdago.GoFunction {
	handlerProps := awslambdago.GoFunctionProps{
		Description:   aws.String("Handler with queue listening to EventBridge events"),
		Runtime:       awslambda.Runtime_PROVIDED_AL2(),
		Architecture:  awslambda.Architecture_ARM_64(),
		Entry:         aws.String(b.Entry),
		Timeout:       awscdk.Duration_Seconds(aws.Float64(28)),
		LoggingFormat: awslambda.LoggingFormat_JSON,
		LogRetention:  awslogs.RetentionDays_FIVE_DAYS,
		Tracing:       awslambda.Tracing_ACTIVE,
		Environment:   &b.Environment,
	}

	handler := awslambdago.NewGoFunction(stack, aws.String(b.HandlerId), &handlerProps)

	// TODO: use alias AFTER the project has been split, and deployments with / without alias can be tested

	eventSourceProps := awslambdaeventsources.SqsEventSourceProps{}
	handler.AddEventSource(awslambdaeventsources.NewSqsEventSource(queue, &eventSourceProps))

	return handler
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c EventHandlerConstruct) LambdaMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Handler.Stack().Region()

	invocationsMetric := c.Dashboard.CreateLambdaMetric(*region, "Invocations", c.Handler.FunctionName(), "Sum")
	errorsMetric := c.Dashboard.CreateLambdaMetric(*region, "Errors", c.Handler.FunctionName(), "Sum")
	metrics := []awscloudwatch.IMetric{invocationsMetric, errorsMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Invocations & Errors", c.Builder.HandlerId), metrics)
}

func (c EventHandlerConstruct) QueueMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Queue.Stack().Region()
	queueName := c.Queue.QueueName()

	sentMetric := c.Dashboard.CreateQueueMetric(*region, "NumberOfMessagesSent", queueName, "Sum")
	visibleMetric := c.Dashboard.CreateQueueMetric(*region, "ApproximateNumberOfMessagesVisible", queueName, "Sum")
	metrics := []awscloudwatch.IMetric{sentMetric, visibleMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%s - Sent & Visible", c.Builder.QueueName), metrics)
}

func (c EventHandlerConstruct) DLQMetricsGraphWidget() awscloudwatch.GraphWidget {
	region := c.Queue.Stack().Region()
	queueName := c.Queue.DeadLetterQueue().Queue.QueueName()

	visibleMetric := c.Dashboard.CreateQueueMetric(*region, "ApproximateNumberOfMessagesVisible", queueName, "Sum")
	invisibleMetric := c.Dashboard.CreateQueueMetric(*region, "ApproximateNumberOfMessagesNotVisible", queueName, "Sum")
	metrics := []awscloudwatch.IMetric{visibleMetric, invisibleMetric}

	return c.Dashboard.CreateGraphWidget(*region, fmt.Sprintf("%sDLQ - Visible & Invisible", c.Builder.QueueName), metrics)
}
