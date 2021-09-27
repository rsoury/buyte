package conf

import (
	config "github.com/spf13/viper"
)

func init() {

	// Logger Defaults
	config.SetDefault("logger.level", "debug")
	config.SetDefault("logger.encoding", "console")
	config.SetDefault("logger.color", true)
	config.SetDefault("logger.dev_mode", true)
	config.SetDefault("logger.disable_caller", false)
	config.SetDefault("logger.disable_stacktrace", false)

	// Pidfile
	config.SetDefault("pidfile", "")

	// Profiler config
	config.SetDefault("profiler.enabled", false)
	config.SetDefault("profiler.host", "")
	config.SetDefault("profiler.port", "6060")

	// Server Configuration
	config.SetDefault("server.container", false)
	config.SetDefault("server.production", false)
	config.SetDefault("server.host", "")
	config.SetDefault("server.port", "3002")
	config.SetDefault("server.tls", false)
	config.SetDefault("server.devcert", false)
	config.SetDefault("server.certfile", "server.crt")
	config.SetDefault("server.keyfile", "server.key")
	config.SetDefault("server.log_requests", true)
	config.SetDefault("server.log_cors", false)
	config.SetDefault("server.profiler_enabled", false)
	config.SetDefault("server.profiler_path", "/debug")
	config.SetDefault("server.sentry", "")
	config.SetDefault("server.mock.authorizer", true)

	// Database Settings
	config.SetDefault("storage.type", "graphql")
	config.SetDefault("storage.endpoint", "")

	// Merchant Settings -- Apple Pay
	config.SetDefault("apple.merchant.id", "merchant.com.buytecheckout")
	config.SetDefault("apple.merchant.name", "Buyte Apple Pay Checkout")
	config.SetDefault("apple.merchant.domain", "go.buytecheckout.com")
	config.SetDefault("apple.certs", "")

	// Merchant Settings -- Google Pay
	config.SetDefault("google.merchant.id", "05174216476243863888")
	config.SetDefault("google.merchant.name", "Buyte Google Pay Checkout")
	config.SetDefault("google.merchant.domain", "go.buytecheckout.com")

	// Lambda Functions Settings
	config.SetDefault("func.region", "ap-southeast-2")
	config.SetDefault("func.adyen_cse", "buyte-dev-adyen_cse")

	// Proxy Settings
	config.SetDefault("proxy.token", "06935288689c5a2935575744fcacff3e6bc92900")

	// User Pool
	config.SetDefault("config.clientId", "")
	config.SetDefault("config.userPoolId", "")

	// Stripe Settings
	config.SetDefault("stripe.live.secret", "")
	config.SetDefault("stripe.live.public", "")
	config.SetDefault("stripe.test.secret", "")
	config.SetDefault("stripe.test.public", "")
}
