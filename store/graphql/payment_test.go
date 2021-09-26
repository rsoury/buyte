package graphql

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rsoury/buyte/pkg/googlepay"

	"github.com/rsoury/buyte/buyte"
)

func TestCreatePaymentToken(t *testing.T) {
	ctx := Context()
	// user := ctx.Value("user").(*User)
	// t.Log(user)
	token, err := New().CreatePaymentToken(ctx, &buyte.CreatePaymentTokenInput{
		Value: map[string]interface{}{
			"hello": "world",
		},
		CheckoutId:       "a1aeb05b-3f02-4dd5-b9a9-941377fb9c15",
		PaymentMethodId:  "11cfbaf1-8094-498d-af51-c38d7cf4bc0d",
		ShippingMethodId: "18c13641-c8a4-4e5d-a237-c6a7803477b3",
		Currency:         "AUD",
		Country:          "AU",
		Amount:           1,
		RawPaymentRequest: map[string]interface{}{
			"rawpaymentrequest": "somestring",
		},
	})
	if err != nil {
		t.Error("Error with CreatePaymentToken", err)
	}
	t.Log(token)
}

func TestCreatePaymentTokenWithGooglePayResponse(t *testing.T) {
	valueJSON := `
	{
		"apiVersionMinor": 0,
		"apiVersion": 2,
		"paymentMethodData": {
		  "description": "Mastercard •••• 9416",
		  "tokenizationData": {
			"type": "PAYMENT_GATEWAY",
			"token": "{\n  \"id\": \"tok_1EfQrtITpNO3y6fqbxaBcdjw\",\n  \"object\": \"token\",\n  \"card\": {\n    \"id\": \"card_1EfQrtITpNO3y6fqEEA9m7EA\",\n    \"object\": \"card\",\n    \"address_city\": null,\n    \"address_country\": null,\n    \"address_line1\": null,\n    \"address_line1_check\": null,\n    \"address_line2\": null,\n    \"address_state\": null,\n    \"address_zip\": null,\n    \"address_zip_check\": null,\n    \"brand\": \"MasterCard\",\n    \"country\": \"US\",\n    \"cvc_check\": null,\n    \"dynamic_last4\": \"4242\",\n    \"exp_month\": 6,\n    \"exp_year\": 2020,\n    \"funding\": \"credit\",\n    \"last4\": \"9416\",\n    \"metadata\": {\n    },\n    \"name\": \"Ryan Soury\",\n    \"tokenization_method\": \"android_pay\"\n  },\n  \"client_ip\": \"74.125.113.98\",\n  \"created\": 1559132817,\n  \"livemode\": false,\n  \"type\": \"card\",\n  \"used\": false\n}\n"
		  },
		  "type": "CARD",
		  "Info": {
			"cardNetwork": "MASTERCARD",
			"cardDetails": "9416"
		  }
		},
		"shippingOptionData": {
		  "id": "9e5caf12-feeb-4694-9578-34d9171b0b2d"
		},
		"shippingAddress": {
		  "phoneNumber": "+61 405 227 363",
		  "address3": "",
		  "sortingCode": "",
		  "address2": "Lalor Park",
		  "countryCode": "AU",
		  "address1": "27 Morton Rd",
		  "postalCode": "2147",
		  "name": "Ryan Soury",
		  "locality": "Sydney",
		  "administrativeArea": "NSW"
		},
		"billingAddress": {
		  "phoneNumber": "",
		  "address3": "",
		  "sortingCode": "",
		  "address2": "",
		  "countryCode": "",
		  "address1": "",
		  "postalCode": "",
		  "name": "",
		  "locality": "",
		  "administrativeArea": ""
		},
		"email": "rsoury318@gmail.com"
	  }
	`
	rawPaymentRequestJSON := `
	{
		"allowedPaymentMethods": [
		  {
			"parameters": {
			  "allowedAuthMethods": [
				"PAN_ONLY",
				"CRYPTOGRAM_3DS"
			  ],
			  "allowedCardNetworks": [
				"AMEX",
				"DISCOVER",
				"JCB",
				"MASTERCARD",
				"VISA"
			  ]
			},
			"tokenizationSpecification": {
			  "parameters": {
				"gateway": "stripe",
				"stripe:publishableKey": "pk_test_uLptUrNCbh4hZuJQN9RSlmEO",
				"stripe:version": "2018-10-31"
			  },
			  "type": "PAYMENT_GATEWAY"
			},
			"type": "CARD"
		  }
		],
		"apiVersion": 2,
		"apiVersionMinor": 0,
		"callbackIntents": [
		  "SHIPPING_ADDRESS",
		  "SHIPPING_OPTION"
		],
		"emailRequired": true,
		"environment": "TEST",
		"i": {
		  "googleTransactionId": "E6D63D50-ECE9-462E-AA4C-7E7F85BA663E.TEST",
		  "startTimeMs": 1559132810544
		},
		"merchantInfo": {
		  "merchantId": "",
		  "merchantName": "Test Store Name"
		},
		"shippingAddressParameters": {
		  "phoneNumberRequired": true
		},
		"shippingAddressRequired": true,
		"shippingOptionParameters": {
		  "defaultSelectedOptionId": "9e5caf12-feeb-4694-9578-34d9171b0b2d",
		  "shippingOptions": [
			{
			  "description": "Delivery in a week",
			  "id": "9e5caf12-feeb-4694-9578-34d9171b0b2d",
			  "label": "$9.90: Standard Shipping"
			},
			{
			  "description": "Delivers overnight",
			  "id": "0cff5d8c-d3fa-4a02-a7ff-234394e24907",
			  "label": "$29.00: Express Delivery"
			}
		  ]
		},
		"transactionInfo": {
		  "currencyCode": "AUD",
		  "displayItems": [
			{
			  "label": "5x This is my cool product!",
			  "price": "20.00",
			  "type": "LINE_ITEM"
			},
			{
			  "label": "Standard Shipping",
			  "price": "9.90",
			  "type": "LINE_ITEM"
			}
		  ],
		  "totalPrice": "29.90",
		  "totalPriceLabel": "Total",
		  "totalPriceStatus": "FINAL"
		}
	  }
	`
	value := &googlepay.Response{}
	var rawPaymentRequest map[string]interface{}
	_ = json.Unmarshal([]byte(valueJSON), value)
	_ = json.Unmarshal([]byte(rawPaymentRequestJSON), &rawPaymentRequest)

	response := buyte.AuthorizedPaymentResponse{
		CheckoutId:      "a1aeb05b-3f02-4dd5-b9a9-941377fb9c15",
		PaymentMethodId: "5126415b-71ec-40d6-94a7-5d3bc6e3f37f",
		ShippingMethod: &buyte.AuthorizedPaymentResponseShippingMethod{
			ID:          "standard-shipping",
			Label:       "Standard Shipping",
			Description: "Delivery in a week",
			Rate:        990,
		},
		Currency:          "AUD",
		Country:           "AU",
		Amount:            2990,
		RawPaymentRequest: rawPaymentRequest,
	}

	input := buyte.NewGooglePayPaymentTokenInput(&buyte.GooglePayAuthorizedPaymentResponse{
		response,
		value,
	})

	ctx := Context()
	// user := ctx.Value("user").(*User)
	// t.Log(user)
	token, err := New().CreatePaymentToken(ctx, input)
	if err != nil {
		t.Error("Error with CreatePaymentToken", err)
	}
	t.Log(token)
}

func TestGetPaymentToken(t *testing.T) {
	assert := assert.New(t)
	ctx := Context()
	tokenId := "tok_bkdm65srtr38h3shle80"
	token, err := New().GetPaymentToken(ctx, tokenId)
	if err != nil {
		t.Error("Error with GetPaymentToken", err)
	}
	t.Log(token.ID)
	assert.Equal(tokenId, token.ID, "Token retrieved!")
}
