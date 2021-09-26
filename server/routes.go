package server

import (
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/render"

	"github.com/go-chi/chi"

	"github.com/rsoury/buyte/conf"
	"github.com/rsoury/buyte/pkg/util"
)

// SetupRoutes configures all the routes for this service
func (s *Server) SetupRoutes() {
	// Version: 1.2.3 -- Major.Minor.Maintenance
	// Get version, get major version number.
	versionSplit := strings.Split(conf.Version, ".")
	var major string
	if len(versionSplit) > 0 {
		major = versionSplit[0]
	} else {
		major = "0"
	}

	s.logger.Debugw("Instantiate routes", "version", "/v"+major)

	// Health checks
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.NoContent(w, r)
	})
	s.router.Route("/v"+major, func(r chi.Router) {
		r.Post("/charges", s.CreateCharge())
		r.Get("/charges/{id}", s.GetCharge())
		// r.Post("/charges/{id}")
		// r.Post("/charges/{id}/capture")

		r.Get("/token/{id}", s.GetPaymentToken())

		// Wrap all routes accessable using the Public Key with a /public route.
		r.Route("/public", func(r chi.Router) {
			// Once it passes the authroizer which basically asks if it is a public key and if so, are you hitting a public endpoint, we need to obtain the public key and the checkout_id and then try to get the checkout details for the given user's checkout.
			r.Route("/checkout", func(r chi.Router) {
				r.Get("/{id}", s.GetFullCheckout())
			})
			r.Route("/applepay", func(r chi.Router) {
				r.Post("/session", s.GetApplePaySession())
				r.Post("/process", s.ProcessApplePayResponse())
			})
			r.Route("/googlepay", func(r chi.Router) {
				r.Post("/process", s.ProcessGooglePayResponse())
			})
		})
	})
}

// Routes specifically for development
func (s *Server) SetupDevRoutes() {
	// Setup Apple Pay Paths
	root := http.Dir(path.Join(util.DirName(), "../examples/applepay"))
	prefix := "/dev/applepay"
	fs := http.FileServer(root)
	sFs := http.StripPrefix(prefix, fs)
	s.router.Route("/.well-known", func(r chi.Router) {
		r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		}))
	})
	s.router.Route(prefix, func(r chi.Router) {
		r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sFs.ServeHTTP(w, r)
		}))
	})
}
