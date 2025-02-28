package sub

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type SubService interface {
	Handle(ctx context.Context, record events.SQSMessage) (err error)
}
