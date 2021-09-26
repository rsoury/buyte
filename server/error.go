package server

import (
	"net/http"

	"github.com/getsentry/raven-go"
	"github.com/go-chi/render"
	config "github.com/spf13/viper"
)

// ErrResponse is a generic struct for returning a standard error document
type ErrResponse struct {
	Err        error `json:"-"` // low-level runtime error
	StatusCode int   `json:"-"` // http response status code

	Message   string `json:"message"`         // user-level status message
	ErrorCode int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText string `json:"error,omitempty"` // application-level error message, for debugging
}

// ErrNotFound is a pre-built not-found error
var ErrNotFound = &ErrResponse{StatusCode: 404, Message: "Resource not found."}

// Render is the Renderer for ErrResponse struct
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Check if errortext is empty
	// TODO: Create some user-friendly error text dictionary...
	if e.Err != nil && e.ErrorText == "" {
		if !config.GetBool("server.production") {
			e.ErrorText = e.Err.Error()
		}
	}
	render.Status(r, e.StatusCode)
	return nil
}

// (*Server) ErrInvalidRequestWithMessage will log an error (as a debug log) and return an invalid request error to the user with a message.
func (s *Server) ErrInvalidRequestWithMessage(err error) render.Renderer {
	s.logger.Debugw("Invalid Request", "error", err)
	return ErrInvalidRequestWithMessage(err)
}

// ErrInvalidRequestWithMessage will log an error (as a debug log) and return an invalid request error to the user with a message.
func ErrInvalidRequestWithMessage(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: 400,
		Message:    "Invalid request: " + err.Error() + ".",
	}
}

// (*Server) ErrInvalidRequest will log an error (as a debug log) and return an invalid request error to the user
func (s *Server) ErrInvalidRequest(err error) render.Renderer {
	s.logger.Debugw("Invalid Request", "error", err)
	return ErrInvalidRequest(err)
}

// ErrInvalidRequest is used to indicate an error on user input (with wrapped error)
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: 400,
		Message:    "Invalid request.",
	}
}

// (*Server) ErrRequestFailed will log an error (as a debug log) and return an request failed error to the user
func (s *Server) ErrRequestFailed(err error) render.Renderer {
	s.logger.Debugw("Request Failed", "error", err)
	return ErrRequestFailed(err)
}

// ErrRequestFailed is used to indicate that user input was valid, but incorrect, hence an error
func ErrRequestFailed(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: 402,
		Message:    "Request failed.",
	}
}

// (*Server) ErrRequestUnauthorized will log an error (as a debug log) and return an request failed error to the user
func (s *Server) ErrRequestUnauthorized(err error) render.Renderer {
	s.logger.Debugw("Request unauthorized.", "error", err)
	return ErrRequestUnauthorized(err)
}

// ErrRequestUnauthorized is used to indicate that user input was valid, but incorrect, hence an error
func ErrRequestUnauthorized(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: 401,
		Message:    "Request unauthorized.",
	}
}

// (*Server) ErrRender will log an error (as a debug log) and return an render error to the user
func (s *Server) ErrRender(err error) render.Renderer {
	s.logger.Debugw("Render Error", "error", err)
	return ErrRender(err)
}

// ErrRender is used to indicate that there was an error rendering the response
func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: 422,
		Message:    "Error rendering response.",
	}
}

// (*Server) ErrInternalServer will log an error and return a generic server error to the user
func (s *Server) ErrInternalServer(err error) render.Renderer {
	s.logger.Errorw("Server Error", "error", err)
	return ErrInternalServer(err)
}

// ErrInternalServer returns a generic server error to the user
func ErrInternalServer(err error) render.Renderer {
	raven.CaptureError(err, map[string]string{
		"type": "Internal Server Error",
	})
	return &ErrResponse{
		Err:        err,
		StatusCode: http.StatusInternalServerError,
		Message:    "Something went wrong. Please contact Buyte Support.",
	}
}
