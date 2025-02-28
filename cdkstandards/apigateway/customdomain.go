package apigateway

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CustomDomainProps struct {
	CustomDomainName string
	Api              awsapigateway.IRestApi
}

const (
	constructName string = "create custom domain"
	owner         string = "team name"
)

func CreateCustomDomain(scope constructs.Construct, id *string, props CustomDomainProps) {
	subDomainName, domainName := splitDomainAndSubdomain(props.CustomDomainName)

	hostedZone := awsroute53.PublicHostedZone_FromLookup(scope, jsii.String("hostedZone"), &awsroute53.HostedZoneProviderProps{
		DomainName: jsii.String(domainName),
	})

	// create certificate
	acmCertId := fmt.Sprintf("%s-%s", *id, "Certificate")
	acmCert := awscertificatemanager.NewCertificate(scope, jsii.String(acmCertId), &awscertificatemanager.CertificateProps{
		DomainName:      jsii.String(props.CustomDomainName),
		CertificateName: id,
		KeyAlgorithm:    awscertificatemanager.KeyAlgorithm_RSA_2048(),
		Validation:      awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})
	addTags(acmCert)

	// create custom domain
	customDomainName := awsapigateway.NewDomainName(scope, id, &awsapigateway.DomainNameProps{
		Certificate: acmCert,
		DomainName:  jsii.String(props.CustomDomainName),
	})
	addTags(customDomainName)

	// add api mapping to the custom domain
	pathMappingId := fmt.Sprintf("%s-%s", *id, "BasePathMapping")
	awsapigateway.NewBasePathMapping(scope, jsii.String(pathMappingId), &awsapigateway.BasePathMappingProps{
		DomainName: customDomainName,
		RestApi:    props.Api,
	})

	// add Route 53 records
	aRecordId := fmt.Sprintf("%s-%s", *id, "ARecord")
	awsroute53.NewARecord(scope, jsii.String(aRecordId), &awsroute53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(subDomainName),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewApiGatewayDomain(customDomainName)),
	})

	caaRecordId := fmt.Sprintf("%s-%s", *id, "CaaRecord")
	awsroute53.NewCaaRecord(scope, jsii.String(caaRecordId), &awsroute53.CaaRecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(subDomainName),
		Values: &[]*awsroute53.CaaRecordValue{
			{
				Flag:  jsii.Number(0),
				Tag:   awsroute53.CaaTag_ISSUE,
				Value: jsii.String("amazon.com"),
			},
			{
				Flag:  jsii.Number(0),
				Tag:   awsroute53.CaaTag_ISSUEWILD,
				Value: jsii.String("amazon.com"),
			},
		},
	})
}

func addTags(scope constructs.Construct) {
	tags := awscdk.Tags_Of(scope)

	tagsVals := map[string]string{
		"construct:name": constructName,
		"owner":          owner,
	}

	for tName, tValue := range tagsVals {
		tags.Add(jsii.String(tName), jsii.String(tValue), nil)
	}
}

// Given url passed in is "api.testing.verde.systems", this will return "api" as the subdomain and "testing.verde.systems" as the domain
// If you need to split your url in any other way, please tweak this function in your use case accordingly
func splitDomainAndSubdomain(d string) (string, string) {
	els := strings.SplitN(d, ".", 2)

	if len(els) < 2 {
		panic("invalid custom domain name!")
	}

	domain := els[1]
	subdomain := els[0]
	if subdomain == "" {
		panic("invalid subdomain in custom domain name!")
	}

	return subdomain, domain
}
