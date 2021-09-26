/*
	A Custom Authentication Flow to essentially Authenticate a User with their Secret Key...

	Please refer to:
	https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-lambda-challenge.html
	https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-lambda-define-auth-challenge.html#aws-lambda-triggers-define-auth-challenge-example

	There may be some nuances that do not correlate to AWS Docs.
	Please see:
	https://stackoverflow.com/questions/54238884/unrecognized-verify-auth-challenge-lambda-response-c-sharp
*/

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

type CognitoDefineAuthChallengeResult struct {
	ChallengeName     string `json:"challengeName"`
	ChallengeResult   bool   `json:"challengeResult"`
	ChallengeMetadata string `json:"challengeMetadata"`
}
type CognitoDefineAuthChallengeRequest struct {
	UserAttributes map[string]string                  `json:"userAttributes"`
	Session        []CognitoDefineAuthChallengeResult `json:"session"`
}
type CognitoDefineAuthChallengeResponse struct {
	ChallengeName      string `json:"challengeName"`
	IssueTokens        bool   `json:"issueTokens"`
	FailAuthentication bool   `json:"failAuthentication"`
}
type CognitoDefineAuthChallenge struct {
	events.CognitoEventUserPoolsHeader
	Request  CognitoDefineAuthChallengeRequest  `json:"request"`
	Response CognitoDefineAuthChallengeResponse `json:"response"`
}

var Logger *zap.Logger

func init() {
	Logger, _ = zap.NewDevelopment()
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(event CognitoDefineAuthChallenge) (CognitoDefineAuthChallenge, error) {
	defer Logger.Sync()

	Logger.Debug("Start Cognito Event", zap.Any("event", event))
	if len(event.Request.Session) > 0 {
		if len(event.Request.Session) == 1 {
			if event.Request.Session[0].ChallengeName == "CUSTOM_CHALLENGE" && event.Request.Session[0].ChallengeMetadata == "API_CREDENTIALS_CHALLENGE" {
				if event.Request.Session[0].ChallengeResult {
					event.Response.IssueTokens = true
					return event, nil
				}
			}
		}
		event.Response.FailAuthentication = true
	} else {
		event.Response.ChallengeName = "CUSTOM_CHALLENGE"
	}
	Logger.Debug("Response Cognito Event", zap.Any("event", event))
	return event, nil
}

func main() {
	lambda.Start(Handler)
}
