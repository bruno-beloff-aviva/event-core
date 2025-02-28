package apigateway

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awswafv2"
	"github.com/aws/jsii-runtime-go"
)

type Scope string

const (
	ScopeRegional   Scope = "REGIONAL"
	ScopeCloudfront Scope = "CLOUDFRONT"
)

type visibilityProps struct {
	EnableCloudWatchMetrics bool
	EnableSampleRequests    bool
	MetricName              string
}

type ManagedRuleSet struct {
	visibilityProps
	RuleSet    string
	Exclusions []string
}

func NewManagedRuleSet(ruleSet string, enableCloudWatchMetrics, enableSampleRequests bool, exclusions []string) ManagedRuleSet {
	return ManagedRuleSet{
		RuleSet: ruleSet,
		visibilityProps: visibilityProps{
			EnableCloudWatchMetrics: enableCloudWatchMetrics,
			EnableSampleRequests:    enableSampleRequests,
		},
		Exclusions: exclusions,
	}
}

type wafProps struct {
	visibilityProps
	ManagedRuleSets []ManagedRuleSet
	CustomRules     []awswafv2.CfnWebACL_RuleProperty
	WafId           string
	Scope           Scope
	BlockByDefault  bool
}

func NewWafProps(wafId string, managedRuleSets []ManagedRuleSet, customRules []awswafv2.CfnWebACL_RuleProperty, scope Scope, blockRequestsByDefault bool, enableCloudWatchMetrics, enableSampleRequests bool) wafProps {
	return wafProps{
		ManagedRuleSets: managedRuleSets,
		CustomRules:     customRules,
		WafId:           wafId,
		Scope:           scope,
		BlockByDefault:  blockRequestsByDefault,
		visibilityProps: visibilityProps{
			EnableCloudWatchMetrics: enableCloudWatchMetrics,
			EnableSampleRequests:    enableSampleRequests,
		},
	}
}

// Returns attribute ARN of a WAF with given props.
func NewWaf(stack awscdk.Stack, props wafProps) *string {
	allRules := make([]awswafv2.CfnWebACL_RuleProperty, 0)
	for idx, ruleSet := range props.ManagedRuleSets {
		allRules = append(allRules, createRule(ruleProps{
			Index: idx,
			Name:  ruleSet.RuleSet,
			visibilityProps: visibilityProps{
				EnableCloudWatchMetrics: ruleSet.EnableCloudWatchMetrics,
				EnableSampleRequests:    ruleSet.EnableSampleRequests,
			},
			Exclusions: ruleSet.Exclusions,
		}))
	}

	for _, cr := range props.CustomRules {
		cr.Priority = jsii.Number(len(allRules))
		allRules = append(allRules, cr)

	}

	nnv := struct{}{}

	var defaultAction awswafv2.CfnWebACL_DefaultActionProperty
	if props.BlockByDefault {
		defaultAction = awswafv2.CfnWebACL_DefaultActionProperty{
			Block: &nnv,
		}
	} else {
		defaultAction = awswafv2.CfnWebACL_DefaultActionProperty{
			Allow: &nnv,
		}
	}

	waf := awswafv2.NewCfnWebACL(stack, jsii.String("webApiAclArn"), &awswafv2.CfnWebACLProps{
		Name:          jsii.String(props.WafId),
		Scope:         jsii.String(string(props.Scope)),
		DefaultAction: defaultAction,
		VisibilityConfig: awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(props.EnableSampleRequests),
			CloudWatchMetricsEnabled: jsii.Bool(props.EnableCloudWatchMetrics),
			MetricName:               jsii.String(props.WafId),
		},
		Rules: allRules,
	})

	return waf.AttrArn()
}

func AttachWafToApiGateway(stack awscdk.Stack, api awsapigateway.LambdaRestApi, wafARN *string) {
	association := awswafv2.NewCfnWebACLAssociation(stack, jsii.String("webApiAclAssociation"), &awswafv2.CfnWebACLAssociationProps{
		ResourceArn: api.DeploymentStage().StageArn(),
		WebAclArn:   wafARN,
	})

	association.Node().AddDependency(api)
}

type ruleProps struct {
	visibilityProps
	Index      int
	Name       string
	Exclusions []string
}

func createRule(props ruleProps) awswafv2.CfnWebACL_RuleProperty {
	nonNullValue := struct{}{}
	var exclusionsList []awswafv2.CfnWebACL_ExcludedRuleProperty
	for _, v := range props.Exclusions {
		exclusionsList = append(exclusionsList, awswafv2.CfnWebACL_ExcludedRuleProperty{Name: jsii.String(v)})
	}
	return awswafv2.CfnWebACL_RuleProperty{
		Name: jsii.String(props.Name),
		Statement: awswafv2.CfnWebACL_StatementProperty{
			ManagedRuleGroupStatement: awswafv2.CfnWebACL_ManagedRuleGroupStatementProperty{
				VendorName:    jsii.String("AWS"),
				Name:          jsii.String(props.Name),
				ExcludedRules: exclusionsList,
			},
		},
		OverrideAction: awswafv2.CfnWebACL_OverrideActionProperty{
			None: &nonNullValue,
		},
		Priority: jsii.Number(props.Index),
		VisibilityConfig: awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(props.EnableSampleRequests),
			CloudWatchMetricsEnabled: jsii.Bool(props.EnableCloudWatchMetrics),
			MetricName:               jsii.String(props.Name),
		},
	}
}

type wafRateLimitingRuleProps struct {
	visibilityProps
	Name  string
	Regex string
	Limit int
}

func WafRateLimitingRuleProps(name, regex string, limit int, enableCloudWatchMetrics, enableSampleRequests bool) wafRateLimitingRuleProps {
	return wafRateLimitingRuleProps{
		Name:  name,
		Regex: regex,
		Limit: limit,
		visibilityProps: visibilityProps{
			EnableCloudWatchMetrics: enableCloudWatchMetrics,
			EnableSampleRequests:    enableSampleRequests,
		},
	}
}

func WafRateLimitingRule(props wafRateLimitingRuleProps) awswafv2.CfnWebACL_RuleProperty {
	nnv := struct{}{}
	return awswafv2.CfnWebACL_RuleProperty{
		Name: jsii.String(props.Name),
		Action: awswafv2.CfnWebACL_RuleActionProperty{
			Block: awswafv2.CfnWebACL_BlockActionProperty{},
		},
		VisibilityConfig: awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(props.EnableSampleRequests),
			CloudWatchMetricsEnabled: jsii.Bool(props.EnableCloudWatchMetrics),
			MetricName:               jsii.String(props.Name),
		},
		Statement: awswafv2.CfnWebACL_StatementProperty{
			RateBasedStatement: awswafv2.CfnWebACL_RateBasedStatementProperty{
				Limit:            jsii.Number(props.Limit),
				AggregateKeyType: jsii.String("IP"),
				ScopeDownStatement: scopeDown{awswafv2.CfnWebACL_RegexMatchStatementProperty{
					FieldToMatch: awswafv2.CfnWebACL_FieldToMatchProperty{
						UriPath: &nnv,
					},
					RegexString: &props.Regex,
					TextTransformations: []awswafv2.CfnWebACL_TextTransformationProperty{{
						Priority: jsii.Number(0),
						Type:     jsii.String("NONE"),
					}},
				}},
			},
		},
	}
}

type scopeDown struct {
	RegexMatchStatement interface{} `json:"regexMatchStatement" yaml:"regexMatchStatement"`
}

type wafIpConstraintRuleProps struct {
	visibilityProps
	ListName             string
	Scope                Scope
	AllowedIPV4Addresses []string
	AllowedIPV6Addresses []string
}

func WafIPWhitelistConstraintRuleProps(listName string, scope Scope, v4Addresses, v6addresses []string, enableCloudWatchMetrics, enableSampleRequests bool) wafIpConstraintRuleProps {
	return wafIpConstraintRuleProps{
		ListName:             listName,
		Scope:                scope,
		AllowedIPV4Addresses: v4Addresses,
		AllowedIPV6Addresses: v6addresses,
		visibilityProps: visibilityProps{
			EnableCloudWatchMetrics: enableCloudWatchMetrics,
			EnableSampleRequests:    enableSampleRequests,
		},
	}
}

func WafIPWhitelistConstraintRule(stack awscdk.Stack, props wafIpConstraintRuleProps) awswafv2.CfnWebACL_RuleProperty {
	var ipStatements []any

	if len(props.AllowedIPV4Addresses) > 0 {
		n := fmt.Sprintf("%sIPv4AllowedList", props.ListName)
		ipSet := awswafv2.NewCfnIPSet(stack, jsii.String(n), &awswafv2.CfnIPSetProps{
			Scope:            jsii.String(string(props.Scope)),
			Addresses:        jsii.Strings(props.AllowedIPV4Addresses...),
			IpAddressVersion: jsii.String("IPV4"),
		})
		ipStatements = append(ipStatements, createNotMatchIPStatement(ipSet))
	}

	if len(props.AllowedIPV6Addresses) > 0 {
		n := fmt.Sprintf("%sIPv6AllowedList", props.ListName)
		ipSet := awswafv2.NewCfnIPSet(stack, jsii.String(n), &awswafv2.CfnIPSetProps{
			Scope:            jsii.String(string(props.Scope)),
			Addresses:        jsii.Strings(props.AllowedIPV6Addresses...),
			IpAddressVersion: jsii.String("IPV6"),
		})
		ipStatements = append(ipStatements, createNotMatchIPStatement(ipSet))
	}

	var statement any
	if len(ipStatements) == 1 {
		// And statements must have at least two items.
		// So in this case we unwrap the list and just stick in the single NOT
		statement = ipStatements[0]
	} else {
		statement = awswafv2.CfnWebACL_StatementProperty{
			AndStatement: awswafv2.CfnRuleGroup_AndStatementProperty{
				Statements: ipStatements,
			},
		}
	}

	return awswafv2.CfnWebACL_RuleProperty{
		Name: jsii.String(fmt.Sprintf("%sWhitelistRule", props.ListName)),
		Action: awswafv2.CfnWebACL_RuleActionProperty{
			Block: awswafv2.CfnRuleGroup_BlockProperty{},
		},
		Statement: statement,
		VisibilityConfig: awswafv2.CfnWebACL_VisibilityConfigProperty{
			SampledRequestsEnabled:   jsii.Bool(props.EnableSampleRequests),
			CloudWatchMetricsEnabled: jsii.Bool(props.EnableCloudWatchMetrics),
			MetricName:               jsii.String(props.MetricName),
		},
	}
}

func createNotMatchIPStatement(ipSet awswafv2.CfnIPSet) awswafv2.CfnWebACL_StatementProperty {
	return awswafv2.CfnWebACL_StatementProperty{
		NotStatement: awswafv2.CfnRuleGroup_NotStatementProperty{
			Statement: awswafv2.CfnWebACL_StatementProperty{
				IpSetReferenceStatement: awswafv2.CfnRuleGroup_IPSetReferenceStatementProperty{
					Arn: ipSet.AttrArn(),
				},
			},
		},
	}
}
