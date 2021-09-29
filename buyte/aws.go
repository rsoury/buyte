package buyte

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

// Compatible with "github.com/caarlos0/env"
type AWSConfig struct {
	Region                string `env:"AWS_REGION" envDefault:"ap-southeast-2"`
	APIGatewayId          string `env:"API_GATEWAY_ID"`
	APIGatewayName 				string `env:"API_GATEWAY_NAME"`
	APIGatewayStage       string `env:"API_GATEWAY_STAGE"`
	APIGatewayUsagePlanId string `env:"API_GATEWAY_USAGE_PLAN_ID"`
	CognitoUserPoolId     string `env:"COGNITO_USERPOOLID"`
	CognitoClientId     	string `env:"COGNITO_CLIENTID"`
}

func NewEnvConfig() *AWSConfig {
	config := &AWSConfig{}
	err := env.Parse(config)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot Marshal Environment into Config"))
	}
	return config
}