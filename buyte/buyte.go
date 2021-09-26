package buyte

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Create a overarching store that inherits all methods of sub store interfaces
type Store interface {
	CheckoutStore
	PaymentTokenStore
	ChargeStore
}

// Some Util
func EnsureJSON(value interface{}) (string, error) {
	var err error
	var input []byte
	switch val := value.(type) {
	case []byte:
		input = val
	case string:
		input = []byte(val)
	default:
		input, err = json.Marshal(val)
		if err != nil {
			return "", errors.Wrap(err, "Could not marshal input")
		}
	}
	if !json.Valid(input) {
		return "", errors.New("Input Value not valid JSON.")
	}
	return string(input), nil
}
