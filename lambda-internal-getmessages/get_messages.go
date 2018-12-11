package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ringoid/commons"
	"../apimodel"
)

func init() {
	apimodel.InitLambdaVars("internal-get-messages-messages")
}

func handler(ctx context.Context, request commons.InternalGetMessagesReq) (commons.InternalGetMessagesResp, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	apimodel.Anlogger.Debugf(lc, "get_messages.go : start handle request, isItWarmUpRequest [%v],  request %v", request.WarmUpRequest, request)

	if request.WarmUpRequest {
		return commons.InternalGetMessagesResp{}, nil
	}

	respChan := make(chan map[string][]commons.Message)
	requestCounter := 0

	for _, eachTargetUserId := range request.TargetUserIds {
		conversionId := apimodel.GenerateConversationId(request.SourceUserId, eachTargetUserId)
		go conversation(conversionId, eachTargetUserId, respChan, lc)
		requestCounter++
	}

	allMessageNum := 0
	finalMap := make(map[string][]commons.Message)
	for i := 0; i < requestCounter; i++ {
		resMap := <-respChan
		for k, v := range resMap {
			finalMap[k] = v
			allMessageNum += len(v)
		}
	}

	resp := commons.InternalGetMessagesResp{
		ConversationsMap: finalMap,
	}

	apimodel.Anlogger.Debugf(lc, "get_messages.go : return successful resp %v", resp)

	apimodel.Anlogger.Infof(lc, "get_messages.go : successfully return [%d] conversations with all messages num [%d] for userId [%s]",
		len(resp.ConversationsMap), allMessageNum, request.SourceUserId)

	return resp, nil
}

func conversation(conversationId, targetUserId string, respChan chan<- map[string][]commons.Message, lc *lambdacontext.LambdaContext) {
	apimodel.Anlogger.Debugf(lc, "get_messages.go : make query to fetch conversation, conversationId [%s], targetUserId [%s]",
		conversationId, targetUserId)

	var exclusiveStartKey map[string]*dynamodb.AttributeValue
	finalResultMap := make(map[string][]commons.Message)

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
			ExclusiveStartKey: exclusiveStartKey,
			ConsistentRead:    aws.Bool(true),
			//ScanIndexForward:       aws.Bool(false),
			KeyConditionExpression: aws.String("#conversionId = :conversionIdV"),
			TableName:              aws.String(apimodel.MessageTableName),
		}

		result, err := apimodel.AwsDynamoDbClient.Query(input)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "get_messages.go : error while reading conversation [%s] : %v", conversationId, err)
			respChan <- finalResultMap
			return
		}

		messages := make([]commons.Message, 0)
		for _, item := range result.Items {
			senderId := *item[commons.MessagesSenderIdColumnName].S
			text := *item[commons.MessagesTextColumnName].S
			messages = append(messages, commons.Message{
				WasYouSender: !(senderId == targetUserId),
				Text:         text,
			})
		}
		apimodel.Anlogger.Debugf(lc, "get_messages.go : fetch [%d] messages for conversation [%s]", len(messages), conversationId)

		messageSlice, ok := finalResultMap[targetUserId]
		if !ok {
			messageSlice = messages
		} else {
			messageSlice = append(messageSlice, messages...)
		}
		finalResultMap[targetUserId] = messageSlice

		if len(result.LastEvaluatedKey) == 0 {
			break
		}

		exclusiveStartKey = result.LastEvaluatedKey
	}

	apimodel.Anlogger.Debugf(lc, "get_messages.go : successfully return [%d] messages for conversation [%s]", finalResultMap[targetUserId], conversationId)
	respChan <- finalResultMap
}

func main() {
	basicLambda.Start(handler)
}
