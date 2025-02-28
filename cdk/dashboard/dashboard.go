package dashboard

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/jsii-runtime-go"
)

const (
	statisticPeriod = 1  // minutes
	displayInterval = 30 // minutes
)

type Dashboard struct {
	Dashboard awscloudwatch.Dashboard
}

func NewDashboard(stack awscdk.Stack, name string) Dashboard {
	dashboard := awscloudwatch.NewDashboard(stack, aws.String(name), &awscloudwatch.DashboardProps{
		DashboardName:   aws.String(name + "-" + *stack.Region()),
		DefaultInterval: awscdk.Duration_Minutes(aws.Float64(displayInterval)),
	})

	return Dashboard{Dashboard: dashboard}
}

func (d *Dashboard) AddWidgetsRow(widgets ...awscloudwatch.IWidget) {
	row := awscloudwatch.NewRow(widgets...)
	d.Dashboard.AddWidgets(row)
}

func (d *Dashboard) CreateLambdaMetric(region string, metricName string, functionName *string, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String("AWS/Lambda"),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"FunctionName": functionName,
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(statisticPeriod)),
		Statistic: jsii.String(statistic),
	})
}

func (d *Dashboard) CreateGatewayMetric(region string, metricName string, apiName string, stage string, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String("AWS/ApiGateway"),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"ApiName": &apiName,
			"Stage":   &stage,
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(statisticPeriod)),
		Statistic: jsii.String(statistic),
	})
}

func (d *Dashboard) CreateTopicMetric(region string, metricName string, topicName *string, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String("AWS/SNS"),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"TopicName": topicName,
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(statisticPeriod)),
		Statistic: jsii.String(statistic),
	})
}

func (d *Dashboard) CreateQueueMetric(region string, metricName string, queueName *string, statistic string) awscloudwatch.IMetric {
	return awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
		Region:     jsii.String(region),
		Namespace:  jsii.String("AWS/SQS"),
		MetricName: jsii.String(metricName),
		DimensionsMap: &map[string]*string{
			"QueueName": queueName,
		},
		Period:    awscdk.Duration_Minutes(jsii.Number(statisticPeriod)),
		Statistic: jsii.String(statistic),
	})
}

func (d *Dashboard) CreateGraphWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.GraphWidget {
	return awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
		Region: jsii.String(region),
		Title:  jsii.String(title),
		Left:   &metrics,
		Height: jsii.Number(6),
		Width:  jsii.Number(6),
	})
}

func (d *Dashboard) CreateSingleValueWidget(region string, title string, metrics []awscloudwatch.IMetric) awscloudwatch.SingleValueWidget {
	return awscloudwatch.NewSingleValueWidget(&awscloudwatch.SingleValueWidgetProps{
		Region:               jsii.String(region),
		Title:                jsii.String(title),
		Metrics:              &metrics,
		SetPeriodToTimeRange: jsii.Bool(true),
		Height:               jsii.Number(6),
		Width:                jsii.Number(4),
	})
}
