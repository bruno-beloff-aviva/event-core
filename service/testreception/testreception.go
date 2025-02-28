package testreception

import (
	"fmt"
	"sqstest/manager/dbmanager"
	"sqstest/service/testmessage"
	"time"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
)

var DeletionKeys = []string{"PK", "Received"}

type TestReception struct {
	testmessage.TestMessage
	PK         string
	Received   string
	Subscriber string
}

func DynamoPartitionKey() *awsdynamodb.Attribute {
	return dbmanager.StringAttribute("PK")
}

func DynamoSortKey() *awsdynamodb.Attribute {
	return dbmanager.StringAttribute("Received")
}

func NewTestReception(subscriber string, message testmessage.TestMessage) TestReception {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pk := message.Sent + "/" + subscriber

	return TestReception{TestMessage: message, PK: pk, Received: now, Subscriber: subscriber}
}

func (r *TestReception) String() string {
	return fmt.Sprintf("TestReception:{Received:%s Subscriber:%s Sent:%s Path:%s Client:%s}", r.Received, r.Subscriber, r.Sent, r.Path, r.Client)
}

func (r *TestReception) PartitionKey() map[string]any {
	return map[string]any{"PK": r.PK}
}
