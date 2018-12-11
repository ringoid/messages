package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"fmt"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
	"github.com/ringoid/commons"
	"github.com/ringoid/messages/apimodel"
)

func init() {
	apimodel.InitLambdaVars("internal-handle-stream-messages")
}

func handler(ctx context.Context, event events.KinesisEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	apimodel.Anlogger.Debugf(lc, "handle_stream.go : start handle request with [%d] records", len(event.Records))

	for _, record := range event.Records {
		body := record.Kinesis.Data

		var aEvent commons.BaseInternalEvent
		err := json.Unmarshal(body, &aEvent)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "handle_stream.go : error unmarshal body [%s] to BaseInternalEvent : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		apimodel.Anlogger.Debugf(lc, "handle_stream.go : handle record %v", aEvent)

		switch aEvent.EventType {
		case commons.UserMessageEvent:
			err = message(body, apimodel.MessageTableName, apimodel.AwsDynamoDbClient, lc, apimodel.Anlogger)
			if err != nil {
				return err
			}
		}
	}

	apimodel.Anlogger.Debugf(lc, "handle_stream.go : successfully complete handle request with [%d] records", len(event.Records))
	return nil
}

func main() {
	basicLambda.Start(handler)
}
