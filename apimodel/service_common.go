package apimodel

import (
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/firehose"
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"sort"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var Anlogger *commons.Logger
var DeliveryStreamName string
var AwsDeliveryStreamClient *firehose.Firehose
var AwsDynamoDbClient *dynamodb.DynamoDB
var AwsLambdaClient *lambda.Lambda

var MessageTableName string
var AllLambdaNames string

func InitLambdaVars(lambdaName string) {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : service_common.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : service_common.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : service_common.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : service_common.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	Anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, lambdaName))
	if err != nil {
		fmt.Errorf("lambda-initialization : service_common.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : logger was successfully initialized")

	DeliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty DELIVERY_STREAM")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with DELIVERY_STREAM = [%s]", DeliveryStreamName)

	MessageTableName, ok = os.LookupEnv("MESSAGE_TABLE_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty MESSAGE_TABLE_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with MESSAGE_TABLE_NAME = [%s]", MessageTableName)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { Anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : error during initialization : %v", err)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : aws session was successfully initialized")

	AllLambdaNames, ok = os.LookupEnv("NEED_WARM_UP_LAMBDA_NAMES")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : warm_up_image.go : env can not be empty NEED_WARM_UP_LAMBDA_NAMES")
	}
	Anlogger.Debugf(nil, "lambda-initialization : warm_up_image.go : start with NEED_WARM_UP_LAMBDA_NAMES = [%s]", AllLambdaNames)

	AwsDeliveryStreamClient = firehose.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : firehose client was successfully initialized")

	AwsDynamoDbClient = dynamodb.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : dynamodb client was successfully initialized")

	AwsLambdaClient = lambda.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : lambda client was successfully initialized")
}

func GenerateConversationId(arr ... string) string {
	sort.Strings(arr)
	if len(arr) == 0 {
		return ""
	}
	key := arr[0]
	for index, each := range arr {
		if index != 0 {
			key += "_" + each
		}
	}
	return key
}
