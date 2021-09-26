package test

import (
	"strings"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	config "github.com/spf13/viper"

	"github.com/rsoury/buyte/pkg/util"
)

type EnvConfig struct {
	AuthToken string `env:"AUTH_TOKEN,required"`
}

func init() {
	err := util.DotenvLoad(
		".env.development.local",
		".env.development",
		".env.local",
		".env",
		".env.test.local",
		".env.test",
	)
	if err != nil {
		panic(errors.Wrap(err, "Cannot Load .env files -- TEST."))
	}

	// Sets up the config file, environment etc
	config.SetTypeByDefaultValue(true)                      // If a default value is []string{"a"} an environment variable of "a b" will end up []string{"a","b"}
	config.AutomaticEnv()                                   // Automatically use environment variables where available
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Environement variables use underscores instead of periods
}

func Env() *EnvConfig {
	testEnv := &EnvConfig{}
	err := env.Parse(testEnv)
	if err != nil {
		panic(errors.Wrap(err, "Cannot Marshal Environment into Config -- TEST."))
	}
	return testEnv
}
