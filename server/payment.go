package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rsoury/buyte/store"
)

func (s *Server) GetPaymentToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paymentTokenId := chi.URLParam(r, "id")
		paymentToken, err := s.store.GetPaymentToken(r.Context(), paymentTokenId)
		if err != nil {
			if store.IsConnectionUnauthorized(err) {
				_ = render.Render(w, r, ErrNotFound)
			} else {
				_ = render.Render(w, r, s.ErrInternalServer(err))
			}
			return
		}
		if paymentToken.ID == "" {
			_ = render.Render(w, r, ErrNotFound)
			return
		}

		// formattedToken, err := paymentToken.Format()
		// if err != nil {
		// 	_ = render.Render(w, r, s.ErrInternalServer(err))
		// 	return
		// }

		render.JSON(w, r, paymentToken)
	}
}
