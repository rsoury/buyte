package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rsoury/buyte/buyte"
)

func TestCreateCharge(t *testing.T) {
	ctx := Context()

	params := &buyte.CreateChargeParams{
		Source:      "tok_bicqcg4rtr3954id8tc0",
		Amount:      1824,
		Currency:    "aud",
		Captured:    true,
		Description: "This is within a unit test.",
	}
	err := params.SetMetadata(map[string]interface{}{
		"this":     "is",
		"metadata": 101,
	})
	if err != nil {
		t.Error(err)
	}
	err = params.SetOrder(&buyte.ChargeOrder{
		Reference: "some-uuid",
		Platform:  "Some Shopping CMS",
		Items: []map[string]interface{}{
			map[string]interface{}{
				"name":   "item1",
				"amount": 1000,
			},
			map[string]interface{}{
				"name":   "item2",
				"amount": 2000,
			},
		},
		Shipping: map[string]interface{}{
			"name": "Express Shipping",
		},
		Customer: map[string]interface{}{
			"name": "Bob Dillion",
		},
	})
	if err != nil {
		t.Error(err)
	}
	params.SetProviderCharge(&buyte.GatewayCharge{
		Reference: "some-charge-id",
		Type:      "STRIPE",
	})

	charge, err := New().CreateCharge(ctx, params)
	if err != nil {
		t.Error("Error with CreateCharge", err)
	}
	t.Log(charge)
}

func TestGetCharge(t *testing.T) {
	assert := assert.New(t)
	ctx := Context()
	chargeId := "ch_bih0i0crtr3chmg67ss0"
	charge, err := New().GetCharge(ctx, chargeId)
	if err != nil {
		t.Error("Error with GetCharge", err)
	}
	assert.Equal(chargeId, charge.ID, "The two Charge Ids should be the same.")
}
