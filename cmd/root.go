package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/util"

	"net/http"
	_ "net/http/pprof" // Import for pprof

	cli "github.com/spf13/cobra"
	config "github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rsoury/buyte/conf"
)

var (

	// Config and global logger
	configFile string
	pidFile    string
	verbose    bool
	logger     *zap.SugaredLogger

	// The Root Cli Handler
	rootCmd = &cli.Command{
		Version: conf.Version,
		Use:     conf.Executable,
		Short:   conf.Title,
		Long:    conf.Description,
		PersistentPreRunE: func(cmd *cli.Command, args []string) error {
			// Create Pid File
			pidFile = config.GetString("pidfile")
			if pidFile != "" {
				file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
				if err != nil {
					return fmt.Errorf("Could not create pid file: %s Error:%v", pidFile, err)
				}
				defer file.Close()
				_, err = fmt.Fprintf(file, "%d\n", os.Getpid())
				if err != nil {
					return fmt.Errorf("Could not create pid file: %s Error:%v", pidFile, err)
				}
			}
			return nil
		},
		PersistentPostRun: func(cmd *cli.Command, args []string) {
			// Remove Pid file
			if pidFile != "" {
				os.Remove(pidFile)
			}
		},
	}
)

// Execute starts the program
func Execute() {
	// Run the program
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func StartEnv() {
	initEnv()
	initConfig()
	initLog()
	initProfiler()
}

// This is the main initializer handling cli, config and log
func init() {
	// Initialize configuration
	cli.OnInitialize(initEnv, initConfig, initLog, initProfiler)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose console output")
}

// initEnv reads in the Enviroment Variables in any applicable .env files.
func initEnv() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}
	_ = util.DotenvLoad(
		".env."+appEnv+".local",
		".env."+appEnv,
		".env.local",
		".env",
	)
	// Set maps/aliases
	if serverPort := os.Getenv("SERVER_PORT"); serverPort == "" {
		os.Setenv("SERVER_PORT", os.Getenv("PORT"))
	}
	if sentryEnv := os.Getenv("SENTRY_ENVIRONMENT"); sentryEnv == "" {
		os.Setenv("SENTRY_ENVIRONMENT", appEnv)
	}
	if sentryRelease := os.Getenv("SENTRY_RELEASE"); sentryRelease == "" {
		os.Setenv("SENTRY_RELEASE", conf.Version)
	}
	// Overwrites
	if region := os.Getenv("AWS_REGION"); region != "" {
		os.Setenv("FUNC_REGION", region)
	}

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Could not load environment from .env files ERROR: %s\n", err.Error())
	// 	os.Exit(1)
	// }
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// Sets up the config file, environment etc
	config.SetTypeByDefaultValue(true)                      // If a default value is []string{"a"} an environment variable of "a b" will end up []string{"a","b"}
	config.AutomaticEnv()                                   // Automatically use environment variables where available
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Environement variables use underscores instead of periods

	// If a config file is found, read it in.
	if configFile != "" {
		config.SetConfigFile(configFile)
		err := config.ReadInConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read config file: %s ERROR: %s\n", configFile, err.Error())
			os.Exit(1)
		}
	}

}

func initLog() {

	logConfig := zap.NewProductionConfig()

	// Log Level
	var logLevel zapcore.Level
	if err := logLevel.Set(config.GetString("logger.level")); err != nil {
		zap.S().Fatalw("Could not determine logger.level", "error", err)
	}
	logConfig.Level.SetLevel(logLevel)

	// Settings
	logConfig.Encoding = config.GetString("logger.encoding")
	logConfig.Development = config.GetBool("logger.dev_mode")
	logConfig.DisableCaller = config.GetBool("logger.disable_caller")
	logConfig.DisableStacktrace = config.GetBool("logger.disable_stacktrace")

	// Enable Color
	if config.GetBool("logger.color") {
		logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Use sane timestamp when logging to console
	if logConfig.Encoding == "console" {
		logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// JSON Fields
	logConfig.EncoderConfig.MessageKey = "msg"
	logConfig.EncoderConfig.LevelKey = "level"
	logConfig.EncoderConfig.CallerKey = "caller"

	// Build the logger
	globalLogger, _ := logConfig.Build()
	zap.ReplaceGlobals(globalLogger)
	logger = globalLogger.Sugar().With("package", "cmd")

	if verbose {
		logger.Infow("Buyte CLI Verbose", "Config", config.AllSettings())
	}

}

// Profliter can explicitly listen on address/port
func initProfiler() {
	if config.GetBool("profiler.enabled") {
		hostPort := net.JoinHostPort(config.GetString("profiler.host"), config.GetString("profiler.port"))
		go http.ListenAndServe(hostPort, nil)
		logger.Infof("Profiler enabled on http://%s", hostPort)
	}
}

func NewAWSConfigFromCmd(cmd *cli.Command) *buyte.AWSConfig {
	stage, _ := cmd.Flags().GetString("stage")
	region, _ := cmd.Flags().GetString("region")
	apiGatewayId, _ := cmd.Flags().GetString("api-gateway-id")
	apiGatewayUsagePlanId, _ := cmd.Flags().GetString("api-gateway-usage-plan-id")
	cognitoUserPoolId, _ := cmd.Flags().GetString("cognito-user-pool-id")
	cognitoClientId, _ := cmd.Flags().GetString("cognito-client-id")

	return &buyte.AWSConfig{
		Region:                region,
		APIGatewayId:          apiGatewayId,
		APIGatewayStage:       stage,
		APIGatewayUsagePlanId: apiGatewayUsagePlanId,
		CognitoUserPoolId:     cognitoUserPoolId,
		CognitoClientId:       cognitoClientId,
	}
}

func AssignAWSFlags(cmd *cli.Command) {
	envConfig := buyte.NewEnvConfig()
	cmd.PersistentFlags().StringP("region", "r", envConfig.Region, "The region of the environment.")
	cmd.PersistentFlags().StringP("stage", "s", envConfig.APIGatewayStage, "The stage environment.")
	cmd.PersistentFlags().String("api-gateway-id", envConfig.APIGatewayId, "The API Gateway ID to use.")
	cmd.PersistentFlags().String("api-gateway-usage-plan-id", envConfig.APIGatewayUsagePlanId, "The API Gateway Usage Plan ID to associate the new API keys to.")
	cmd.PersistentFlags().String("cognito-user-pool-id", envConfig.CognitoUserPoolId, "The Cognito User Pool ID that the User belongs to.")
	cmd.PersistentFlags().String("cognito-client-id", envConfig.CognitoClientId, "The Cognito Client ID that the User belongs to.")
}
