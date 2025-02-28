package apigateway

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/jsii-runtime-go"
)

type APIGatewayProps struct {
	Stack          awscdk.Stack
	Name           string
	Description    string
	DefaultHandler awslambdago.GoFunction
}

type PublicAPIGatewayProps struct {
	APIGatewayProps
	EndpointTypes               *[]awsapigateway.EndpointType
	DefaultCorsPreflightOptions awsapigateway.CorsOptions
}

type PrivateAPIGatewayProps struct {
	APIGatewayProps
	VpcEndpoints *[]awsec2.IVpcEndpoint
}

func NewPublicAPIGateway(props PublicAPIGatewayProps) awsapigateway.LambdaRestApi {
	return awsapigateway.NewLambdaRestApi(props.Stack, jsii.String(props.Name), &awsapigateway.LambdaRestApiProps{
		DefaultCorsPreflightOptions: &props.DefaultCorsPreflightOptions,
		CloudWatchRole:              jsii.Bool(true), // BB - was false
		Description:                 jsii.String(props.Description),
		EndpointExportName:          nil,
		EndpointTypes:               props.EndpointTypes,
		// EndpointConfiguration: &awsapigateway.EndpointConfiguration{
		// 	Types: props.EndpointTypes,
		// },
		Handler: props.DefaultHandler,
		Proxy:   jsii.Bool(false),
		DeployOptions: &awsapigateway.StageOptions{
			LoggingLevel:   awsapigateway.MethodLoggingLevel_INFO,
			TracingEnabled: jsii.Bool(true),
		},
	})
}

func NewPublicAPIGatewayWithWAF(props PublicAPIGatewayProps, wafProps wafProps) awsapigateway.LambdaRestApi {
	api := NewPublicAPIGateway(props)
	waf := NewWaf(props.Stack, wafProps)
	AttachWafToApiGateway(props.Stack, api, waf)
	return api
}

func NewPrivateAPIGateway(props PrivateAPIGatewayProps) awsapigateway.LambdaRestApi {
	return awsapigateway.NewLambdaRestApi(props.Stack, jsii.String(props.Name), &awsapigateway.LambdaRestApiProps{
		CloudWatchRole:     jsii.Bool(false),
		Description:        jsii.String(props.Description),
		EndpointExportName: nil,
		EndpointTypes: &[]awsapigateway.EndpointType{
			awsapigateway.EndpointType_PRIVATE,
		},
		EndpointConfiguration: &awsapigateway.EndpointConfiguration{
			Types: &[]awsapigateway.EndpointType{
				awsapigateway.EndpointType_PRIVATE,
			},
			VpcEndpoints: props.VpcEndpoints,
		},
		Handler: props.DefaultHandler,
		Proxy:   jsii.Bool(false),
		DeployOptions: &awsapigateway.StageOptions{
			LoggingLevel:   awsapigateway.MethodLoggingLevel_INFO,
			TracingEnabled: jsii.Bool(true),
		},
	})
}

type ApiValues struct {
	ID  string
	Arn string
	Url string
}

func GetAllAPIs(stack awscdk.Stack, apiNames []string) (map[string]ApiValues, error) {
	mapOfApiNameToApiConstructs := map[string]ApiValues{}
	for _, apiName := range apiNames {
		apiUrl := awsssm.StringParameter_ValueFromLookup(stack, jsii.String("/api/"+apiName), jsii.String("default_value"))
		if *apiUrl == "default_value" {
			fmt.Println("WARNING: in ", *stack.Region(), "the parameter /api/", apiName, " not found in SSM, returning with defaults")
			defaultString := fmt.Sprint("arn:aws:execute-api:", *stack.Region(), ":", *stack.Account(), ":*/*/*/*")
			mapOfApiNameToApiConstructs[apiName] = ApiValues{
				ID:  "",
				Arn: defaultString,
				Url: "",
			}
			continue
		}
		// if it exists, then add values normally
		url, err := url.Parse(*apiUrl)
		if err != nil {
			return nil, err
		}
		apiID := strings.Split(url.Hostname(), ".")[0]
		// arn := fmt.Sprintf("arn:aws:apigateway:%s::/restapis/%s", *stack.Region(), apiID)
		api := awsapigateway.RestApi_FromRestApiId(stack, jsii.String("apigwImport"+apiName), jsii.String(apiID))
		mapOfApiNameToApiConstructs[apiName] = ApiValues{
			ID:  apiID,
			Arn: *api.ArnForExecuteApi(jsii.String("*"), jsii.String("/*"), jsii.String("*")),
			Url: *apiUrl,
		}
	}
	return mapOfApiNameToApiConstructs, nil
}
