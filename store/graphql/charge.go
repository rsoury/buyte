package graphql

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/user"
)

const locationAddressQLModel = `
	addressLines
	administrativeArea
	country
	countryCode
	locality
	postalCode
	subAdministrativeArea
	subLocality
`
const chargeQLModel = `
	id
	source {
		id
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
		checkout {
			id
			label
			description
		}
	}
	amount
	feeAmount
	currency
	captured
	description
	metadata
	customer {
		name
		givenName
		familyName
		emailAddress
		phoneNumber
		shippingAddress {
			` + locationAddressQLModel + `
		}
		billingAddress {
			` + locationAddressQLModel + `
		}
	}
	order {
		reference
		platform
		items
		shipping
		customer
	}
	createdAt
`

func responseToChargeDecodeHook() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// String to Map JSON
		if f.Kind() == reflect.String && t.Kind() == reflect.Map {
			// Convert JSON string to map[string]interface{}
			var result map[string]interface{}
			err := json.Unmarshal([]byte(data.(string)), &result)
			if err != nil {
				return data, errors.Wrap(err, "mapstructure decodehookfunc: ")
			}
			return result, nil
		}

		// Order items.
		if f.Kind() == reflect.String && (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) {
			// Convert JSON string to []map[string]interface{}
			var result []map[string]interface{}
			err := json.Unmarshal([]byte(data.(string)), &result)
			if err != nil {
				return data, errors.Wrap(err, "mapstructure decodehookfunc: ")
			}
			return result, nil
		}

		return data, nil
	}
}

func (c *Client) CreateCharge(ctx context.Context, params *buyte.CreateChargeParams) (*buyte.Charge, error) {
	// Validate params
	if params.Currency == "" || params.Amount <= 0 || params.Source == "" {
		return &buyte.Charge{}, errors.New("Missing required parameters")
	}

	u := user.FromContext(ctx)
	auth := u.AccessToken
	userAttributes := u.UserAttributes

	// Create token hash from the timestamp x MAC Address based uuid, set as ID for store.
	params.ID = c.newID("ch")

	// Create request to store charge data
	req := graphql.NewRequest(`
		mutation CreateCharge($input: CreateChargeInput!) {
			createCharge(input: $input) {
				` + chargeQLModel + `
			}
		}
	`)

	req.Var("input", params)
	req.Header.Set("Authorization", auth)

	var respData map[string]interface{}
	if err := c.Run(ctx, req, &respData); err != nil {
		return &buyte.Charge{}, err
	}

	charge := &buyte.Charge{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       responseToChargeDecodeHook(),
		WeaklyTypedInput: true,
		Result:           charge,
	})
	if err != nil {
		return &buyte.Charge{}, err
	}
	if err := decoder.Decode(respData["createCharge"]); err != nil {
		return &buyte.Charge{}, err
	}

	charge.Object = buyte.CHARGE

	if userAttributes.ShippingModule == 0 {
		charge.Source.ShippingMethod = buyte.CopySelectedShippingMethodToShippingMethod(charge.Source.SelectedShippingMethod)
	}
	charge.Source.SelectedShippingMethod = nil

	c.logger.Infow("Charge", "action", "create", "id", params.ID)

	return charge, nil
}

func (c *Client) GetCharge(ctx context.Context, chargeId string) (*buyte.Charge, error) {
	u := user.FromContext(ctx)
	auth := u.AccessToken
	userAttributes := u.UserAttributes

	// Create request to store charge data
	req := graphql.NewRequest(`
		query GetCharge($id: ID!) {
			getCharge(id: $id) {
				` + chargeQLModel + `
			}
		}
	`)

	req.Var("id", chargeId)
	req.Header.Set("Authorization", auth)

	var respData map[string]interface{}
	if err := c.Run(ctx, req, &respData); err != nil {
		return &buyte.Charge{}, err
	}

	charge := &buyte.Charge{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       responseToChargeDecodeHook(),
		WeaklyTypedInput: true,
		Result:           charge,
	})
	if err != nil {
		return &buyte.Charge{}, err
	}
	if err := decoder.Decode(respData["getCharge"]); err != nil {
		return &buyte.Charge{}, err
	}

	charge.Object = buyte.CHARGE

	if userAttributes.ShippingModule == 0 {
		charge.Source.ShippingMethod = buyte.CopySelectedShippingMethodToShippingMethod(charge.Source.SelectedShippingMethod)
	}
	charge.Source.SelectedShippingMethod = nil

	c.logger.Infow("Charge", "action", "get", "id", charge.ID)

	return charge, nil
}
