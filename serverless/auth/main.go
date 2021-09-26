package main

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/rsoury/buyte/pkg/authenticate"
)

const (
	AuthPattern = `^Bearer [-0-9a-zA-z\.]*$`
)

var (
	Logger *zap.Logger
)

func init() {
	Logger, _ = zap.NewProduction()
}

// Help function to generate an IAM policy
func AddPolicy(authResponse *events.APIGatewayCustomAuthorizerResponse, effect string, resource string) {
	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
}

func Handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	defer Logger.Sync()

	Logger.Info("Successfully enterered lambda handler",
		zap.Any("event", event),
	)

	// Validate Token first, then use it to get User and authenticate user.
	// Ensure Bearer Token for Ergonomics
	Authorization := event.Headers["Authorization"]
	if !regexp.MustCompile(AuthPattern).MatchString(Authorization) {
		Logger.Warn("Unauthorized at bearer", zap.String("Authorization", Authorization))
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Remove Bearer
	token := strings.Replace(Authorization, "Bearer ", "", 1)
	// Return invalid of token is tiny...
	if len(token) < 12 {
		Logger.Warn("Unauthorized at token length check", zap.String("Token", token))
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Create User
	user, err := authenticate.NewUserWithEnv(token)
	if err != nil {
		Logger.Error("Error creating user object to authenticate",
			zap.Error(err),
			zap.String("Token", token),
		)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Then check destination request path and if it starts with /public/
	if user.IsPublic {
		g := glob.MustCompile("/v*/public/**")
		if !g.Match(event.Path) {
			Logger.Warn("Unauthorized at public key vs path check",
				zap.String("Token", token),
				zap.String("Path", event.Path),
			)
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized") // Return a 401 Unauthorized response
		}
	}

	// Authenticate User.
	err = user.Authenticate()
	if err != nil {
		Logger.Error("Error authenticating user",
			zap.Error(err),
			zap.Any("user", user),
		)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	// Build allow response.
	authResponse := &events.APIGatewayCustomAuthorizerResponse{PrincipalID: user.Id, UsageIdentifierKey: user.Token}
	methodSplit := bytes.Split([]byte(event.MethodArn), []byte("/"))
	resource := bytes.Join(methodSplit[:2], []byte("/"))
	resource = append(resource, "/*"...) // The spread operator opens string as array of bytes as individual arguments
	AddPolicy(authResponse, "Allow", string(resource))

	// Stringify map[string]string
	userAttributes, err := json.Marshal(user.UserAttributes)
	if err != nil {
		Logger.Error("Error marshalling User Attributes into JSON String",
			zap.String("UserId", user.Id),
			zap.Any("UserAttributes", user.UserAttributes),
			zap.Error(err),
		)
	}

	authResponse.Context = map[string]interface{}{
		"userId":          user.Id,
		"token":           user.Token,
		"bareToken":       user.BareToken,
		"isPublic":        user.IsPublic,
		"isAuthenticated": user.IsAuthenticated,
		"userAttributes":  string(userAttributes),
		"accessToken":     *user.AuthenticationResult.AccessToken,
	}

	Logger.Info("Auth Response",
		zap.Any(`response`, *authResponse),
	)

	return *authResponse, nil
}

func main() {
	lambda.Start(Handler)
}
