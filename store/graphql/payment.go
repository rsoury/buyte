package graphql

import (
	"context"

	"github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/user"
)

const paymentQLModel = `
	id
	value
	amount
	currency
	shippingMethod {
		id
		label
		description
		rate
	}
	selectedShippingMethod {
		reference
		label
		description
		rate
	}
	paymentMethod {
		name
	}
	checkout{
		id
		label
		description
		connection {
			type
			isTest
			credentials
			provider {
				name
			}
		}
	}
`

func (c *Client) CreatePaymentToken(ctx context.Context, paymentTokenInput *buyte.CreatePaymentTokenInput) (*buyte.PaymentToken, error) {
	u := user.FromContext(ctx)
	auth := u.AccessToken
	userAttributes := u.UserAttributes

	// Create token hash from the timestamp x MAC Address based uuid, set as ID for store.
	paymentTokenInput.ID = c.newID("tok")

	// Create payment token data store.
	req := graphql.NewRequest(`
		mutation CreatePaymentToken($input: CreatePaymentTokenInput!) {
			createPaymentToken(input: $input) {
				` + paymentQLModel + `
			}
		}
	`)

	if err := paymentTokenInput.Format(); err != nil {
		return &buyte.PaymentToken{}, err
	}

	req.Var("input", paymentTokenInput)
	req.Header.Set("Authorization", auth)

	var respData map[string]interface{}
	if err := c.Run(ctx, req, &respData); err != nil {
		return &buyte.PaymentToken{}, err
	}

	paymentToken := &buyte.PaymentToken{}
	if err := mapstructure.Decode(respData["createPaymentToken"], paymentToken); err != nil {
		return &buyte.PaymentToken{}, err
	}
	paymentToken.Object = buyte.PAYMENT_TOKEN

	// Clean output shipping method.
	// If no shipping module, set shipping method to whatever is retrieved by selectedShippingMethod
	if userAttributes.ShippingModule == 0 {
		paymentToken.ShippingMethod = buyte.CopySelectedShippingMethodToShippingMethod(paymentToken.SelectedShippingMethod)
	}
	paymentToken.SelectedShippingMethod = nil

	c.logger.Infow("Payment Token", "action", "create", "id", paymentTokenInput.ID)

	// Initialises with promoted BarePaymentToken, and an Empty Value.
	return paymentToken, nil
}

func (c *Client) GetPaymentToken(ctx context.Context, paymentTokenId string) (*buyte.PaymentToken, error) {
	u := user.FromContext(ctx)
	auth := u.AccessToken
	userAttributes := u.UserAttributes

	// Create request
	req := graphql.NewRequest(`
		query GetPaymentToken($id: ID!) {
			getPaymentToken(id: $id) {
				` + paymentQLModel + `
			}
		}
	`)
	req.Var("id", paymentTokenId)
	req.Header.Set("Authorization", auth)
	var respData map[string]interface{}
	if err := c.Run(ctx, req, &respData); err != nil {
		return &buyte.PaymentToken{}, err
	}

	paymentToken := &buyte.PaymentToken{}
	if err := mapstructure.Decode(respData["getPaymentToken"], paymentToken); err != nil {
		return &buyte.PaymentToken{}, err
	}

	paymentToken.Object = buyte.PAYMENT_TOKEN

	// Clean output shipping method.
	// If no shipping module, set shipping method to whatever is retrieved by selectedShippingMethod
	if userAttributes.ShippingModule == 0 {
		paymentToken.ShippingMethod = buyte.CopySelectedShippingMethodToShippingMethod(paymentToken.SelectedShippingMethod)
	}
	paymentToken.SelectedShippingMethod = nil

	c.logger.Infow("Payment Token", "action", "get", "id", paymentTokenId)

	return paymentToken, nil
}
