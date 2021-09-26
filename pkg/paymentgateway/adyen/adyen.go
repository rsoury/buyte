package adyen

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/buger/jsonparser"
	"github.com/palantir/stacktrace"
	"github.com/pkg/errors"
	config "github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/util"
)

type Gateway buyte.Gateway

type CardEncryptParams struct {
	Number     string `json:"number"`
	ExpMonth   string `json:"expiryMonth"`
	ExpYear    string `json:"expiryYear"`
	Cvc        string `json:"cvc,omitempty"`
	HolderName string `json:"holderName"`
}

type AdyenAuthoriseParams struct {
	Reference       string                             `json:"reference"`
	MerchantAccount string                             `json:"merchantAccount"`
	Amount          AdyenAmountParams                  `json:"amount"`
	AdditionalData  AdyenAuthoriseAdditionalDataParams `json:"additionalData"`
	MpiData         AdyenAuthoriseMpiDataParams        `json:"mpiData"`
}
type AdyenAuthoriseAdditionalDataParams struct {
	Card               string `json:"card.encrypted.json"`
	Type               string `json:"paymentdatasource.type"`
	SelectedBrand      string `json:"selectedBrand"`
	ShopperInteraction string `json:"shopperInteraction"`
}
type AdyenAmountParams struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}
type AdyenAuthoriseMpiDataParams struct {
	AuthenticationResponse string `json:"authenticationResponse"`
	DirectoryResponse      string `json:"directoryResponse"`
	Cavv                   string `json:"cavv"`
	Eci                    string `json:"eci"`
}
type AdyenCaptureParams struct {
	Reference          string            `json:"reference"`
	MerchantAccount    string            `json:"merchantAccount"`
	ModificationAmount AdyenAmountParams `json:"modificationAmount"`
	OriginalReference  string            `json:"originalReference"`
}
type AdyenGooglePayParams struct {
	Reference       string                            `json:"reference"`
	MerchantAccount string                            `json:"merchantAccount"`
	Amount          AdyenAmountParams                 `json:"amount"`
	PaymentMethod   AdyenGooglePayPaymentMethodParams `json:"paymentMethod"`
}
type AdyenGooglePayPaymentMethodParams struct {
	Type  string `json:"type"`
	Token string `json:"paywithgoogle.token"`
}

var CardTypeSource = map[string]string{
	"Apple Pay":  "applepay",
	"Google Pay": "paywithgoogle",
}

type AdyenCredentials struct {
	Username        []byte `json:"username"`
	Password        []byte `json:"password"`
	MerchantAccount string `json:"merchantAccount"`
	CsePublicKey    string `json:"csePublicKey"`
	LiveUrlPrefix   string `json:"liveUrlPrefix"`
}

func (a *AdyenCredentials) AuthKey() string {
	raw := string(a.Username) + ":" + string(a.Password)
	auth := []byte(raw)
	return base64.StdEncoding.EncodeToString(auth)
}

func New(ctx context.Context, connection *buyte.ProviderCheckoutConnection) (*Gateway, error) {
	var err error
	credentials := &AdyenCredentials{}
	creds := []byte(connection.Credentials)
	err = jsonparser.ObjectEach(creds, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		keyStr := string(key)
		switch keyStr {
		case "username":
			credentials.Username, _ = jsonparser.Unescape(value, []byte(""))
		case "password":
			credentials.Password, _ = jsonparser.Unescape(value, []byte(""))
		case "merchantAccount":
			credentials.MerchantAccount = string(value)
		case "csePublicKey":
			credentials.CsePublicKey = string(value)
		case "liveUrlPrefix":
			credentials.LiveUrlPrefix = string(value)
		}
		return nil
	})
	if err != nil {
		return &Gateway{}, err
	}

	return &Gateway{
		Type:        connection.Type,
		IsTest:      connection.IsTest,
		Credentials: credentials,
		Context:     ctx,
		Logger:      zap.S().With("package", "paymentgateway.adyen"),
	}, nil
}

func (g *Gateway) AdyenCredentials() *AdyenCredentials {
	return g.Credentials.(*AdyenCredentials)
}

// For now.
func (g *Gateway) IsConnect() bool {
	return false
}

func (g *Gateway) Charge(input *buyte.CreateChargeInput, networkToken *buyte.NetworkToken, paymentToken *buyte.PaymentToken) (*buyte.GatewayCharge, error) {
	// Get encrypted data
	name := networkToken.CardholderName
	if name == "" {
		name = "Not Provided"
	}
	cseCard, err := g.Encrypt(
		&CardEncryptParams{
			Number:     networkToken.ApplicationPrimaryAccountNumber,
			ExpMonth:   networkToken.ExpMonth(),
			ExpYear:    networkToken.ExpYear(),
			HolderName: name,
		},
	)
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not encrypt payment data")
	}

	// Build Authorise request
	cryptogram, err := util.DecodeCryptogram(networkToken.PaymentData.OnlinePaymentCryptogram)
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not build authorisation request")
	}
	description := g.getDescription(input, paymentToken)
	eci := "07"
	if networkToken.PaymentData.ECIIndicator != "" {
		eci = util.Rjust(networkToken.PaymentData.ECIIndicator, 2, "0")
	}
	authParams := &AdyenAuthoriseParams{
		Reference:       description,
		MerchantAccount: g.AdyenCredentials().MerchantAccount,
		Amount: AdyenAmountParams{
			Value:    input.Amount,
			Currency: strings.ToUpper(input.Currency),
		},
		AdditionalData: AdyenAuthoriseAdditionalDataParams{
			Card:               cseCard,
			Type:               CardTypeSource[paymentToken.PaymentMethod.Name],
			SelectedBrand:      CardTypeSource[paymentToken.PaymentMethod.Name],
			ShopperInteraction: "Ecommerce",
		},
		MpiData: AdyenAuthoriseMpiDataParams{
			AuthenticationResponse: "Y",
			DirectoryResponse:      "Y",
			Eci:                    eci,
			Cavv:                   cryptogram,
		},
	}

	// Execute authorisation
	authoriseResponse, err := g.authorise(authParams)
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not execute authorisation request")
	}

	g.Logger.Infow("Authorise", "response", authoriseResponse)

	// Get PSP
	pspBytes, _, _, err := jsonparser.Get(authoriseResponse, "pspReference")
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not obtain PSP")
	}
	psp := string(pspBytes)

	// Build Capture Request
	captureParams := &AdyenCaptureParams{
		Reference:       description,
		MerchantAccount: g.AdyenCredentials().MerchantAccount,
		ModificationAmount: AdyenAmountParams{
			Value:    input.Amount,
			Currency: strings.ToUpper(input.Currency),
		},
		OriginalReference: psp,
	}

	// Execute request
	captureResponse, err := g.capture(captureParams)
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not execute capture request")
	}
	g.Logger.Infow("Capture", "response", captureResponse)

	// Return Charge
	return &buyte.GatewayCharge{
		Reference: psp,
		Type:      g.Type,
	}, nil
}

func (g *Gateway) ChargeNative(input *buyte.CreateChargeInput, nativeToken string, paymentToken *buyte.PaymentToken) (*buyte.GatewayCharge, error) {
	if paymentToken.PaymentMethod.Name == buyte.GOOGLE_PAY {
		description := g.getDescription(input, paymentToken)
		params := &AdyenGooglePayParams{
			Reference:       description,
			MerchantAccount: g.AdyenCredentials().MerchantAccount,
			Amount: AdyenAmountParams{
				Value:    input.Amount,
				Currency: strings.ToUpper(input.Currency),
			},
			PaymentMethod: AdyenGooglePayPaymentMethodParams{
				Type:  "paywithgoogle",
				Token: nativeToken,
			},
		}

		response, err := g.googlepay(params)
		if err != nil {
			return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not execute Google Pay payment request")
		}

		g.Logger.Infow("Google Pay Payment", "response", response)

		// Get PSP
		pspBytes, _, _, err := jsonparser.Get(response, "pspReference")
		if err != nil {
			return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not obtain PSP")
		}
		psp := string(pspBytes)

		// Return Charge
		return &buyte.GatewayCharge{
			Reference: psp,
			Type:      g.Type,
		}, nil
	}

	return &buyte.GatewayCharge{}, nil
}

func (g *Gateway) getDescription(input *buyte.CreateChargeInput, paymentToken *buyte.PaymentToken) string {
	description := input.Description
	if description == "" {
		description = "Buyte: " + paymentToken.PaymentMethod.Name
		if input.Order.Reference != "" {
			description = description + " - " + input.Order.Reference
		}
	}
	return description
}

func (g *Gateway) googlepay(params *AdyenGooglePayParams) ([]byte, error) {
	// Get Endpoint
	url := g.checkoutEndpoint() + "/payments"

	// params to json
	jsonData, err := json.Marshal(params)
	if err != nil {
		return []byte{}, err
	}
	return g.Post(url, jsonData)
}

func (g *Gateway) capture(params *AdyenCaptureParams) ([]byte, error) {
	// Get Endpoint
	url := g.paymentsEndpoint() + "/capture"

	// params to json
	jsonData, err := json.Marshal(params)
	if err != nil {
		return []byte{}, err
	}
	return g.Post(url, jsonData)
}

// Returns response body as byte array
func (g *Gateway) authorise(params *AdyenAuthoriseParams) ([]byte, error) {
	// Get Endpoint
	url := g.paymentsEndpoint() + "/authorise"
	// url := "https://hookb.in/G9Qy1gMmL8s1m1eBNyQ7"

	// params to json
	jsonData, err := json.Marshal(params)
	if err != nil {
		return []byte{}, err
	}
	return g.Post(url, jsonData)
}

func (g *Gateway) Post(url string, jsonBody []byte) ([]byte, error) {
	// Build Request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Authorization", "Basic "+g.AdyenCredentials().AuthKey())
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := g.client()
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode == 401 {
		return []byte{}, errors.New("Unauthorized")
	}

	// Return response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func (g *Gateway) client() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
	}
}

func (g *Gateway) paymentsEndpoint() string {
	if g.IsTest {
		return "https://pal-test.adyen.com/pal/servlet/Payment/v40"
	}
	prefix := g.AdyenCredentials().LiveUrlPrefix
	if prefix == "" {
		return fmt.Sprintf("https://%s-pal-live.adyenpayments.com/pal/servlet/Payment/v40", prefix)
	}
	return "https://pal-live.adyen.com/pal/servlet/Payment/v40"
}
func (g *Gateway) checkoutEndpoint() string {
	if g.IsTest {
		return "https://checkout-test.adyen.com/services/PaymentSetupAndVerification/v46"
	}
	prefix := g.AdyenCredentials().LiveUrlPrefix
	if prefix == "" {
		return fmt.Sprintf("https://%s-checkout-live.adyenpayments.com/checkout/services/PaymentSetupAndVerification/v46", prefix)
	}
	return "https://checkout-live.adyen.com/services/PaymentSetupAndVerification/v46"
}

func (g *Gateway) Encrypt(params *CardEncryptParams) (string, error) {
	// CseKey     string `json:"cseKey"`
	csek := g.AdyenCredentials().CsePublicKey

	// Get Lambda Client
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(config.GetString("func.region"))},
	)
	client := lambda.New(sess)

	// Construct Payload
	payload, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	payload, err = jsonparser.Set(payload, []byte(`"`+csek+`"`), "cseKey")
	if err != nil {
		return "", err
	}

	result, err := client.Invoke(&lambda.InvokeInput{FunctionName: aws.String(config.GetString("func.adyen_cse")), Payload: payload})
	if err != nil {
		return "", err
	}

	cseOutputBytes, _, _, err := jsonparser.Get(result.Payload, "value")
	if err != nil {
		return "", err
	}

	cseOutput := string(cseOutputBytes)

	return cseOutput, nil
}
