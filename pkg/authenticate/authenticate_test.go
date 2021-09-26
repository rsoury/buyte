package authenticate

import (
	"testing"
)

var (
	PublicKey = "pk_YICkLIG2LPQfZvM3YE00LSDtXIQ5KvYfAPGiYSKkYPNyAPM2Xy9j"
	SecretKey = "sk_RBVdEBZ2EIJySoF3RX00ELWmQBJ5DoRyTIZbRLDdRIGrTIF2Qpi1"
)

func TestAuthenticate(t *testing.T) {
	config := &AWSConfig{
		Region:            "ap-southeast-2",
		CognitoUserPoolId: "ap-southeast-2_1q1IuPEcJ",
		CognitoClientId:   "1pmlfaqvc86c1csiq7chsbc7b8",
	}
	publicUser, err := NewUser(PublicKey, config)
	if err != nil {
		t.Error("Error creating public user", err)
	}
	err = publicUser.Authenticate()
	if err != nil {
		t.Error("Error authenticating public user", err)
	}
	t.Log(publicUser)

	secretUser, err := NewUser(SecretKey, config)
	if err != nil {
		t.Error("Error creating secret user", err)
	}
	err = secretUser.Authenticate()
	if err != nil {
		t.Error("Error authenticating secret user", err)
	}
	t.Log(secretUser)
}
