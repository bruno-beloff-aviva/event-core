package testmessage

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRecord(t *testing.T) {
	event := "{\n  \"Type\" : \"Notification\",\n  \"MessageId\" : \"6eccfb51-3aa7-540d-ae24-9b79c28e0437\",\n  \"TopicArn\" : \"arn:aws:sns:eu-west-2:673007244143:SQS1Stack-SQS1SQS1TestTopic4F5B763D-XGSZiGPVpt8o\",\n  \"Message\" : \"{\\\"Sent\\\":\\\"2025-02-14T07:51:04.684550357Z\\\",\\\"Path\\\":\\\"/test1/ok1\\\",\\\"Client\\\":\\\"31.94.60.181\\\"}\",\n  \"Timestamp\" : \"2025-02-14T07:51:05.183Z\",\n  \"SignatureVersion\" : \"1\",\n  \"Signature\" : \"Ao1HVZACgCV+C3YiuBLi3soqq7i8U6LZDDs1NpsBlroaaEhFA1W2D2ba4wnqjUdSeno78GRO7YD0vtjvyew0XLQEucUD40ANYzs46dKw0JhH57UCra9oL1XXJBEBWFFuhAY4BqODdE1rqfWlB+JHpjxe4Q12bBDfs3aE46ymDEQOjP7i9TiB1trNug6SZyQqxJqISs2CRdSh4+Q7igz5NVhpAfsJiUw4s9gAlSV/grUH/GaAzv8kITAUegMYE40qU1gCcNXE+uFXxUsfrjZNdfaL/vbQd20NjL+CL0aSZaKTKsZdEh8fj/3K8T+iKujOm0scTZW942XcnDabmpfh1Q==\",\n  \"SigningCertURL\" : \"https://sns.eu-west-2.amazonaws.com/SimpleNotificationService-9c6465fa7f48f5cacd23014631ec1136.pem\",\n  \"UnsubscribeURL\" : \"https://sns.eu-west-2.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:eu-west-2:673007244143:SQS1Stack-SQS1SQS1TestTopic4F5B763D-XGSZiGPVpt8o:fd7f16fd-8040-41fa-ad76-24f73ed90d75\"\n}"
	fmt.Println(event)
	fmt.Println("-")

	var record map[string]string
	var message TestMessage
	var err error

	err = json.Unmarshal([]byte(event), &record)
	if err != nil {
		panic(err)
	}

	fmt.Println(record)
	fmt.Println("-")

	// messageStr, ok := record["Message"].(string)
	// if !ok {
	// 	panic("record[\"Message\"] is not a string")
	// }
	err = json.Unmarshal([]byte(record["Message"]), &message)
	if err != nil {
		panic(err)
	}

	fmt.Println(message.String())
}
