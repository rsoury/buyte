package buyte

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/rsoury/buyte/pkg/googlepay"
	"github.com/rsoury/buyte/pkg/util"

	"github.com/pkg/errors"
	"github.com/rsoury/applepay"
)

const (
	APPLE_PAY  = "Apple Pay"
	GOOGLE_PAY = "Google Pay"
)

type PaymentTokenStore interface {
	CreatePaymentToken(context.Context, *CreatePaymentTokenInput) (*PaymentToken, error)
	GetPaymentToken(context.Context, string) (*PaymentToken, error)
}

type AuthorizedPaymentResponse struct {
	CheckoutId        string                                   `json:"checkoutId"`
	PaymentMethodId   string                                   `json:"paymentMethodId"`
	ShippingMethod    *AuthorizedPaymentResponseShippingMethod `json:"shippingMethod,omitempty"`
	Amount            int                                      `json:"amount"`
	Currency          string                                   `json:"currency"`
	Country           string                                   `json:"country"`
	RawPaymentRequest map[string]interface{}                   `json:"rawPaymentRequest,omitempty"`
}
type AuthorizedPaymentResponseShippingMethod struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Rate        int    `json:"rate"`
	MinOrder    int    `json:"minOrder"`
	MaxOrder    int    `json:"maxOrder,omitempty"`
}
type ApplePayAuthorizedPaymentResponse struct {
	AuthorizedPaymentResponse
	Result *applepay.Response `json:"result"`
}
type GooglePayAuthorizedPaymentResponse struct {
	AuthorizedPaymentResponse
	Result *googlepay.Response `json:"result"`
}

type NetworkToken struct {
	*applepay.Token
}
type ProviderCheckoutConnectionProviderDetails struct {
	Name string `json:"name"`
}
type ProviderCheckoutConnection struct {
	Type        string                                    `json:"type"`
	IsTest      bool                                      `json:"isTest"`
	Credentials string                                    `json:"credentials"`
	Provider    ProviderCheckoutConnectionProviderDetails `json:"provider"`
}
type PaymentMethod struct {
	Name string `json:"name"`
}
type PublicPaymentToken struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
type PaymentTokenShipping struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Rate        int    `json:"rate"`
	MinOrder    int    `json:"minOrderPrice,omitempty" mapstructure:"minOrderPrice"`
	MaxOrder    int    `json:"maxOrderPrice,omitempty" mapstructure:"maxOrderPrice"`
}
type PaymentTokenSelectedShipping struct {
	ID          string `json:"reference" mapstructure:"reference"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Rate        int    `json:"rate"`
	MinOrder    int    `json:"minOrderPrice,omitempty" mapstructure:"minOrderPrice"`
	MaxOrder    int    `json:"maxOrderPrice,omitempty" mapstructure:"maxOrderPrice"`
}
type PaymentTokenBaseCheckout struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}
type PaymentTokenCheckout struct {
	ID          string                      `json:"id"`
	Label       string                      `json:"label"`
	Description string                      `json:"description,omitempty"`
	Connection  *ProviderCheckoutConnection `json:"connection,omitempty"`
}
type PaymentToken struct {
	ID                     string                        `json:"id"`
	Object                 string                        `json:"object"`
	Value                  string                        `json:"value"`
	PaymentMethod          *PaymentMethod                `json:"paymentMethod"`
	Amount                 int                           `json:"amount"`
	Currency               string                        `json:"currency"`
	ShippingMethod         *PaymentTokenShipping         `json:"shippingMethod,omitempty"`
	SelectedShippingMethod *PaymentTokenSelectedShipping `json:"selectedShippingMethod,omitempty"`
	Checkout               *PaymentTokenCheckout         `json:"checkout"`
}
type ApplePayPaymentToken struct {
	*PaymentToken
	Response *applepay.Response `json:"response"`
}
type GooglePayPaymentToken struct {
	*PaymentToken
	Response *googlepay.Response `json:"response"`
}
type CreatePaymentTokenInput struct {
	ID                string                        `json:"id"`
	Value             interface{}                   `json:"value"`
	CheckoutId        string                        `json:"paymentTokenCheckoutId"`
	PaymentMethodId   string                        `json:"paymentTokenPaymentMethodId"`
	ShippingMethodId  string                        `json:"paymentTokenShippingMethodId,omitempty"`
	ShippingMethod    *PaymentTokenSelectedShipping `json:"selectedShippingMethod,omitempty"`
	Amount            int                           `json:"amount"`
	Currency          string                        `json:"currency"`
	Country           string                        `json:"country"`
	RawPaymentRequest interface{}                   `json:"rawPaymentRequest,omitempty"`
}

func (p *PaymentToken) IsApplePay() bool {
	return p.PaymentMethod.Name == APPLE_PAY
}

func (p *PaymentToken) IsGooglePay() bool {
	return p.PaymentMethod.Name == GOOGLE_PAY
}

func (p *PaymentToken) GooglePay() (*GooglePayPaymentToken, error) {
	if !p.IsGooglePay() {
		return &GooglePayPaymentToken{}, errors.New("PaymentMethod not Google Pay")
	}
	var googlePayResponse googlepay.Response
	err := json.Unmarshal([]byte(p.Value), &googlePayResponse)
	if err != nil {
		return &GooglePayPaymentToken{}, errors.Wrap(err, "Could not format PaymentToken to GooglePayPaymentToken")
	}
	return &GooglePayPaymentToken{
		p,
		&googlePayResponse,
	}, nil
}

func (p *PaymentToken) ApplePay() (*ApplePayPaymentToken, error) {
	if !p.IsApplePay() {
		return &ApplePayPaymentToken{}, errors.New("PaymentMethod not Apple Pay")
	}
	var applePayResponse applepay.Response
	err := json.Unmarshal([]byte(p.Value), &applePayResponse)
	if err != nil {
		return &ApplePayPaymentToken{}, errors.Wrap(err, "Could not format PaymentToken to ApplePayPaymentToken")
	}
	return &ApplePayPaymentToken{
		p,
		&applePayResponse,
	}, nil
}

func (p *PaymentToken) Format() (interface{}, error) {
	if p.PaymentMethod.Name == "" {
		return p, errors.New("PaymentMethod not present")
	}
	switch p.PaymentMethod.Name {
	case APPLE_PAY:
		return p.ApplePay()
	case GOOGLE_PAY:
		return p.GooglePay()
	default:
	}
	return p, nil
}

func (p *PaymentToken) Customer() *Customer {
	if p.IsApplePay() {
		applePayPaymentToken, err := p.ApplePay()
		if err != nil {
			return &Customer{}
		}
		var applePayContact applepay.Contact
		if applePayPaymentToken.Response.ShippingContact.EmailAddress != "" {
			applePayContact = applePayPaymentToken.Response.ShippingContact
		} else if applePayPaymentToken.Response.BillingContact.EmailAddress != "" {
			applePayContact = applePayPaymentToken.Response.BillingContact
		}
		customer := &Customer{
			Name:         strings.TrimSpace(applePayContact.GivenName + " " + applePayContact.FamilyName),
			GivenName:    applePayContact.GivenName,
			FamilyName:   applePayContact.FamilyName,
			EmailAddress: applePayContact.EmailAddress,
			PhoneNumber:  applePayContact.PhoneNumber,
		}
		customer.SetShippingAddress(&CustomerAddress{
			AddressLines:          applePayPaymentToken.Response.ShippingContact.AddressLines,
			AdministrativeArea:    applePayPaymentToken.Response.ShippingContact.AdministrativeArea,
			Country:               applePayPaymentToken.Response.ShippingContact.Country,
			CountryCode:           applePayPaymentToken.Response.ShippingContact.CountryCode,
			Locality:              applePayPaymentToken.Response.ShippingContact.Locality,
			PostalCode:            applePayPaymentToken.Response.ShippingContact.PostalCode,
			SubAdministrativeArea: applePayPaymentToken.Response.ShippingContact.SubAdministrativeArea,
			SubLocality:           applePayPaymentToken.Response.ShippingContact.SubLocality,
		})
		customer.SetBillingAddress(&CustomerAddress{
			AddressLines:          applePayPaymentToken.Response.BillingContact.AddressLines,
			AdministrativeArea:    applePayPaymentToken.Response.BillingContact.AdministrativeArea,
			Country:               applePayPaymentToken.Response.BillingContact.Country,
			CountryCode:           applePayPaymentToken.Response.BillingContact.CountryCode,
			Locality:              applePayPaymentToken.Response.BillingContact.Locality,
			PostalCode:            applePayPaymentToken.Response.BillingContact.PostalCode,
			SubAdministrativeArea: applePayPaymentToken.Response.BillingContact.SubAdministrativeArea,
			SubLocality:           applePayPaymentToken.Response.BillingContact.SubLocality,
		})
		return customer
	} else if p.IsGooglePay() {
		googlePayPaymentToken, err := p.GooglePay()
		if err != nil {
			return &Customer{}
		}
		shippingAddress := googlePayPaymentToken.Response.ShippingAddress
		var billingAddress googlepay.Address
		if googlePayPaymentToken.Response.BillingAddress.PhoneNumber != "" {
			billingAddress = googlePayPaymentToken.Response.BillingAddress
		} else if googlePayPaymentToken.Response.PaymentMethodData.Info.BillingAddress.PhoneNumber != "" {
			billingAddress = googlePayPaymentToken.Response.PaymentMethodData.Info.BillingAddress
		}
		var name string
		if shippingAddress.Name != "" {
			name = shippingAddress.Name
		} else if billingAddress.Name != "" {
			name = billingAddress.Name
		}
		var phoneNumber string
		if shippingAddress.PhoneNumber != "" {
			phoneNumber = shippingAddress.PhoneNumber
		} else if billingAddress.PhoneNumber != "" {
			phoneNumber = billingAddress.PhoneNumber
		}
		givenName, familyName := util.Namesplit(name)
		customer := &Customer{
			Name:         name,
			GivenName:    givenName,
			FamilyName:   familyName,
			EmailAddress: googlePayPaymentToken.Response.Email,
			PhoneNumber:  phoneNumber,
		}
		customer.SetShippingAddress(&CustomerAddress{
			AddressLines: []string{
				shippingAddress.Address1,
				shippingAddress.Address2,
				shippingAddress.Address3,
			},
			AdministrativeArea: shippingAddress.AdministrativeArea,
			CountryCode:        shippingAddress.CountryCode,
			Locality:           shippingAddress.Locality,
			PostalCode:         shippingAddress.PostalCode,
		})
		customer.SetBillingAddress(&CustomerAddress{
			AddressLines: []string{
				billingAddress.Address1,
				billingAddress.Address2,
				billingAddress.Address3,
			},
			AdministrativeArea: billingAddress.AdministrativeArea,
			CountryCode:        billingAddress.CountryCode,
			Locality:           billingAddress.Locality,
			PostalCode:         billingAddress.PostalCode,
		})
		return customer
	}
	return &Customer{}
}

func (p *PaymentToken) IsRejectedSigningTimeDelta(err error) bool {
	return strings.Contains(err.Error(), "invalid token signature: rejected signing time delta")
}

// Copy selected shipping data to ShippingMethod.
func CopySelectedShippingMethodToShippingMethod(selected *PaymentTokenSelectedShipping) *PaymentTokenShipping {
	if selected != nil {
		return &PaymentTokenShipping{
			ID:          selected.ID,
			Label:       selected.Label,
			Description: selected.Description,
			Rate:        selected.Rate,
			MinOrder:    selected.MinOrder,
			MaxOrder:    selected.MaxOrder,
		}
	}

	return nil
}

func NewPaymentTokenInput(response *AuthorizedPaymentResponse) *CreatePaymentTokenInput {
	tokenInput := &CreatePaymentTokenInput{
		CheckoutId:        response.CheckoutId,
		PaymentMethodId:   response.PaymentMethodId,
		Amount:            response.Amount,
		Currency:          response.Currency,
		Country:           response.Country,
		RawPaymentRequest: response.RawPaymentRequest,
	}

	// If memory address for shipping is not nil
	if response.ShippingMethod != nil {
		tokenInput.ShippingMethodId = response.ShippingMethod.ID
		tokenInput.ShippingMethod = &PaymentTokenSelectedShipping{
			ID:          response.ShippingMethod.ID,
			Label:       response.ShippingMethod.Label,
			Description: response.ShippingMethod.Description,
			Rate:        response.ShippingMethod.Rate,
			MinOrder:    response.ShippingMethod.MinOrder,
			MaxOrder:    response.ShippingMethod.MaxOrder,
		}
	}

	return tokenInput
}

// NewApplePayPaymentTokenInput Sets result of Authed Payment Response as the value
func NewApplePayPaymentTokenInput(response *ApplePayAuthorizedPaymentResponse) *CreatePaymentTokenInput {
	input := NewPaymentTokenInput(&response.AuthorizedPaymentResponse)
	input.Value = response.Result
	return input
}

// NewApplePayPaymentTokenInput Sets result of Authed Payment Response as the value
func NewGooglePayPaymentTokenInput(response *GooglePayAuthorizedPaymentResponse) *CreatePaymentTokenInput {
	input := NewPaymentTokenInput(&response.AuthorizedPaymentResponse)
	input.Value = response.Result
	return input
}

// Format Payment Token Input
func (i *CreatePaymentTokenInput) Format() error {
	// Ensure input Value is formatted appropriately -- in JSON.
	inputValue, err := EnsureJSON(i.Value)
	if err != nil {
		return errors.Wrap(err, "Could not ensure JSON for Value")
	}
	i.Value = inputValue

	// Ensure input RawPaymentRequest is formatted appropriately -- in JSON.
	inputRawPaymentRequest, err := EnsureJSON(i.RawPaymentRequest)
	if err != nil {
		return errors.Wrap(err, "Could not ensure JSON for RawPaymentRequest")
	}
	i.RawPaymentRequest = inputRawPaymentRequest

	// Keep currency lowercase
	i.Currency = strings.ToLower(i.Currency)

	return nil
}

// Network Token applicationExpirationDate is in YYMMDD format
// ApplicationExpirationDate https://www.emvlab.org/emvtags/show/t5F24/
func (n *NetworkToken) ExpMonth() string {
	if len(n.ApplicationExpirationDate) == 6 {
		chars := []rune(n.ApplicationExpirationDate)
		result := string(chars[2])
		result += string(chars[3])
		return result
	}
	return ""
}
func (n *NetworkToken) ExpYear() string {
	if len(n.ApplicationExpirationDate) == 6 {
		chars := []rune(n.ApplicationExpirationDate)
		result := string(chars[0])
		result += string(chars[1])
		return "20" + result
	}
	return ""
}
