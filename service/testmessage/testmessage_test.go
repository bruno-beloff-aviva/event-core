package testmessage

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	message := NewTestMessage("client", "path")
	fmt.Println(message.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}

func TestNewMessageJSON(t *testing.T) {
	var message TestMessage
	var jmsg []byte
	var err error

	message = NewTestMessage("client", "path")
	jmsg, err = json.Marshal(message)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	fmt.Println(strmsg)

	err = json.Unmarshal([]byte(strmsg), &message)
	if err != nil {
		panic(err)
	}

	fmt.Println(message.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}
