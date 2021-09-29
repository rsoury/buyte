package keymanager

import (
	"encoding/base64"
	"log"

	"github.com/alexjohnj/caesar"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/thanhpk/randstr"

	"github.com/rsoury/buyte/buyte"
)

const (
	RawDataLength          = 42
	PublicKeyCipherNumber  = 12
	SecretKeyCipherNumber  = 5
	CongitoAttrPublicKeyId = "custom:public_key_id"
	CongitoAttrSecretKeyId = "custom:secret_key_id"
)

type ApiKey struct {
	*apigateway.ApiKey
	UsagePlanKey *apigateway.UsagePlanKey
}
type CognitoAssociateData struct {
	IsPublic bool
	ApiKeyId string
}

type Keys struct {
	UserId string
	Name   string
	Config *buyte.AWSConfig
}

func NewKeyManager(userId string, name string, awsConfig *buyte.AWSConfig) *Keys {
	return &Keys{
		UserId: userId,
		Name:   name,
		Config: awsConfig,
	}
}

func (k *Keys) GenerateKey(isPublic bool) string {
	// Make a 64 byte sized key -- Minus 1 for the dot point, minus 3 for the descriptor
	randKey := randstr.String(RawDataLength - len(k.UserId) - 4)
	data := base64.StdEncoding.EncodeToString([]byte(k.UserId + "." + randKey))

	// Prepend descriptor
	if isPublic {
		data = "pk_" + caesar.EncryptPlaintext(data, PublicKeyCipherNumber)
	} else {
		data = "sk_" + caesar.EncryptPlaintext(data, SecretKeyCipherNumber)
	}

	return data
}

func (k *Keys) CreateApiKey(key string, isPublic bool) (*ApiKey, error) {
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(k.Config.Region)},
	)
	// Create an APIGateway client from a aws session
	svc := apigateway.New(sess)

	// Create API Key
	input := &apigateway.CreateApiKeyInput{}
	input.SetEnabled(true)
	input.SetGenerateDistinctId(true)
	inputDesc := "key for the Buyte Primary API"
	var inputName string
	if isPublic {
		inputDesc = "Public " + inputDesc
		inputName = "public-" + k.Name + "-" + xid.New().String()
	} else {
		inputDesc = "Secret " + inputDesc
		inputName = "secret-" + k.Name + "-" + xid.New().String()
	}
	input.SetName(inputName)
	input.SetDescription(inputDesc)
	input.SetValue(key)
	stageKey := &apigateway.StageKey{}
	stageKey.SetRestApiId(k.Config.APIGatewayId)
	stageKey.SetStageName(k.Config.APIGatewayStage)
	input.SetStageKeys([]*apigateway.StageKey{stageKey})

	apiKey, err := svc.CreateApiKey(input)
	if err != nil {
		return nil, errors.Wrap(err, "Could not Create Api Key")
	}

	// Associate API Key to Usage Plan, if fails, delete the key.
	associateInput := &apigateway.CreateUsagePlanKeyInput{}
	associateInput.SetKeyId(*apiKey.Id)
	associateInput.SetKeyType("API_KEY")
	associateInput.SetUsagePlanId(k.Config.APIGatewayUsagePlanId)
	usagePlanKey, err := svc.CreateUsagePlanKey(associateInput)
	if err != nil {
		log.Println("Deleting API Key ...")
		_ = k.DeleteApiKey(*apiKey.Id)
		return nil, errors.Wrapf(err, "Could not associate Api Key %s with Usage Plan %s", *apiKey.Id, k.Config.APIGatewayUsagePlanId)
	}

	return &ApiKey{apiKey, usagePlanKey}, nil
}

func (k *Keys) AssociateApiKeyIdsWithCognito(data []*CognitoAssociateData) error {
	// isPublic bool, apiKeyId string
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(k.Config.Region)},
	)
	// Create an APIGateway client from a aws session
	svc := cognito.New(sess)

	var attributes []*cognito.AttributeType
	for _, value := range data {
		var Name string
		if value.IsPublic {
			Name = CongitoAttrPublicKeyId
		} else {
			Name = CongitoAttrSecretKeyId
		}
		attributes = append(attributes, &cognito.AttributeType{
			Name:  &Name,
			Value: &value.ApiKeyId,
		})
	}

	input := &cognito.AdminUpdateUserAttributesInput{}
	input.SetUserPoolId(k.Config.CognitoUserPoolId)
	input.SetUsername(k.UserId)
	input.SetUserAttributes(attributes)

	err := input.Validate()
	if err != nil {
		return errors.Wrap(err, "Error with the input when associating Api Key ids with Cognito")
	}

	_, err = svc.AdminUpdateUserAttributes(input)
	if err != nil {
		return errors.Wrap(err, "Cannot update user attributes")
	}

	return nil
}

// Just some helpers...
func (k *Keys) DeleteApiKey(apiKeyId string) error {
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(k.Config.Region)},
	)

	svc := apigateway.New(sess)
	input := &apigateway.DeleteApiKeyInput{}
	input.SetApiKey(apiKeyId)
	_, err := svc.DeleteApiKey(input)
	if err != nil {
		return errors.Wrap(err, "Could not Delete Api Key")
	}

	return nil
}

func (k *Keys) DeleteUserKeys() error {
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(k.Config.Region)},
	)
	cognitoSvc := cognito.New(sess)
	user, err := cognitoSvc.AdminGetUser(&cognito.AdminGetUserInput{
		UserPoolId: &k.Config.CognitoUserPoolId,
		Username:   &k.UserId,
	})
	if err != nil {
		return err
	}

	for _, attr := range user.UserAttributes {
		if *attr.Name == CongitoAttrPublicKeyId {
			k.DeleteApiKey(*attr.Value)
		}
		if *attr.Name == CongitoAttrSecretKeyId {
			k.DeleteApiKey(*attr.Value)
		}
	}

	return nil
}
