package testreception

import (
	"encoding/json"
	"fmt"
	"sqstest/service/testmessage"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewReception(t *testing.T) {
	message := testmessage.NewTestMessage("client", "path")
	reception := NewTestReception("sub1", message)

	fmt.Println(reception.String())

	assert.Equal(t, message.Client, "client")
	assert.Equal(t, message.Path, "path")
}

func TestNewReceptionJSON(t *testing.T) {
	var message testmessage.TestMessage
	var reception TestReception
	var jmsg []byte
	var err error

	message = testmessage.NewTestMessage("client", "path")
	time.Sleep(1 * time.Second)

	reception = NewTestReception("sub1", message)

	jmsg, err = json.Marshal(reception)
	strmsg := string(jmsg)

	if err != nil {
		panic(err)
	}

	fmt.Println(strmsg)

	err = json.Unmarshal([]byte(strmsg), &reception)
	if err != nil {
		panic(err)
	}

	fmt.Println(reception.String())

	assert.Equal(t, reception.Client, "client")
	assert.Equal(t, reception.Path, "path")
}
