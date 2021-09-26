package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/store"
)

// ApplePayResponse - ...
type ApplePaySessionResponse struct {
	EpochTimestamp            int64  `json:"epochTimestamp"`
	ExpiresAt                 int64  `json:"expiresAt"`
	MerchantSessionIdentifier string `json:"merchantSessionIdentifier"`
	Nonce                     string `json:"nonce"`
	MerchantIdentifier        string `json:"merchantIdentifier"`
	DomainName                string `json:"domainName"`
	DisplayName               string `json:"displayName"`
	Signature                 string `json:"signature"`
}

func (s *Server) GetApplePaySession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appleData := &struct {
			URL string `json:"url"`
		}{}
		if err := render.DecodeJSON(r.Body, appleData); err != nil {
			_ = render.Render(w, r, s.ErrInvalidRequest(err))
			return
		}

		payload, err := s.applepay.Session(appleData.URL)
		if err != nil {
			_ = render.Render(w, r, s.ErrInternalServer(err))
			return
		}

		// s.logger.Debugw("Apple Session", "url", appleData.URL, "session", payload)

		// Payload is a JSON string in byte array format
		var response ApplePaySessionResponse
		err = json.Unmarshal(payload, &response)
		if err != nil {
			_ = render.Render(w, r, s.ErrRequestFailed(err))
			return
		}
		render.JSON(w, r, response)
	}
}

func (s *Server) ProcessApplePayResponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &buyte.ApplePayAuthorizedPaymentResponse{}
		if err := render.DecodeJSON(r.Body, response); err != nil {
			_ = render.Render(w, r, s.ErrInvalidRequest(err))
			return
		}

		input := buyte.NewApplePayPaymentTokenInput(response)
		paymentToken, err := s.store.CreatePaymentToken(r.Context(), input)
		if err != nil {
			if store.IsConnectionInvalid(err) {
				_ = render.Render(w, r, s.ErrInvalidRequest(err))
			} else {
				s.logger.Errorw("Create Payment Token from Apple Pay", "Params", input)
				_ = render.Render(w, r, s.ErrInternalServer(err))
			}
			return
		}

		// We now have the payment data.
		render.JSON(w, r, &buyte.PublicPaymentToken{
			ID:       paymentToken.ID,
			Object:   paymentToken.Object,
			Amount:   paymentToken.Amount,
			Currency: paymentToken.Currency,
		})
	}
}
