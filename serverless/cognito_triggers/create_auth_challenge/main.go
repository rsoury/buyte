/*
	Please refer to:
	https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-lambda-create-auth-challenge.html
*/

package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"go.uber.org/zap"
)

type CognitoDefineAuthChallengeResult struct {
	ChallengeName     string `json:"challengeName"`
	ChallengeResult   bool   `json:"challengeResult"`
	ChallengeMetadata string `json:"challengeMetadata"`
}
type CognitoCreateAuthChallengeRequest struct {
	UserAttributes map[string]string                  `json:"userAttributes"`
	ChallengeName  string                             `json:"challengeName"`
	Session        []CognitoDefineAuthChallengeResult `json:"session"`
}
type CognitoCreateAuthChallengeResponse struct {
	PublicChallengeParameters  map[string]string `json:"publicChallengeParameters"`
	PrivateChallengeParameters map[string]string `json:"privateChallengeParameters"`
	ChallengeMetadata          string            `json:"challengeMetadata"`
}
type CognitoCreateAuthChallenge struct {
	events.CognitoEventUserPoolsHeader
	Request  CognitoCreateAuthChallengeRequest  `json:"request"`
	Response CognitoCreateAuthChallengeResponse `json:"response"`
}

var Logger *zap.Logger

func init() {
	Logger, _ = zap.NewDevelopment()
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(event CognitoCreateAuthChallenge) (CognitoCreateAuthChallenge, error) {
	event.Response.PublicChallengeParameters = make(map[string]string)
	event.Response.PrivateChallengeParameters = make(map[string]string)

	defer Logger.Sync()

	Logger.Debug("Start Cognito Event", zap.Any("event", event))

	if event.Request.ChallengeName == "CUSTOM_CHALLENGE" {
		event.Response.ChallengeMetadata = "API_CREDENTIALS_CHALLENGE"

		region := os.Getenv("AWS_REGION")
		sess, _ := session.NewSession(
			&aws.Config{Region: aws.String(region)},
		)
		// Create an APIGateway client from a aws session
		svc := apigateway.New(sess)
		keyId := event.Request.UserAttributes["custom:secret_key_id"]
		includeValue := true
		apiKey, err := svc.GetApiKey(&apigateway.GetApiKeyInput{
			ApiKey:       &keyId,
			IncludeValue: &includeValue,
		})
		if err != nil {
			return event, err
		}
		// Need to instantiate these response values.
		event.Response.PrivateChallengeParameters["answer"] = *apiKey.Value
	}

	Logger.Debug("Response Cognito Event", zap.Any("event", event))
	return event, nil
}

func main() {
	lambda.Start(Handler)
}
