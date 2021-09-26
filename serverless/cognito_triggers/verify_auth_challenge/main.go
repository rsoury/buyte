/*
	Please refer to:
	https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-lambda-verify-auth-challenge-response.html#aws-lambda-triggers-verify-auth-challenge-response-example
*/

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

type CognitoVerifyAuthChallengeRequest struct {
	UserAttributes             map[string]string `json:"userAttributes"`
	PrivateChallengeParameters map[string]string `json:"privateChallengeParameters"`
	ChallengeAnswer            string            `json:"challengeAnswer"`
}
type CognitoVerifyAuthChallengeResponse struct {
	AnswerCorrect bool `json:"answerCorrect"`
}
type CognitoVerifyAuthChallenge struct {
	events.CognitoEventUserPoolsHeader
	Request  CognitoVerifyAuthChallengeRequest  `json:"request"`
	Response CognitoVerifyAuthChallengeResponse `json:"response"`
}

var Logger *zap.Logger

func init() {
	Logger, _ = zap.NewDevelopment()
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(event CognitoVerifyAuthChallenge) (CognitoVerifyAuthChallenge, error) {
	defer Logger.Sync()

	Logger.Debug("Verify events",
		zap.String("answer", event.Request.PrivateChallengeParameters["answer"]),
		zap.String("submitted answer", event.Request.ChallengeAnswer),
	)

	if event.Request.PrivateChallengeParameters["answer"] == event.Request.ChallengeAnswer {
		event.Response.AnswerCorrect = true
	}

	Logger.Debug("Response Verify Events", zap.Any("event", event))

	return event, nil
}

func main() {
	lambda.Start(Handler)
}
