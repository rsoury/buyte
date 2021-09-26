package buyte

import (
	"context"
	"math"
	"os"
	"strconv"
)

type ChargeStore interface {
	CreateCharge(context.Context, *CreateChargeParams) (*Charge, error)
	GetCharge(context.Context, string) (*Charge, error)
}

// Represents the Request Body data sent in POST /charges request.
type CreateChargeInput struct {
	Source      string                 `json:"source"`
	Amount      int                    `json:"amount"`
	FeeAmount   int                    `json:"feeAmount"`
	Currency    string                 `json:"currency"`
	Capture     *bool                  `json:"capture"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Order       ChargeOrder            `json:"order"`
}
type ChargeOrder struct {
	Reference string                   `json:"reference,omitempty"`
	Platform  string                   `json:"platform,omitempty"`
	Items     []map[string]interface{} `json:"items,omitempty"`
	Shipping  map[string]interface{}   `json:"shipping,omitempty"`
	Customer  map[string]interface{}   `json:"customer,omitempty"`
}
type ChargeSource struct {
	ID                     string                        `json:"id"`
	PaymentMethod          *PaymentMethod                `json:"paymentMethod"`
	ShippingMethod         *PaymentTokenShipping         `json:"shippingMethod,omitempty"`
	SelectedShippingMethod *PaymentTokenSelectedShipping `json:"selectedShippingMethod,omitempty"`
	Checkout               *PaymentTokenBaseCheckout     `json:"checkout"`
}
type Charge struct {
	ID             string                 `json:"id"`
	Object         string                 `json:"object"`
	Source         *ChargeSource          `json:"source"`
	Amount         int                    `json:"amount"`
	FeeAmount      int                    `json:"feeAmount"`
	Currency       string                 `json:"currency"`
	Captured       bool                   `json:"captured"`
	ProviderCharge *GatewayCharge         `json:"providerCharge,omitempty"`
	Description    string                 `json:"description"`
	Customer       *Customer              `json:"customer"`
	Metadata       map[string]interface{} `json:"metadata"`
	Order          *ChargeOrder           `json:"order,omitempty"`
	CreatedAt      string                 `json:"createdAt"`
}
type GatewayCharge struct {
	Reference string `json:"reference"`
	Type      string `json:"type"`
}

// Represent request body to GraphQL API to create a charge
type CreateChargeParams struct {
	ID             string                   `json:"id"`
	Source         string                   `json:"chargeSourceId"` // This is the payment token id.
	Amount         int                      `json:"amount"`
	FeeAmount      int                      `json:"feeAmount"`
	Currency       string                   `json:"currency"`
	Captured       bool                     `json:"captured"`
	Description    string                   `json:"description,omitempty"`
	Metadata       string                   `json:"metadata,omitempty"`
	ProviderCharge *GatewayCharge           `json:"providerCharge"`
	Customer       *Customer                `json:"customer"`
	Order          *CreateChargeOrderParams `json:"order,omitempty"`
	CreatedAt      string                   `json:"createdAt"`
}
type CreateChargeOrderParams struct {
	Reference string `json:"reference,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Items     string `json:"items,omitempty"`
	Shipping  string `json:"shipping,omitempty"`
	Customer  string `json:"customer,omitempty"`
}

func (c *CreateChargeInput) SetFee(feeMultiplier float64, region string) {
	c.FeeAmount = fee(c.Amount, feeMultiplier, region)
}

func (c *CreateChargeParams) SetFee(feeMultiplier float64, region string) {
	c.FeeAmount = fee(c.Amount, feeMultiplier, region)
}

func fee(baseAmount int, feeMultiplier float64, region string) int {
	// Either use assigned fee multiplier or derive from user region
	if feeMultiplier == 0 {
		feeMultiplierFromEnv := os.Getenv("TRANSACTION_FEE_MULTIPLIER")
		regionalFeeMultiplierFromEnv := os.Getenv("TRANSACTION_FEE_MULTIPLIER_" + region)
		if regionalFeeMultiplierFromEnv != "" {
			val, err := strconv.ParseFloat(regionalFeeMultiplierFromEnv, 64)
			if err == nil {
				feeMultiplier = val
			}
		} else if feeMultiplierFromEnv != "" {
			val, err := strconv.ParseFloat(feeMultiplierFromEnv, 64)
			if err == nil {
				feeMultiplier = val
			}
		}
	}

	// Apply fee multiplier
	fee := int(math.Round(feeMultiplier * float64(baseAmount)))

	// Set minimum fee
	regionalMinimumFee := os.Getenv("MINIMUM_TRANSACTION_FEE_" + region)
	defaultMinimumFee := os.Getenv("MINIMUM_TRANSACTION_FEE")
	if regionalMinimumFee != "" {
		val, err := strconv.Atoi(regionalMinimumFee)
		if err == nil && fee < val {
			fee = val
		}
	} else if defaultMinimumFee != "" {
		val, err := strconv.Atoi(defaultMinimumFee)
		if err == nil && fee < val {
			fee = val
		}
	}

	return fee
}

func (c *CreateChargeParams) SetMetadata(data interface{}) error {
	str, err := EnsureJSON(data)
	if err != nil {
		return err
	}
	c.Metadata = str
	return nil
}

func (c *CreateChargeParams) SetProviderCharge(gc *GatewayCharge) {
	c.ProviderCharge = gc
}
func (c *CreateChargeParams) SetOrder(co *ChargeOrder) error {
	params := &CreateChargeOrderParams{
		Reference: co.Reference,
		Platform:  co.Platform,
	}
	if len(co.Items) > 0 {
		err := params.SetItems(co.Items)
		if err != nil {
			return err
		}
	}
	if len(co.Shipping) > 0 {
		err := params.SetShipping(co.Shipping)
		if err != nil {
			return err
		}
	}
	if len(co.Customer) > 0 {
		err := params.SetCustomer(co.Customer)
		if err != nil {
			return err
		}
	}
	c.Order = params
	return nil
}

func (c *CreateChargeOrderParams) SetItems(data interface{}) error {
	str, err := EnsureJSON(data)
	if err != nil {
		return err
	}
	c.Items = str
	return nil
}
func (c *CreateChargeOrderParams) SetShipping(data interface{}) error {
	str, err := EnsureJSON(data)
	if err != nil {
		return err
	}
	c.Shipping = str
	return nil
}
func (c *CreateChargeOrderParams) SetCustomer(data interface{}) error {
	str, err := EnsureJSON(data)
	if err != nil {
		return err
	}
	c.Customer = str
	return nil
}

func (co *ChargeOrder) AddItem(item map[string]interface{}) {
	co.Items = append(co.Items, item)
}
