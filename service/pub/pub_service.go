package pub

import (
	"context"
	"sqstest/service/testmessage"
)

type PubService interface {
	Publish(ctx context.Context, clientId string, path string) (testmessage.TestMessage, error)
}
