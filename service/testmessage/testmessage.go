package testmessage

import (
	"fmt"
	"time"
)

type TestMessage struct {
	Sent   string
	Path   string
	Client string
}

func NewTestMessage(client string, path string) TestMessage {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	return TestMessage{Sent: now, Path: path, Client: client}
}

func (m *TestMessage) String() string {
	return fmt.Sprintf("TestMessage:{Sent:%s Path:%s Client:%s}", m.Sent, m.Path, m.Client)
}
