/*
	A function triggered after an account has been verified by email confirmation
	Inside this repo because this API is meant to depend on the values produced by the Amplify Dashboard project. Keep tree of dependencies consistent.

	In this function we will:
	- Send a welcome email -- LATER
	- Create api keys for the user, associate the user id with the api keys, and set api key ids as values for user attributes of user.
*/

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/keymanager"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	userId := event.UserName
	email := event.Request.UserAttributes["email"]
	envConfig := buyte.NewEnvConfig()
	manager := keymanager.NewKeyManager(userId, email, envConfig)

	/*
		TODO: Optimise this by running both api key creations in goroutines.
	*/

	// Public Key
	publicKey := manager.GenerateKey(true)
	publicApiKey, err := manager.CreateApiKey(publicKey, true)
	if err != nil {
		return event, errors.Wrapf(err, "Could not create public API key for user %s", manager.UserId)
	}

	// Secret Key
	secretKey := manager.GenerateKey(false)
	secretApiKey, err := manager.CreateApiKey(secretKey, false)
	if err != nil {
		return event, errors.Wrapf(err, "Could not create secret API key for user %s", manager.UserId)
	}

	// Associate
	err = manager.AssociateApiKeyIdsWithCognito([]*keymanager.CognitoAssociateData{
		&keymanager.CognitoAssociateData{
			IsPublic: true,
			ApiKeyId: *publicApiKey.Id,
		},
		&keymanager.CognitoAssociateData{
			IsPublic: false,
			ApiKeyId: *secretApiKey.Id,
		},
	})
	if err != nil {
		return event, errors.Wrapf(err, "Could associate API Keys with Cognito Account for user %s", manager.UserId)
	}

	return event, nil
}

func main() {
	lambda.Start(Handler)
}
