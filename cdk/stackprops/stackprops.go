package stackprops

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type CdkStackProps struct {
	awscdk.StackProps
	Version string
}

func (p CdkStackProps) NewStack(scope constructs.Construct, id string) awscdk.Stack {
	return awscdk.NewStack(scope, &id, &p.StackProps)

}
