package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/store"
)

func (s *Server) GetFullCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkoutWidgetId := chi.URLParam(r, "id")
		userCountryCode := r.URL.Query().Get("user_country_code")
		if userCountryCode == "" {
			userCountryCode = r.Header.Get("Cloudfront-Viewer-Country")
		}
		if userCountryCode == "" {
			_ = render.Render(w, r, s.ErrInvalidRequest(errors.New("User Country Code is a required query parameter eg. ?user_country_code=AU")))
			return
		}
		checkout, err := s.store.GetFullCheckout(r.Context(), checkoutWidgetId, &buyte.FullCheckoutOptions{
			UserCountryCode: userCountryCode,
		})
		if err != nil {
			if store.IsConnectionUnauthorized(err) {
				_ = render.Render(w, r, ErrNotFound)
			} else {
				_ = render.Render(w, r, s.ErrInternalServer(err))
			}
			return
		}
		if checkout.ID == "" {
			_ = render.Render(w, r, ErrNotFound)
			return
		}

		render.JSON(w, r, checkout)
	}
}
