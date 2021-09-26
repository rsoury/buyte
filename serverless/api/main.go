package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	config "github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/cmd"
	"github.com/rsoury/buyte/conf"
	"github.com/rsoury/buyte/server"
	"github.com/rsoury/buyte/store/graphql"
)

var (
	httpLambda *httpadapter.HandlerAdapter
	logger     *zap.SugaredLogger
)

func init() {
	conf.InitSignalHandler()
	defer func() {
		<-conf.Stop.Chan() // Wait until StopChan
		conf.Stop.Wait()   // Wait until everyone cleans up
		zap.L().Sync()     // Flush the logger
	}()

	// Start log, config, etc.
	cmd.StartEnv()

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

	handler := s.InitialiseServer()
	httpLambda = httpadapter.New(handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return httpLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
