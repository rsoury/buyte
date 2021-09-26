package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/paymentgateway"
	"github.com/rsoury/buyte/pkg/user"
	"github.com/rsoury/buyte/store"
)

func (s *Server) CreateCharge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode input
		input := &buyte.CreateChargeInput{}
		if err := render.DecodeJSON(r.Body, input); err != nil {
			_ = render.Render(w, r, s.ErrInvalidRequest(err))
			return
		}

		// Validate input
		if input.Source == "" || input.Amount <= 0 {
			_ = render.Render(w, r, s.ErrInvalidRequestWithMessage(errors.New("Missing required parameters")))
			return
		}
		if input.Amount <= 50 {
			_ = render.Render(w, r, s.ErrInvalidRequestWithMessage(errors.New("Amount must be greater than 50 cents")))
			return
		}
		if input.Currency == "" {
			input.Currency = "aud"
		}

		// Get Payment Token Data
		paymentToken, err := s.store.GetPaymentToken(r.Context(), input.Source)
		if err != nil {
			s.logger.Warnw("Create Charge", "error", errors.Wrap(err, "Cannot get Payment Token"))
			_ = render.Render(w, r, s.ErrInvalidRequest(err))
			return
		}

		// Validate amount in input
		// In the future, you'd allow for partial payments... ie input.Amount <= paymentToken.Amount
		// TODO: Add a way to include meaningful error message for production...
		if input.Amount != paymentToken.Amount {
			_ = render.Render(w, r, s.ErrInvalidRequestWithMessage(errors.New("Amount does not equal amount in authorized payment.")))
			return
		}
		if !strings.EqualFold(input.Currency, paymentToken.Currency) {
			_ = render.Render(w, r, s.ErrInvalidRequestWithMessage(errors.New("Currency does not equal currency in authorized payment.")))
			return
		}

		s.logger.Infow("Create Charge", "token", paymentToken.ID, "message", "Passed validation")

		// Once validated, Get User
		u := user.FromContext(r.Context())

		// Create the charge params
		customer := paymentToken.Customer()
		params := &buyte.CreateChargeParams{
			Source:      paymentToken.ID,
			Amount:      input.Amount,
			Currency:    input.Currency,
			Captured:    true,
			Description: input.Description,
			Customer:    customer,
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
		err = params.SetMetadata(input.Metadata)
		if err != nil {
			_ = render.Render(w, r, s.ErrInternalServer(err))
			return
		}
		err = params.SetOrder(&input.Order)
		if err != nil {
			_ = render.Render(w, r, s.ErrInternalServer(err))
			return
		}

		s.logger.Infow("Create Charge", "token", paymentToken.ID, "message", "Charge params created.")
		s.logger.Debugw("Create Charge", "params", params)

		var networkToken *buyte.NetworkToken
		var nativeToken string
		if paymentToken.IsApplePay() {
			// Decrypt Apple Pay Token
			applePayPaymentToken, err := paymentToken.ApplePay()
			if err != nil {
				_ = render.Render(w, r, s.ErrInternalServer(err))
				return
			}
			applePayNetworkToken, err := s.applepay.DecryptResponse(applePayPaymentToken.Response)
			if err != nil {
				if paymentToken.IsRejectedSigningTimeDelta(err) {
					_ = render.Render(w, r, s.ErrRequestFailed(err))
					return
				} else {
					_ = render.Render(w, r, s.ErrInternalServer(err))
					return
				}
			}
			networkToken = &buyte.NetworkToken{
				applePayNetworkToken,
			}
		} else if paymentToken.IsGooglePay() {
			// Extract Native token
			googlePayPaymentToken, err := paymentToken.GooglePay()
			if err != nil {
				_ = render.Render(w, r, s.ErrInternalServer(err))
				return
			}
			nativeToken = googlePayPaymentToken.Response.PaymentMethodData.TokenizationData.Token
		} else {
			_ = render.Render(w, r, s.ErrInternalServer(errors.New("Payment Token type not valid")))
			return
		}

		s.logger.Infow("Create Charge", "message", "Network/Native token attained.")

		// Get Payment Provider from used Checkout
		// s.logger.Debugw("Create Charge", "Token", paymentToken)
		paymentProvider, err := paymentgateway.New(r.Context(), paymentToken.Checkout.Connection)
		if err != nil {
			_ = render.Render(w, r, s.ErrInternalServer(err))
			return
		}

		// Check if provider is connect or not. Connect is now the keyword for our locally used Payment Provider.
		if paymentProvider.Gateway.IsConnect() {
			// Set fee amount
			params.SetFee(u.UserAttributes.FeeMultiplier, u.UserAttributes.Country)
			input.SetFee(u.UserAttributes.FeeMultiplier, u.UserAttributes.Country)

			s.logger.Infow("Create Charge", "message", "Fee applied")
		}

		// Execute Charge on Payment Provider
		var result *buyte.GatewayCharge
		if nativeToken != "" {
			result, err = paymentProvider.Gateway.ChargeNative(input, nativeToken, paymentToken)
			if err != nil {
				_ = render.Render(w, r, s.ErrInternalServer(err))
				return
			}
			s.logger.Infow("Create Charge", "message", "Gateway charge executed successfully", "type", "native")
		} else {
			result, err = paymentProvider.Gateway.Charge(input, networkToken, paymentToken)
			if err != nil {
				_ = render.Render(w, r, s.ErrInternalServer(err))
				return
			}
			s.logger.Infow("Create Charge", "message", "Gateway charge executed successfully", "type", "network")
		}
		params.SetProviderCharge(result)

		// Return Charge
		charge, err := s.store.CreateCharge(r.Context(), params)
		if err != nil {
			s.logger.Errorw("Create Charge", "Params", params)
			_ = render.Render(w, r, s.ErrInternalServer(err))
			return
		}

		if paymentProvider.Gateway.IsConnect() {
			// Increment account balance.
			go func() {
				newBalance := charge.Amount - charge.FeeAmount
				err = u.IncrementAccountBalance(newBalance)
				if err != nil {
					s.logger.Errorw("Create Charge", "Incrementing Account Balance", err, "Charge", charge.ID, "New Balance", newBalance)
				} else {
					s.logger.Infow("Create Charge", "Incrementing Account Balance", "Success", "Charge", charge.ID, "New Balance", newBalance)
				}
			}()
		}

		s.logger.Infow("Create Charge", "Charge", charge.ID, "Gateway Charge", result.Reference, "Payment Token", paymentToken.ID)

		// Return Charge
		render.JSON(w, r, charge)
	}
}

func (s *Server) GetCharge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chargeId := chi.URLParam(r, "id")
		charge, err := s.store.GetCharge(r.Context(), chargeId)
		if err != nil {
			if store.IsConnectionUnauthorized(err) {
				_ = render.Render(w, r, ErrNotFound)
			} else {
				_ = render.Render(w, r, s.ErrInternalServer(err))
			}
			return
		}
		if charge.ID == "" {
			_ = render.Render(w, r, ErrNotFound)
			return
		}

		s.logger.Infow("Get Charge", "Charge", charge.ID)

		render.JSON(w, r, charge)
	}
}
