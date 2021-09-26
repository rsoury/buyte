package test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/rsoury/buyte/pkg/authenticate"
	"go.uber.org/zap"
)

type Mock struct {
	logger *zap.SugaredLogger
}

func NewMock() *Mock {
	return &Mock{
		logger: zap.S().With("package", "test.mock"),
	}
}

// Standalone func for testing.
func (m *Mock) Authentication(authToken string) map[string]string {
	user, err := authenticate.NewUserWithEnv(authToken)
	if err != nil {
		m.logger.Errorw("Could not create user to authenticate in MockAPIGateway...", "user", user.Id, "error", err)
		return nil
	}
	err = user.Authenticate()
	if err != nil {
		m.logger.Errorw("Could not authenticate user in MockAPIGateway...", "user", user.Id, "error", err)
		return nil
	}
	userAttr, err := json.Marshal(user.UserAttributes)
	if err != nil {
		m.logger.Errorw("Could not Marshal UserAttributes to JSON String in MockAPIGateway...", "user", user.Id, "error", err)
		return nil
	}
	devData := map[string]string{
		"UserId":          user.Id,
		"Token":           authToken,
		"BareToken":       user.BareToken,
		"IsPublic":        strconv.FormatBool(user.IsPublic),
		"IsAuthenticated": strconv.FormatBool(user.IsAuthenticated),
		"UserAttributes":  string(userAttr),
		"AccessToken":     *user.AuthenticationResult.AccessToken,
	}
	return devData
}

func (m *Mock) APIGatewayHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Inside func so it's triggered on request, not on app start.
		bearerToken := r.Header.Get("Authorization")
		if bearerToken == "" {
			m.logger.Error("Missing Authorization Bearer Token")
			return
		}
		authToken := strings.Replace(bearerToken, "Bearer ", "", 1)
		if authToken == "" {
			m.logger.Error("Invalid Authorization Bearer Token")
			return
		}

		devData := m.Authentication(authToken)

		// Append Mock Header values to Request.
		r.Header.Set("Accept", "*/*")
		r.Header.Set("Accept-Encoding", "gzip, deflate")
		r.Header.Set("Authorization", "Bearer "+devData["Token"])
		r.Header.Set("Cache-Control", "no-cache")
		r.Header.Set("Cloudfront-Forwarded-Proto", "https")
		r.Header.Set("Cloudfront-Is-Desktop-Viewer", "true")
		r.Header.Set("Cloudfront-Is-Mobile-Viewer", "false")
		r.Header.Set("Cloudfront-Is-Smarttv-Viewer", "false")
		r.Header.Set("Cloudfront-Is-Tablet-Viewer", "false")
		r.Header.Set("Cloudfront-Viewer-Country", "AU")
		r.Header.Set("Via", "1.1 eda9fe2763cea4a982a09ceb352512a6.cloudfront.net (CloudFront)")
		r.Header.Set("X-Amz-Cf-Id", "99oMbTGMnVqaNPx7umg2gs5VYjhNne16lPEXbgVEGD06ogKLdR8Rpg==")
		r.Header.Set("X-Amzn-Apigateway-Api-Id", "ftk4ht087h")
		r.Header.Set("X-Amzn-Trace-Id", "Root=1-5c7de124-f14e8ef0a1c011ec1527c530")
		// r.Header.Set("X-Forwarded-For", []string{"1.41.16.137, 70.132.29.79", "13.[13:38:29] 54.41.68"})
		r.Header.Set("X-Forwarded-Port", "443")
		r.Header.Set("X-Forwarded-Proto", "https")
		for key, value := range devData {
			r.Header.Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}
