package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"strings"
	"github.com/ringoid/commons"
	"../apimodel"
)

func init() {
	apimodel.InitLambdaVars("warm-up-messages")
}

func handler(ctx context.Context, request events.CloudWatchEvent) error {
	lc, _ := lambdacontext.FromContext(ctx)
	names := strings.Split(apimodel.AllLambdaNames, ",")
	for _, n := range names {
		commons.WarmUpLambda(n, apimodel.AwsLambdaClient, apimodel.Anlogger, lc)
	}
	return nil
}

func main() {
	basicLambda.Start(handler)
}
