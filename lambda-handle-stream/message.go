package main

import (
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/aws"
	"../apimodel"
)

func message(body []byte, userMessageTable string, awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {

	anlogger.Debugf(lc, "message.go : handle event and message, body %s", string(body))
	var aEvent commons.UserSendMessageEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "message.go : error unmarshal body [%s] to UserSendMessageEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}

	conversationId := apimodel.GenerateConversationId(aEvent.UserId, aEvent.TargetUserId)
	sortKey := fmt.Sprintf("%v_%v", aEvent.MessageAt, commons.UnixTimeInMillis())
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#senderId": aws.String(commons.MessagesSenderIdColumnName),
			"#text":     aws.String(commons.MessagesTextColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":senderIdV": {
				S: aws.String(aEvent.UserId),
			},
			":textV": {
				S: aws.String(aEvent.Text),
			},
		},

		Key: map[string]*dynamodb.AttributeValue{
			commons.MessagesConversationIdColumnName: {
				S: aws.String(conversationId),
			},
			commons.MessagesCreatedAtColumnName: {
				S: aws.String(sortKey),
			},
		},
		TableName:        aws.String(userMessageTable),
		UpdateExpression: aws.String("SET #senderId = :senderIdV, #text = :textV"),
	}

	_, err = awsDbClient.UpdateItem(input)
	if err != nil {
		anlogger.Errorf(lc, "message.go : error save message for conversationId [%s], userId [%s] : %v", conversationId, aEvent.UserId, err)
		return errors.New(fmt.Sprintf("error save message for userId [%s] : %v", aEvent.UserId, err))
	}

	anlogger.Debugf(lc, "message.go : successfully handle event and save message, body %s", string(body))
	return nil
}
