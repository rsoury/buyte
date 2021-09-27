package stripe

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rsoury/buyte/buyte"
)

const credentials = `{
	"accessToken": "sk_test_xxxx",
	"refreshToken": "rt_xxxx",
	"scope": "read_write",
	"stripePublishableKey": "pk_test_xxx",
	"stripeUserId": "acct_xxx",
	"tokenType": "bearer"
}`
const networkTokenData = `{
    "applicationPrimaryAccountNumber": "4817499130172785",
    "applicationExpirationDate": "231231",
    "currencyCode": "36",
    "transactionAmount": 1,
    "deviceManufacturerIdentifier": "040010030273",
    "paymentDataType": "3DSecure",
    "paymentData": {
        "onlinePaymentCryptogram": "Ag0wIaIAHrzC2TyUMqHLMAABAAA=",
        "eciIndicator": "7"
    }
}`

const connectCredentials = `{
	"stripeUserId": "acct_1FBaGoKVy2PztxCa",
	"isConnect": true
}`

// For now, you need to provide a payment token id...
// In the future, we will try to mock this process...
var (
	chargeInput = &buyte.CreateChargeInput{
		Amount:      3200,
		Currency:    "aud",
		Description: "This is a Adyen test",
		Metadata: map[string]interface{}{
			"Test": true,
		},
	}
	paymentToken = &buyte.PaymentToken{
		PaymentMethod: &buyte.PaymentMethod{
			Name: "Apple Pay",
		},
	}
)

func TestCharge(t *testing.T) {
	gateway, err := New(context.Background(), &buyte.ProviderCheckoutConnection{
		IsTest:      true,
		Credentials: credentials,
		Provider: buyte.ProviderCheckoutConnectionProviderDetails{
			Name: "Stripe",
		},
	})
	if err != nil {
		t.Error(err)
	}

	networkToken := &buyte.NetworkToken{}
	err = json.Unmarshal([]byte(networkTokenData), networkToken)
	if err != nil {
		t.Error(err)
	}

	result, err := gateway.Charge(chargeInput, networkToken, paymentToken)
	if err != nil {
		t.Error(err)
	}

	t.Log(result)
}

func TestConnectCharge(t *testing.T) {
	gateway, err := New(context.Background(), &buyte.ProviderCheckoutConnection{
		IsTest:      true,
		Credentials: connectCredentials,
		Provider: buyte.ProviderCheckoutConnectionProviderDetails{
			Name: "Stripe",
		},
	})
	if err != nil {
		t.Error(err)
	}

	networkToken := &buyte.NetworkToken{}
	err = json.Unmarshal([]byte(networkTokenData), networkToken)
	if err != nil {
		t.Error(err)
	}

	result, err := gateway.Charge(chargeInput, networkToken, paymentToken)
	if err != nil {
		t.Error(err)
	}

	t.Log(result)
}
