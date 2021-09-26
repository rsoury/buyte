package adyen

import (
	"encoding/json"
	"testing"

	"github.com/rsoury/buyte/buyte"
	_ "github.com/rsoury/buyte/conf"
)

const credentials_test = `{
	"username": "ws_094450@Company.WebDoodle",
	"password": "YI\jRE<6wZG1#1~u#en[3+PGN",
	"merchantAccount": "WebDoodleAU",
	"csePublicKey": "10001|B7286CB7F7D867378A2530DD753D5FEE59FEFA86F965A7492E5876BF60F1EF0FBBD96E2573822A26F86060D67F8996B8DBF1D207AA5AAA5C6D4302A388DBE4C99C53D544434FBD53142DFD5E9A63D6D6D32DD6C5A4E6C429BA6D1D35D092C976FD134C4A7A9C53828C95534A69597F449C71DEB7074599C4A56E486B2F8DEBB83E1C5475B7F38FF600B9975FD8B3F5B899E6232F4CA9DBA2CC30C8A860DFA4A6709C57074213044AE2665ACEB37D4203B6D22F8DA62F461931B6787A75C825DD7A879B5F44C8BC67B3DFA4A0D181F0DE44A7575244256EEE5D5491851EF51CE85D935501F38E0E86596FD17BEE66385172FBC38F670D6F5EDB8AC4D403F488A5"
}`
const networkTokenData_test = `{
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

var (
	chargeInput = &buyte.CreateChargeInput{
		Amount:   1200,
		Currency: "aud",
		// Description: "This is a test",
		Metadata: map[string]interface{}{
			"Test": true,
		},
		Order: buyte.ChargeOrder{
			Reference: "some-order-id",
		},
	}
	paymentToken = &buyte.PaymentToken{
		PaymentMethod: buyte.PaymentMethod{
			Name: "Apple Pay",
		},
	}
)

func GatewaySetup() (*Gateway, error) {
	connection := &buyte.ProviderCheckoutConnection{
		IsTest:      true,
		Credentials: credentials_test,
		Provider: buyte.ProviderCheckoutConnectionProviderDetails{
			Name: "Adyen",
		},
	}
	gateway, err := New(connection)
	if err != nil {
		return &Gateway{}, err
	}
	return gateway, nil
}

func TestEncrypt(t *testing.T) {
	var err error
	gateway, err := GatewaySetup()
	if err != nil {
		t.Error(err)
	}

	networkToken := &buyte.NetworkToken{}
	err = json.Unmarshal([]byte(networkTokenData_test), networkToken)
	if err != nil {
		t.Error(err)
	}

	result, err := gateway.Encrypt(&CardEncryptParams{
		Number:     networkToken.ApplicationPrimaryAccountNumber,
		ExpMonth:   networkToken.ExpMonth(),
		ExpYear:    networkToken.ExpYear(),
		HolderName: "Hello Test!",
	})
	if err != nil {
		t.Error(err)
	}

	t.Log(result)
}

func TestPost(t *testing.T) {
	gateway, err := GatewaySetup()
	if err != nil {
		t.Error(err)
	}

	jsonStr := `{
		"hello": "world"
	}`
	result, err := gateway.Post("https://hookb.in/PxyeQX9QjZT0j0WBGRMb", []byte(jsonStr))
	if err != nil {
		t.Error(err)
	}

	t.Log(string(result))
}

func TestCharge(t *testing.T) {
	var err error
	gateway, err := GatewaySetup()
	if err != nil {
		t.Error(err)
	}

	networkToken := &buyte.NetworkToken{}
	err = json.Unmarshal([]byte(networkTokenData_test), networkToken)
	if err != nil {
		t.Error(err)
	}

	result, err := gateway.Charge(chargeInput, networkToken, paymentToken)
	if err != nil {
		t.Error(err)
	}

	t.Log(result)
}
