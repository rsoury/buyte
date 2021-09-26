package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/palantir/stacktrace"
	"github.com/pkg/errors"
	config "github.com/spf13/viper"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/source"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/util"
)

type Gateway buyte.Gateway
type StripeCredentials struct {
	UserId      string `json:"stripeUserId"`
	IsConnect   bool   `json:"isConnect"`
	AccessToken string `json:"accessToken"`
}

func New(ctx context.Context, connection *buyte.ProviderCheckoutConnection) (*Gateway, error) {
	credentials := &StripeCredentials{}
	if err := json.Unmarshal([]byte(connection.Credentials), credentials); err != nil {
		return &Gateway{}, err
	}
	return &Gateway{
		Type:        connection.Type,
		IsTest:      connection.IsTest,
		Credentials: credentials,
		Context:     ctx,
		Logger:      zap.S().With("package", "paymentgateway.stripe"),
	}, nil
}

func (g *Gateway) StripeCredentials() *StripeCredentials {
	return g.Credentials.(*StripeCredentials)
}

func (g *Gateway) AuthKey() string {
	// Is connect? Use platform credentials, otherwise use masked oauth generated key
	credentials := g.StripeCredentials()
	if credentials.IsConnect {
		if g.IsTest {
			return config.GetString("stripe.test.secret")
		}
		return config.GetString("stripe.live.secret")
	}
	return credentials.AccessToken
}

func (g *Gateway) IsConnect() bool {
	credentials := g.StripeCredentials()
	return credentials.IsConnect
}

func (g *Gateway) Charge(input *buyte.CreateChargeInput, networkToken *buyte.NetworkToken, paymentToken *buyte.PaymentToken) (*buyte.GatewayCharge, error) {
	// Create source
	cryptogram, err := util.DecodeCryptogram(networkToken.PaymentData.OnlinePaymentCryptogram)

	if err != nil {
		return &buyte.GatewayCharge{}, errors.Wrap(err, "Could not deduce payment cryptogram")
	}

	sourceData := map[string]string{
		"number":              networkToken.ApplicationPrimaryAccountNumber,
		"exp_month":           networkToken.ExpMonth(),
		"exp_year":            networkToken.ExpYear(),
		"cryptogram":          cryptogram,
		"tokenization_method": strings.Replace(strings.ToLower(paymentToken.PaymentMethod.Name), " ", "_", -1),
	}
	if networkToken.PaymentData.ECIIndicator != "" {
		sourceData["eci"] = util.Rjust(networkToken.PaymentData.ECIIndicator, 2, "0")
	}

	stripe.Key = g.AuthKey()
	sourceParams := &stripe.SourceObjectParams{
		Type:     stripe.String("card"),
		TypeData: sourceData,
		Currency: stripe.String(input.Currency),
	}
	src, err := source.New(sourceParams)
	if err != nil {
		return &buyte.GatewayCharge{}, stacktrace.Propagate(err, "Could not create stripe source")
	}
	chargeParams := g.createChargeParams(input, paymentToken)
	return g.executeCharge(chargeParams, src.ID)
}

func (g *Gateway) ChargeNative(input *buyte.CreateChargeInput, nativeToken string, paymentToken *buyte.PaymentToken) (*buyte.GatewayCharge, error) {
	stripe.Key = g.AuthKey()
	tokenId, err := jsonparser.GetString([]byte(nativeToken), "id")
	if err != nil {
		return &buyte.GatewayCharge{}, errors.Wrap(err, "Could not charge stripe token")
	}
	chargeParams := g.createChargeParams(input, paymentToken)
	return g.executeCharge(chargeParams, tokenId)
}

func (g *Gateway) createChargeParams(input *buyte.CreateChargeInput, paymentToken *buyte.PaymentToken) *stripe.ChargeParams {
	description := input.Description
	if description == "" {
		description = "Buyte: " + paymentToken.PaymentMethod.Name
		if input.Order.Reference != "" {
			description = description + " - " + input.Order.Reference
		}
	}
	// Create charge
	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(int64(input.Amount)),
		Capture:     stripe.Bool(true), // For now...
		Currency:    stripe.String(input.Currency),
		Description: stripe.String(description),
	}

	// If connect, set on behalf of stripe user id only if user id is set too.
	if g.IsConnect() {
		credentials := g.StripeCredentials()
		if credentials.UserId != "" {
			chargeParams.Destination = &stripe.DestinationParams{
				Account: stripe.String(credentials.UserId),
			}
			chargeParams.ApplicationFeeAmount = stripe.Int64(int64(input.FeeAmount))
		}
	}

	for key, value := range input.Metadata {
		chargeParams.AddMetadata(key, fmt.Sprintf("%v", value))
	}
	return chargeParams
}

func (g *Gateway) executeCharge(chargeParams *stripe.ChargeParams, token string) (*buyte.GatewayCharge, error) {
	err := chargeParams.SetSource(token)
	if err != nil {
		return &buyte.GatewayCharge{}, errors.Wrap(err, "Could not create stripe charge")
	}
	ch, err := charge.New(chargeParams)
	if err != nil {
		return &buyte.GatewayCharge{}, errors.Wrap(err, "Could not create stripe charge")
	}

	g.Logger.Infow("Stripe Charge", "source_id", token, "charge_id", ch.ID)

	// Return Charge
	return &buyte.GatewayCharge{
		Reference: ch.ID,
		Type:      g.Type,
	}, nil
}
