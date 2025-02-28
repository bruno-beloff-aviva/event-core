package main

// https://github.com/aviva-verde/cdk-standards.git
// https://docs.aws.amazon.com/cdk/v2/guide/resources.html
// https://docs.aws.amazon.com/code-library/latest/ug/go_2_sqs_code_examples.html
// https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.aws_lambda_event_sources.SqsEventSource.html
// https://dev.to/chrisarmstrong/sqs-queues-as-an-eventbridge-rule-target-3d2g

import (
	"sqstest/cdk/dashboard"
	"sqstest/cdk/eventhandler"
	"sqstest/cdk/gatewayhandler"
	"sqstest/cdk/stackprops"
	"sqstest/service/testreception"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/constructs-go/constructs/v10"
)

const (
	project                 = "SQS1"
	version                 = "0.2.27"
	queueKeyAlias           = "QueueKeyLive"
	queue1Name              = "TestQueue1"
	queue2Name              = "TestQueue2"
	queue3Name              = "TestQueue3"
	queueMaxRetries         = 3
	eventBusName            = "TestEventBus"
	eventBusRuleName        = "GatewayTestMessage"
	tableName               = "TestMessageTableV2"
	eventBusId              = project + eventBusName
	tableId                 = project + tableName
	queueKeyId              = project + "QueueKey"
	pubHandlerId            = project + "PubHandler"
	pubEndpointId           = project + "PubEndpoint"
	continuousSubHandlerId  = project + "ContinuousHandler"
	suspendableSubHandlerId = project + "SudspendableHandler"
	stackId                 = project + "Stack"
	dashboardId             = project + "Dashboard"
)

func NewSQSStack(scope constructs.Construct, id string, stackProps *stackprops.CdkStackProps) (stack awscdk.Stack) {
	stack = stackProps.NewStack(scope, id)

	// Dashboard...
	dash := setupDashboard(stack)

	// EventBus...
	eventBus := setupEventBus(stack)

	// DBTable...
	table := setupMessageTable(stack, tableId, tableName)

	// QueueKey...
	queueKey := setupQueueKey(stack)

	// pub lambda...
	pubProps := gatewayhandler.GatewayCommonProps{
		Dashboard: dash,
	}

	c0 := setupPubHandler(stack, *stackProps, pubProps, eventBus)
	eventBus.GrantPutEventsTo(c0.Handler)

	// sub lambdas...
	subProps := eventhandler.EventHandlerCommonProps{
		QueueKey:        queueKey,
		QueueMaxRetries: queueMaxRetries,
		MessageTable:    table,
		Dashboard:       dash,
	}

	c1 := setupContinuousSubHandler(stack, subProps, queue1Name)
	c2 := setupSuspendableSubHandler(stack, subProps, queue2Name)
	c3 := setupEmptySubHandler(stack, subProps, queue3Name)

	// EventBus rule...
	rule, targetInput := setupEventBusRule(stack, eventBus, pubEndpointId)

	rule.AddTarget(awseventstargets.NewSqsQueue(c1.Queue, &targetInput))
	rule.AddTarget(awseventstargets.NewSqsQueue(c2.Queue, &targetInput))
	rule.AddTarget(awseventstargets.NewSqsQueue(c3.Queue, &targetInput))

	// Dashboard widgets...
	dash.AddWidgetsRow(c0.GatewayMetricsGraphWidget(), c0.LambdaMetricsGraphWidget(), c1.LambdaMetricsGraphWidget(), c2.LambdaMetricsGraphWidget())
	dash.AddWidgetsRow(c1.QueueMetricsGraphWidget(), c1.DLQMetricsGraphWidget(), c2.QueueMetricsGraphWidget(), c2.DLQMetricsGraphWidget())
	dash.AddWidgetsRow(c3.QueueMetricsGraphWidget(), c3.DLQMetricsGraphWidget())

	return stack
}

func setupDashboard(stack awscdk.Stack) dashboard.Dashboard {
	dash := dashboard.NewDashboard(stack, dashboardId)

	return dash
}

func setupMessageTable(stack awscdk.Stack, id string, name string) awsdynamodb.ITable {
	tableProps := awsdynamodb.TableProps{
		PartitionKey: testreception.DynamoPartitionKey(),
		SortKey:      testreception.DynamoSortKey(),
		TableName:    aws.String(name),
	}

	return awsdynamodb.NewTable(stack, aws.String(id), &tableProps)
}

func setupQueueKey(stack awscdk.Stack) awskms.IKey {
	keyProps := awskms.KeyProps{
		Alias:             aws.String(queueKeyAlias),
		EnableKeyRotation: aws.Bool(true),
	}

	return awskms.NewKey(stack, aws.String(queueKeyId), &keyProps)
}

func setupEventBus(stack awscdk.Stack) awsevents.IEventBus {
	busProps := awsevents.EventBusProps{
		EventBusName: aws.String(eventBusName),
		DeadLetterQueue: awssqs.NewQueue(stack, aws.String(eventBusId+"DLQ"), &awssqs.QueueProps{
			QueueName: aws.String(eventBusName + "DLQ"),
		}),
	}

	bus := awsevents.NewEventBus(stack, aws.String(eventBusId), &busProps)

	return bus
}

func setupEventBusRule(stack awscdk.Stack, eventBus awsevents.IEventBus, source string) (rule awsevents.Rule, targetInput awseventstargets.SqsQueueProps) {
	eventPattern := &awsevents.EventPattern{
		Source: &[]*string{
			aws.String(source),
		},
	}

	ruleProps := awsevents.RuleProps{
		EventBus:     eventBus,
		EventPattern: eventPattern,
		RuleName:     aws.String(eventBusRuleName + "Rule"),
	}

	rule = awsevents.NewRule(stack, aws.String(project+eventBusRuleName+"Rule"), &ruleProps)

	targetInput = awseventstargets.SqsQueueProps{
		Message:        awsevents.RuleTargetInput_FromEventPath(aws.String("$.detail")),
		MessageGroupId: aws.String(eventBusRuleName + "Group"),
	}

	return rule, targetInput
}

func setupPubHandler(stack awscdk.Stack, stackProps stackprops.CdkStackProps, commonProps gatewayhandler.GatewayCommonProps, eventBus awsevents.IEventBus) gatewayhandler.GatewayConstruct {
	environment := map[string]*string{
		"VERSION":        aws.String(version),
		"EVENT_SOURCE":   aws.String(pubEndpointId),
		"EVENT_BUS_NAME": eventBus.EventBusName(),
	}

	builder := gatewayhandler.GatewayBuilder{
		EndpointId:  pubEndpointId,
		HandlerId:   pubHandlerId,
		EventBus:    eventBus,
		Entry:       "lambda/pub/",
		Environment: environment,
	}

	return builder.Setup(stack, stackProps, commonProps)
}

func setupContinuousSubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, queueName string) eventhandler.EventHandlerConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
	}

	builder := eventhandler.EventHandlerBuilder{
		QueueName:   queueName,
		HandlerId:   continuousSubHandlerId,
		Entry:       "lambda/subcontinuous/",
		Environment: environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupSuspendableSubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, queueName string) eventhandler.EventHandlerConstruct {
	environment := map[string]*string{
		"VERSION":            aws.String(version),
		"MESSAGE_TABLE_NAME": aws.String(tableName),
		"SUSPENDED":          aws.String("false"),
	}

	builder := eventhandler.EventHandlerBuilder{
		QueueName:   queueName,
		HandlerId:   suspendableSubHandlerId,
		Entry:       "lambda/subsuspendable/",
		Environment: environment,
	}

	return builder.Setup(stack, commonProps)
}

func setupEmptySubHandler(stack awscdk.Stack, commonProps eventhandler.EventHandlerCommonProps, queueName string) eventhandler.EventHandlerConstruct {
	builder := eventhandler.EventHandlerBuilder{
		QueueName: queueName,
	}

	return builder.Setup(stack, commonProps)
}

func main() {
	cdkStackProps := stackprops.CdkStackProps{
		StackProps: awscdk.StackProps{},
		Version:    version,
	}

	app := awscdk.NewApp(nil)
	NewSQSStack(app, stackId, &cdkStackProps)

	app.Synth(nil)
}
