package buyte

import "context"

type CheckoutStore interface {
	GetFullCheckout(context.Context, string, *FullCheckoutOptions) (*FullCheckout, error)
}

// Public Load Full Checkout Widget Response
type FullCheckoutOptionResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Image          string            `json:"image"`
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}
type FullCheckoutShippingMethod struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Rate        int         `json:"rate"`
	MinOrder    int         `json:"minOrder"`
	MaxOrder    interface{} `json:"maxOrder,omitempty"`
}
type FullCheckoutMerchant struct {
	StoreName  string `json:"storeName"`
	Website    string `json:"website"`
	Logo       string `json:"logo"`
	CoverImage string `json:"coverImage"`
}
type FullCheckoutGatewayProvider struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	IsTest    bool   `json:"isTest"`
}
type FullCheckout struct {
	ID              string                       `json:"id"`
	Object          string                       `json:"object"`
	Options         []FullCheckoutOptionResponse `json:"options"`
	ShippingMethods []FullCheckoutShippingMethod `json:"shippingMethods"`
	GatewayProvider FullCheckoutGatewayProvider  `json:"gatewayProvider"`
	Currency        string                       `json:"currency"`
	Country         string                       `json:"country"`
	Merchant        FullCheckoutMerchant         `json:"merchant"`
	CustomCSS       string                       `json:"customCss"`
}

type FullCheckoutOptions struct {
	UserCountryCode string
}
