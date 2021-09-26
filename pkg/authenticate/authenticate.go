package authenticate

import (
	"bytes"
	"encoding/base64"

	"github.com/alexjohnj/caesar"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type AWSConfig struct {
	Region            string `env:"AWS_REGION" envDefault:"ap-southeast-2"`
	CognitoUserPoolId string `env:"COGNITO_USERPOOLID"`
	CognitoClientId   string `env:"COGNITO_CLIENTID"`
}

type User struct {
	Id                   string
	Token                string
	BareToken            string
	IsPublic             bool
	IsAuthenticated      bool
	UserAttributes       map[string]string
	AuthenticationResult *cognito.AuthenticationResultType
	Config               *AWSConfig
}

const (
	PublicKeyCipherNumber = 12
	SecretKeyCipherNumber = 5
	PublicDescriptor      = "pk"
	SecretDescriptor      = "sk"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	// Read Env.
	logger = zap.S().With("package", "authenticate")
}

func NewUserWithEnv(token string) (*User, error) {
	return NewUser(token, NewEnvConfig())
}

func NewEnvConfig() *AWSConfig {
	cfg := &AWSConfig{}
	err := env.Parse(cfg)
	if err != nil {
		panic(errors.New("Cannot Marshal Environment into Config."))
	}
	return cfg
}

func NewUser(token string, config *AWSConfig) (*User, error) {
	// Separate descriptor
	descriptorSplit := bytes.Split([]byte(token), []byte("_"))
	descriptor := string(descriptorSplit[0])
	bareTokenBytes := bytes.Join(descriptorSplit[1:], []byte("_"))
	bareToken := string(bareTokenBytes)

	user := &User{
		Token:     token,
		BareToken: bareToken,
		Config:    config,
	}

	// Determine if public.
	var isPublic bool
	if descriptor == PublicDescriptor {
		isPublic = true
	} else if descriptor == SecretDescriptor {
		isPublic = false
	} else {
		return user, errors.New("Token descriptor not valid.")
	}
	user.IsPublic = isPublic

	// obtain user id
	userId, err := user.DecodeToken()
	if err != nil {
		return user, err
	}
	user.Id = userId

	return user, nil
}

func (u *User) Authenticate() error {
	region := u.Config.Region
	userPoolId := u.Config.CognitoUserPoolId
	clientId := u.Config.CognitoClientId

	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(region)},
	)
	cognitoSvc := cognito.New(sess)

	// Create an APIGateway client from a aws session
	user, err := cognitoSvc.AdminGetUser(&cognito.AdminGetUserInput{
		UserPoolId: &userPoolId,
		Username:   &u.Id,
	})
	if err != nil {
		logger.Errorw("Getting user", "user", u.Id, "error", err)
		return err
	}

	// Unauthorized if User is not valid...
	if !*user.Enabled || *user.UserStatus != "CONFIRMED" {
		var errMsg string
		if !*user.Enabled {
			errMsg = "Attempting to authorize a user that is disabled."
		}
		if *user.UserStatus != "CONFIRMED" {
			errMsg = "Attempting to authorize a user that is not confirmed."
		}
		return errors.New(errMsg)
	}

	// Set User attributes
	u.UserAttributes = make(map[string]string)
	for _, attribute := range user.UserAttributes {
		u.UserAttributes[*attribute.Name] = *attribute.Value
	}

	// Can get user in a goroutine if isPublic == false
	// Get Secret key if isPublic
	var secretToken string
	if u.IsPublic {
		// Create an APIGateway client from a aws session
		apigatewaySvc := apigateway.New(sess)
		keyId := u.UserAttributes["custom:secret_key_id"]
		includeValue := true
		apiKey, err := apigatewaySvc.GetApiKey(&apigateway.GetApiKeyInput{
			ApiKey:       &keyId,
			IncludeValue: &includeValue,
		})
		if err != nil {
			logger.Errorw("Getting Api Key", "user", u.Id, "error", err)
			return err
		}
		secretToken = *apiKey.Value
	} else {
		secretToken = u.Token
	}

	// Initiate Authentication
	authFlow := "CUSTOM_AUTH"
	initiateAuthResponse, err := cognitoSvc.AdminInitiateAuth(&cognito.AdminInitiateAuthInput{
		AuthFlow: &authFlow,
		AuthParameters: map[string]*string{
			"USERNAME": &u.Id,
		},
		ClientId:   &clientId,
		UserPoolId: &userPoolId,
	})
	if err != nil {
		logger.Errorw("Admin initiate auth", "user", u.Id, "error", err)
		return err
	}

	// Respond to Authentication
	result, err := cognitoSvc.AdminRespondToAuthChallenge(&cognito.AdminRespondToAuthChallengeInput{
		UserPoolId:    &userPoolId,
		ClientId:      &clientId,
		ChallengeName: initiateAuthResponse.ChallengeName,
		ChallengeResponses: map[string]*string{
			"USERNAME": &u.Id,
			"ANSWER":   &secretToken,
		},
		Session: initiateAuthResponse.Session,
	})
	if err != nil {
		logger.Errorw("Admin respond to auth challenge", "user", u.Id, "error", err)
		return err
	}

	u.AuthenticationResult = result.AuthenticationResult
	u.IsAuthenticated = true

	return nil
}

// bareToken does not include it's descriptor.
func (u *User) DecodeToken() (string, error) {
	if u.BareToken == "" {
		return "", errors.New("Provided token is empty...")
	}

	// Decrypt
	var CipherNumber int
	if u.IsPublic {
		CipherNumber = PublicKeyCipherNumber
	} else {
		CipherNumber = SecretKeyCipherNumber
	}
	plaintext := caesar.DecryptCiphertext(u.BareToken, CipherNumber)

	// Decode Base64
	rawBytes, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return "", errors.Wrapf(err, "Could not decode string %s using Base64 Decoding...", plaintext)
	}

	// Remove random string
	userId := bytes.Split(rawBytes, []byte("."))[0]
	userIdStr := string(userId)

	return userIdStr, nil
}
