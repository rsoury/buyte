package paymentgateway

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/paymentgateway/adyen"
	"github.com/rsoury/buyte/pkg/paymentgateway/stripe"
)

type ProviderDetails struct {
	Name string `json:"name"`
}
type GatewayProvider interface {
	// New(*buyte.ProviderCheckoutConnection) (*buyte.Gateway, error) Not interfacing for now...
	Charge(*buyte.CreateChargeInput, *buyte.NetworkToken, *buyte.PaymentToken) (*buyte.GatewayCharge, error)
	ChargeNative(*buyte.CreateChargeInput, string, *buyte.PaymentToken) (*buyte.GatewayCharge, error)
	IsConnect() bool
}

// Each provider has their own underling gateway provider details.
type Provider struct {
	IsTest  bool            `json:"isTest"`
	Details ProviderDetails `json:"provider"`
	Gateway GatewayProvider
}
type ProviderCharge map[string]interface{} // struct {}

func New(ctx context.Context, connection *buyte.ProviderCheckoutConnection) (*Provider, error) {
	var gatewayProvider GatewayProvider
	switch connection.Type {
	case buyte.STRIPE:
		gateway, err := stripe.New(ctx, connection)
		if err != nil {
			return &Provider{}, errors.Wrap(err, "Could not setup Stripe Gateway")
		}
		gatewayProvider = gateway
	case buyte.ADYEN:
		gateway, err := adyen.New(ctx, connection)
		if err != nil {
			return &Provider{}, errors.Wrap(err, "Could not setup Adyen Gateway")
		}
		gatewayProvider = gateway
	default:
		return &Provider{}, errors.New("Payment Provider " + connection.Provider.Name + " is not supported")
	}
	return &Provider{
		IsTest: connection.IsTest,
		Details: ProviderDetails{
			Name: connection.Provider.Name,
		},
		Gateway: gatewayProvider,
	}, nil
}
