package graphql

import (
	"context"

	"github.com/rsoury/buyte/pkg/user"
	"github.com/rsoury/buyte/test"
)

var (
	TestEnv *test.EnvConfig
)

func init() {
	TestEnv = test.Env()
}

func Context() context.Context {
	data := test.NewMock().Authentication(TestEnv.AuthToken)
	ctx := context.Background()
	userData, _ := user.Setup(func(key string) string {
		return data[key]
	})
	ctx = context.WithValue(ctx, "user", userData)
	return ctx
}
