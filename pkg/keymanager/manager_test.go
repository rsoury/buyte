package keymanager

import (
	"testing"

	"github.com/rsoury/buyte/buyte"
)

var (
	Manager = NewKeyManager("c4cbfa77-300e-4d18-9b70-f1eada8b3aa8", "rsouryis@hotmail.com", &buyte.AWSConfig{
		Region:                "ap-southeast-2",
		APIGatewayId:          "uicv8bfuwi",
		APIGatewayName:        "buyte-primary-api-dev",
		APIGatewayStage:       "dev",
		APIGatewayUsagePlanId: "0jbhf5",
		CognitoUserPoolId:     "ap-southeast-2_1q1IuPEcJ",
	})
)

func TestGenerateKey(t *testing.T) {
	pkey := Manager.GenerateKey(true)
	t.Log(pkey)
	t.Log(pkey[:2])

	skey := Manager.GenerateKey(false)
	t.Log(skey)
	t.Log(skey[:2])
}

func TestCreateApiKey(t *testing.T) {
	pkey := Manager.GenerateKey(true)
	skey := Manager.GenerateKey(false)
	publicApiKey, err := Manager.CreateApiKey(pkey, true)
	if err != nil {
		t.Error("Error creating public key", err)
	}
	secretApiKey, err := Manager.CreateApiKey(skey, false)
	if err != nil {
		t.Error("Error creating secret key", err)
	}

	t.Log(publicApiKey)
	t.Log(secretApiKey)
	t.Log(publicApiKey.UsagePlanKey)
	t.Log(secretApiKey.UsagePlanKey)

	err = Manager.AssociateApiKeyIdsWithCognito([]*CognitoAssociateData{
		&CognitoAssociateData{
			IsPublic: true,
			ApiKeyId: *publicApiKey.Id,
		},
		&CognitoAssociateData{
			IsPublic: false,
			ApiKeyId: *secretApiKey.Id,
		},
	})
	if err != nil {
		t.Error("Error Associating Api Key ids with Congito", err)
	}

	err = Manager.DeleteApiKey(*publicApiKey.Id)
	if err != nil {
		t.Error("Error deleting public key", err)
	}
	err = Manager.DeleteApiKey(*secretApiKey.Id)
	if err != nil {
		t.Error("Error deleting secret key", err)
	}
}
