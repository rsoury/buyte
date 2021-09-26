package cmd

import (
	"strings"

	"github.com/Pallinder/sillyname-go"
	inlets "github.com/inlets/inlets/pkg/client"
	cli "github.com/spf13/cobra"
	config "github.com/spf13/viper"
)

var (
	// proxyCmd -- Will Start the API.
	proxyCmd = &cli.Command{
		Use:   "proxy",
		Short: "Buyte Proxy",
		Long: `
		Start a Buyte Proxy.

		Connects to a remote inlets server which exposes a tunnel to your localhost.
		Essentially works as an ngrok alternative.
		Used to simplify multiple device development without the burder of connection limits.
	`,
		Run: StartProxy,
	}
)

func init() {
	rootCmd.AddCommand(proxyCmd)

	proxyCmd.PersistentFlags().StringP("port", "p", "3000", "The local port to proxy")
	proxyCmd.PersistentFlags().BoolP("tls", "t", false, "Use secure TLS/SSL localhost. ie. https://localhost:3000")
	proxyCmd.PersistentFlags().StringP("proxy-name", "n", "", "Any name to identify your proxy. Will default to a random name. ie. -n hello -> hello.proxy.buytesoftware.com")
}

func StartProxy(cmd *cli.Command, args []string) {
	port, _ := cmd.Flags().GetString("port")
	tls, _ := cmd.Flags().GetBool("tls")
	name, _ := cmd.Flags().GetString("proxy-name")

	token := config.GetString("proxy.token")

	if name == "" {
		name = strings.ToLower(strings.ReplaceAll(sillyname.GenerateStupidName(), " ", "-"))
	}

	proxyHost := name + `.proxy.buytesoftware.com`

	localhost := `http`
	if tls {
		localhost += `s`
	}
	localhost += `://localhost:` + port

	remote := `wss://` + proxyHost
	upstream := make(map[string]string)
	upstream[proxyHost] = localhost

	client := &inlets.Client{
		Remote:      remote,
		UpstreamMap: upstream,
		Token:       token,
	}

	err := client.Connect()
	if err != nil {
		logger.Fatalw("Buyte Proxy Failed. Could not connect to remote inlets server.",
			"remote", remote,
			"upstream", proxyHost+`=`+localhost,
			"token", token,
			"err", err,
		)
	}
}
