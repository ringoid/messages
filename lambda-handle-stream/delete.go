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

func delete(body []byte, userMessageTable string, awsDbClient *dynamodb.DynamoDB, lc *lambdacontext.LambdaContext, anlogger *commons.Logger) error {

	anlogger.Debugf(lc, "delete.go : handle event and message, body %s", string(body))
	var aEvent commons.DeleteUserConversationEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		anlogger.Errorf(lc, "delete.go : error unmarshal body [%s] to DeleteUserConversationEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}

	conversationId := apimodel.GenerateConversationId(aEvent.UserId, aEvent.TargetUserId)

	for {
		input := &dynamodb.QueryInput{
			ExpressionAttributeNames: map[string]*string{
				"#conversionId": aws.String(commons.MessagesConversationIdColumnName),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":conversionIdV": {
					S: aws.String(conversationId),
				},
			},
			ConsistentRead:         aws.Bool(true),
			KeyConditionExpression: aws.String("#conversionId = :conversionIdV"),
			TableName:              aws.String(userMessageTable),
			Limit:                  aws.Int64(25), //limit for batchWriteItem
		}

		result, err := awsDbClient.Query(input)
		if err != nil {
			anlogger.Errorf(lc, "delete.go : error while reading conversation [%s] : %v", conversationId, err)
			return errors.New(fmt.Sprintf("error reading conversation [%s] : %v", conversationId, err))
		}

		if len(result.Items) == 0 {
			break
		}

		batchDelInput := &dynamodb.BatchWriteItemInput{}
		requestItems := make(map[string][]*dynamodb.WriteRequest)
		batchDelInput.RequestItems = requestItems
		deleteRequests := make([]*dynamodb.WriteRequest, 0)

		for _, item := range result.Items {
			key := make(map[string]*dynamodb.AttributeValue)
			key[commons.MessagesConversationIdColumnName] = item[commons.MessagesConversationIdColumnName]
			key[commons.MessagesCreatedAtColumnName] = item[commons.MessagesCreatedAtColumnName]
			deleteRequests = append(deleteRequests,
				&dynamodb.WriteRequest{
					DeleteRequest: &dynamodb.DeleteRequest{
						Key: key,
					},
				})
		}
		requestItems[userMessageTable] = deleteRequests

		_, err = awsDbClient.BatchWriteItem(batchDelInput)
		if err != nil {
			anlogger.Errorf(lc, "delete.go : error while deleting conversation [%s] : %v", conversationId, err)
			return errors.New(fmt.Sprintf("error deleting conversation [%s] : %v", conversationId, err))
		}
	}

	anlogger.Debugf(lc, "delete.go : successfully handle event and save message, body %s", string(body))
	return nil
}
