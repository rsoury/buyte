package buyte

import (
	"context"

	"go.uber.org/zap"
)

// Payment Gateway Types
const (
	STRIPE = "STRIPE"
	ADYEN  = "ADYEN"
)

type Gateway struct {
	Type        string             `json:"type"`
	IsTest      bool               `json:"isTest"`
	Credentials interface{}        `json:"credentials"`
	Context     context.Context    `json:"-"`
	Logger      *zap.SugaredLogger `json:"-"`
}
