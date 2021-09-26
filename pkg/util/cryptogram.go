package util

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Cryptogram struct {
	Value []byte `json:"cryptogram"`
}

func DecodeCryptogram(cryptogram []byte) (string, error) {
	// Feels like a bad approach but it works...
	obj := Cryptogram{
		Value: cryptogram,
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return "", errors.Wrap(err, "Could not Decode Cryptogram")
	}
	var cryptogramJSON map[string]string
	err = json.Unmarshal(data, &cryptogramJSON)
	if err != nil {
		return "", errors.Wrap(err, "Could not Decode Cryptogram")
	}
	result := cryptogramJSON["cryptogram"]

	return result, nil
}
