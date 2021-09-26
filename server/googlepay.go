package server

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/store"
)

func (s *Server) ProcessGooglePayResponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &buyte.GooglePayAuthorizedPaymentResponse{}
		if err := render.DecodeJSON(r.Body, response); err != nil {
			_ = render.Render(w, r, s.ErrInvalidRequest(err))
			return
		}

		input := buyte.NewGooglePayPaymentTokenInput(response)
		paymentToken, err := s.store.CreatePaymentToken(r.Context(), input)
		if err != nil {
			if store.IsConnectionInvalid(err) {
				_ = render.Render(w, r, s.ErrInvalidRequest(err))
			} else {
				s.logger.Errorw("Create Payment Token from Google Pay", "Params", input)
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
