package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/gobwas/glob"
	"github.com/snowzach/certtools"
	"github.com/snowzach/certtools/autocert"
	config "github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rsoury/applepay"
	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/conf"
	"github.com/rsoury/buyte/pkg/user"
	"github.com/rsoury/buyte/pkg/util"
	"github.com/rsoury/buyte/test"
)

// Server is the API web server
type Server struct {
	logger   *zap.SugaredLogger
	router   chi.Router
	server   *http.Server
	store    buyte.Store
	applepay *applepay.Merchant
}

var (
	devRequestGlob    = glob.MustCompile("/{dev,.well-known,favicon}*")
	publicRequestGlob = glob.MustCompile("/v*/public/**")
)

func isDevRequest(uri string) bool {
	return devRequestGlob.Match(uri)
}
func isPublicRequest(uri string) bool {
	return publicRequestGlob.Match(uri)
}

// apiStrictMiddleware exists because their is a is a /dev endpoint for development.
// This wraps middleware and only applies the middleware strictly for API requests
func apiStrictMiddleware(middleware func(next http.Handler) http.Handler) func(next http.Handler) http.Handler {
	if config.GetBool("server.production") {
		return middleware
	} else {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !isDevRequest(r.RequestURI) || r.RequestURI == "/" || r.RequestURI == "" {
					middleware(next).ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		}
	}
}

// New will setup the API listener
func New(store buyte.Store) (*Server, error) {
	r := chi.NewRouter()
	// Set Standard Middleware
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.RequestID,
		middleware.RealIP,
		middleware.StripSlashes,
		middleware.Recoverer,
	)
	// Timeouts
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Config
	// Why not set cors inside route?
	// Because it's an OPTION method, we want to accept Authorizaton, before UserCtx or any Authentication Middleware is triggered.
	r.Use(func(next http.Handler) http.Handler {
		Handler := cors.New(cors.Options{
			// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Referer", "User-Agent", "X-Amz-Date", "X-Api-Key", "X-Amz-Security-Token"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
			Debug:            config.GetBool("server.log_cors"),
		}).Handler(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If public address, use cors handler, otherwiser, move on.
			if isPublicRequest(r.RequestURI) {
				Handler.ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
	// For Development Purposes
	if config.GetBool("server.mock.authorizer") {
		zap.L().Info("Mocking Authorizers for Development")
		r.Use(apiStrictMiddleware(test.NewMock().APIGatewayHeaders))
	}
	// Setup Middleware for User Request Context
	r.Use(apiStrictMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := user.Setup(r.Header.Get)
			// If user is empty, throw a 401. This may occur on internal requests to Load Balancer.
			if err != nil {
				_ = render.Render(w, r, ErrRequestUnauthorized(err))
				return
			}
			ctx := u.WithContext(r.Context())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}))

	// Log Requests
	if config.GetBool("server.log_requests") {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Prevent health checks and nocontent responeses from logging
				if r.RequestURI != "" && r.RequestURI != "/" {
					start := time.Now()
					var requestID string
					if reqID := r.Context().Value(middleware.RequestIDKey); reqID != nil {
						requestID = reqID.(string)
					}
					ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
					next.ServeHTTP(ww, r)

					latency := time.Since(start)

					fields := []zapcore.Field{
						zap.Int("status", ww.Status()),
						zap.Duration("took", latency),
						zap.String("remote", r.RemoteAddr),
						zap.String("request", r.RequestURI),
						zap.String("method", r.Method),
						zap.String("referrer", r.Referer()),
						zap.String("user", user.FromContext(r.Context()).ID),
						zap.String("package", "server.request"),
					}
					if requestID != "" {
						fields = append(fields, zap.String("request-id", requestID))
					}
					clientName := r.Header.Get("Client-Name")
					if clientName != "" {
						fields = append(fields, zap.String("client-name", clientName))
					}
					clientVersion := r.Header.Get("Client-Version")
					if clientVersion != "" {
						fields = append(fields, zap.String("client-version", clientVersion))
					}
					zap.L().Info("API Request", fields...)
				}
			})
		})
	}

	// Setup Apple Pay
	certRoot := config.GetString("apple.certs")
	if certRoot == "" {
		certRoot = path.Join(util.DirName(), "/../")
	}
	applepay.UnsafeSignatureVerification = true
	ap, err := applepay.New(
		config.GetString("apple.merchant.id"),
		applepay.MerchantDisplayName(config.GetString("apple.merchant.name")),
		applepay.MerchantDomainName(config.GetString("apple.merchant.domain")),
		applepay.MerchantCertificateLocation(
			path.Join(certRoot, "/certs/cert-merchant.crt"),
			path.Join(certRoot, "/certs/cert-merchant-key.pem"),
		),
		applepay.ProcessingCertificateLocation(
			path.Join(certRoot, "/certs/cert-processing.crt"),
			path.Join(certRoot, "/certs/cert-processing-key.pem"),
		),
	)
	if err != nil {
		zap.L().Warn("Cannot find Apple Pay Certificates. Running API without Apple Pay authority.", zap.Error(err))
	}

	s := &Server{
		logger:   zap.S().With("package", "server"),
		router:   r,
		store:    store,
		applepay: ap,
	}
	s.SetupRoutes()

	// Keep all "server.production" options in one place.
	if !config.GetBool("server.production") {
		s.SetupDevRoutes()
	}

	return s, nil

}

// ListenAndServe will listen for requests
func (s *Server) ListenAndServe() error {

	s.server = &http.Server{
		Addr:    net.JoinHostPort(config.GetString("server.host"), config.GetString("server.port")),
		Handler: s.router,
	}

	// Listen
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("Could not listen on %s: %v", s.server.Addr, err)
	}

	// Enable TLS?
	if config.GetBool("server.tls") {
		var cert tls.Certificate
		if config.GetBool("server.devcert") {
			s.logger.Warn("WARNING: This server is using an insecure development tls certificate. This is for development only!!!")
			cert, err = autocert.New(autocert.InsecureStringReader("localhost"))
			if err != nil {
				return fmt.Errorf("Could not autocert generate server certificate: %v", err)
			}
		} else {
			// Load keys from file
			cert, err = tls.LoadX509KeyPair(config.GetString("server.certfile"), config.GetString("server.keyfile"))
			if err != nil {
				return fmt.Errorf("Could not load server certificate: %v", err)
			}
		}

		// Enabed Certs - TODO Add/Get a cert
		s.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   certtools.SecureTLSMinVersion(),
			CipherSuites: certtools.SecureTLSCipherSuites(),
		}
		// Wrap the listener in a TLS Listener
		listener = tls.NewListener(listener, s.server.TLSConfig)
	}

	// Setup Sentry.
	if config.GetString("server.sentry") != "" {
		if err := raven.SetDSN(config.GetString("server.sentry")); err != nil {
			s.logger.Errorw("API Sentry Setup error", "error", err, "address", s.server.Addr)
		}
	}

	go func() {
		if err = s.server.Serve(listener); err != nil {
			s.logger.Fatalw("API Listen error", "error", err, "address", s.server.Addr)
		}
	}()
	s.logger.Infow("API Listening",
		"address", s.server.Addr,
		"tls", config.GetBool("server.tls"),
		"production", config.GetBool("server.production"),
		"sentry", config.GetString("server.sentry") != "",
		"version", conf.Version,
	)

	// Enable profiler
	if config.GetBool("server.profiler_enabled") && config.GetString("server.profiler_path") != "" {
		zap.S().Debugw("Profiler enabled on API", "path", config.GetString("server.profiler_path"))
		s.router.Mount(config.GetString("server.profiler_path"), middleware.Profiler())
	}

	return nil
}

func (s *Server) InitialiseServer() http.Handler {
	s.server = &http.Server{
		Addr:    net.JoinHostPort(config.GetString("server.host"), config.GetString("server.port")),
		Handler: s.router,
	}

	// Setup Sentry.
	if config.GetString("server.sentry") != "" {
		if err := raven.SetDSN(config.GetString("server.sentry")); err != nil {
			s.logger.Errorw("API Sentry Setup error", "error", err, "address", s.server.Addr)
		}
	}

	s.logger.Infow("Started Mux",
		"address", s.server.Addr,
		"production", config.GetBool("server.production"),
		"sentry", config.GetString("server.sentry") != "",
		"version", conf.Version,
	)

	// Enable profiler
	if config.GetBool("server.profiler_enabled") && config.GetString("server.profiler_path") != "" {
		zap.S().Debugw("Profiler enabled on API", "path", config.GetString("server.profiler_path"))
		s.router.Mount(config.GetString("server.profiler_path"), middleware.Profiler())
	}

	return s.router
}
