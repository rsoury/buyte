package graphql

import (
	"context"
	"errors"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"
	config "github.com/spf13/viper"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/user"
)

type (
	ShippingZone struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Countries []struct {
			Iso  string `json:"iso"`
			Name string `json:"name"`
		} `json:"countries"`
		PriceRates struct {
			Items []struct {
				ID            string      `json:"id"`
				Label         string      `json:"label"`
				Description   string      `json:"description"`
				MinOrderPrice int         `json:"minOrderPrice"`
				MaxOrderPrice interface{} `json:"maxOrderPrice"`
				Rate          int         `json:"rate"`
			}
		}
	}
	CheckoutData struct {
		ID         string `json:"id"`
		IsArchived bool   `json:"isArchived"`
		Connection struct {
			Type     string `json:"type"`
			Provider struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			Credentials string `json:"credentials"`
			IsTest      bool   `json:"isTest"`
		}
		PaymentOptions struct {
			Items []struct {
				PaymentOption struct {
					ID    string `json:"id"`
					Image string `json:"image"`
					Name  string `json:"name"`
				} `json:"paymentOption"`
			} `json:"items"`
		} `json:"paymentOptions"`
		ShippingZone *ShippingZone `json:"-"`
	}
)

// Using Main User Struct
// SUGGEST: Filter the shipping zones within the Graphql Query...
// --> Current Fix: Filtering in Go
func (c *Client) GetFullCheckout(ctx context.Context, checkoutId string, options *buyte.FullCheckoutOptions) (*buyte.FullCheckout, error) {
	u := user.FromContext(ctx)
	auth := u.AccessToken
	userAttributes := u.UserAttributes

	c.logger.Debugw("Graphql: Get Checkout", "User Attributes", userAttributes)

	shippingZoneQuery := ``
	if userAttributes.ShippingModule == 1 {
		shippingZoneQuery = `
			listShippingZones {
				items {
					id
					name
					countries {
						iso
						name
					}
					priceRates {
						items {
							id
							label
							description
							minOrderPrice
							maxOrderPrice
							rate
						}
					}
				}
			}
		`
	}
	req := graphql.NewRequest(`
		query GetFullCheckout($id: ID!) {
			getCheckout(id: $id) {
				id
				connection {
					type
					provider {
						id
						name
					}
					credentials
					isTest
				}
				paymentOptions {
					items {
						paymentOption{
							id
							name
							image
						}
					}
				}
				isArchived
			}
			` + shippingZoneQuery + `
		}
	`)
	req.Var("id", checkoutId)
	req.Header.Set("Authorization", auth)

	var respData map[string]interface{}
	if err := c.Run(ctx, req, &respData); err != nil {
		return &buyte.FullCheckout{}, err
	}
	c.logger.Debugw("Graphql: Get Checkout", "Response Data", respData)

	checkout := &CheckoutData{}
	if err := mapstructure.Decode(respData["getCheckout"], checkout); err != nil {
		return &buyte.FullCheckout{}, err
	}

	// if requesting an archived checkout, throw unauthorized.
	if checkout.IsArchived {
		return &buyte.FullCheckout{}, errors.New("graphql: Not Authorized")
	}

	// Configure shipping zones, filter by provided country code.
	filteredZone := &ShippingZone{}
	if userAttributes.ShippingModule == 1 {
		shippingZones := map[string][]ShippingZone{
			"items": []ShippingZone{},
		}
		if err := mapstructure.Decode(respData["listShippingZones"], &shippingZones); err != nil {
			return &buyte.FullCheckout{}, err
		}
		for _, zone := range shippingZones["items"] {
			for _, country := range zone.Countries {
				if strings.EqualFold(country.Iso, options.UserCountryCode) {
					filteredZone = &zone
					break
				}
			}
			if filteredZone.ID != "" {
				break
			}
		}
		c.logger.Debugw("Graphql: Get Checkout", "Shipping Zone", filteredZone)
	}
	checkout.ShippingZone = filteredZone

	// Format checkout options
	var checkoutOptions []buyte.FullCheckoutOptionResponse
	for _, item := range checkout.PaymentOptions.Items {
		option := buyte.FullCheckoutOptionResponse{
			ID:    item.PaymentOption.ID,
			Name:  item.PaymentOption.Name,
			Image: item.PaymentOption.Image,
		}
		switch option.Name {
		case buyte.APPLE_PAY:
			option.AdditionalData = map[string]string{
				"merchantId":   config.GetString("apple.merchant.id"),
				"merchantName": config.GetString("apple.merchant.name"),
			}
		case buyte.GOOGLE_PAY:
			option.AdditionalData = map[string]string{
				"merchantId":   config.GetString("google.merchant.id"),
				"merchantName": config.GetString("google.merchant.name"),
			}
		default:
		}
		checkoutOptions = append(checkoutOptions, option)
	}
	// Format Shipping Methods
	var shippingMethods []buyte.FullCheckoutShippingMethod
	for _, rate := range checkout.ShippingZone.PriceRates.Items {
		method := buyte.FullCheckoutShippingMethod{
			ID:          rate.ID,
			Name:        rate.Label,
			Description: rate.Description,
			Rate:        rate.Rate,
			MinOrder:    rate.MinOrderPrice,
			MaxOrder:    rate.MaxOrderPrice,
		}
		shippingMethods = append(shippingMethods, method)
	}

	var publicKeyBytes []byte
	switch checkout.Connection.Type {
	case buyte.STRIPE:
		isConnect, _ := jsonparser.GetBoolean([]byte(checkout.Connection.Credentials), "isConnect")
		if isConnect {
			if checkout.Connection.IsTest {
				publicKeyBytes = []byte(config.GetString("stripe.test.public"))
			} else {
				publicKeyBytes = []byte(config.GetString("stripe.live.public"))
			}
		} else {
			publicKeyBytes, _, _, _ = jsonparser.Get([]byte(checkout.Connection.Credentials), "stripePublishableKey")
		}
	case buyte.ADYEN:
		publicKeyBytes, _, _, _ = jsonparser.Get([]byte(checkout.Connection.Credentials), "merchantAccount")
	default:
	}
	gatewayProvider := buyte.FullCheckoutGatewayProvider{
		ID:        checkout.Connection.Provider.ID,
		Name:      checkout.Connection.Provider.Name,
		PublicKey: string(publicKeyBytes),
		IsTest:    checkout.Connection.IsTest,
	}

	return &buyte.FullCheckout{
		ID:              checkout.ID,
		Object:          buyte.FULL_CHECKOUT,
		GatewayProvider: gatewayProvider,
		Options:         checkoutOptions,
		ShippingMethods: shippingMethods,
		Currency:        userAttributes.Currency,
		Country:         userAttributes.Country,
		Merchant: buyte.FullCheckoutMerchant{
			StoreName:  userAttributes.StoreName,
			Website:    userAttributes.Website,
			Logo:       userAttributes.Logo,
			CoverImage: userAttributes.CoverImage,
		},
		CustomCSS: userAttributes.CustomCSS,
	}, nil
}
