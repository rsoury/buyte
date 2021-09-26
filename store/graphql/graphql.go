package graphql

import (
	"github.com/machinebox/graphql"
	"github.com/rs/xid"
	config "github.com/spf13/viper"
	"go.uber.org/zap"
)

type Client struct {
	*graphql.Client
	logger *zap.SugaredLogger
	newID  func(descriptor string) string
}

func New() *Client {
	logger := zap.S().With("package", "storage")
	endpoint := config.GetString("storage.endpoint")

	graphqlClient := graphql.NewClient(endpoint)

	logger.Debugw("Connected to graphql server",
		"storage.endpoint", config.GetString("storage.endpoint"),
	)

	c := &Client{
		graphqlClient,
		logger,
		func(descriptor string) string {
			return descriptor + "_" + xid.New().String()
		},
	}

	return c
}
