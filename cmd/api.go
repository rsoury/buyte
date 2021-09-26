package cmd

import (
	"errors"

	cli "github.com/spf13/cobra"
	config "github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/conf"
	"github.com/rsoury/buyte/server"
	"github.com/rsoury/buyte/store/graphql"
)

var (
	// apiCmd -- Will Start the API.
	apiCmd = &cli.Command{
		Use:   "api",
		Short: "Buyte API",
		Long:  "Start Buyte API",
		Run:   StartApi,
	}
)

func init() {
	rootCmd.AddCommand(apiCmd)
}

func StartApi(cmd *cli.Command, args []string) {
	conf.InitSignalHandler()
	defer func() {
		<-conf.Stop.Chan() // Wait until StopChan
		conf.Stop.Wait()   // Wait until everyone cleans up
		zap.L().Sync()     // Flush the logger
	}()

	// Using an interface for store abstracts the server logic's dependencies on graphql...
	var store buyte.Store
	var err error
	switch config.GetString("storage.type") {
	case "graphql":
		if config.GetString("storage.endpoint") == "" {
			logger.Fatalw("Could not start server",
				"error", errors.New("'storage.endpoint' is a required configuration. Please set via Env as STORAGE_ENDPOINT or in configuration file."),
				"value", config.GetString("storage.endpoint"),
			)
		}
		// Despite being the Graphql Client type, it satisfies the store interface
		store = graphql.New()
	default:
		logger.Fatalw("Could not start server",
			"error", errors.New("Invalid 'storage.type'"),
			"value", config.GetString("storage.type"),
		)
	}

	// Create the server
	s, err := server.New(store)
	if err != nil {
		logger.Fatalw("Could not start server",
			"error", err,
		)
	}

	err = s.ListenAndServe()
	if err != nil {
		logger.Fatalw("Could not start server",
			"error", err,
		)
	}
}
