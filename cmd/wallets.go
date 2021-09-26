package cmd

import (
	"time"

	"github.com/briandowns/spinner"
	// . "github.com/logrusorgru/aurora"
	cli "github.com/spf13/cobra"

	"github.com/rsoury/buyte/pkg/keymanager"
)

// walletsCmd represents the wallets command
var walletsCmd = &cli.Command{
	Use:   "wallets",
	Short: "Manage Buyte Wallet Payment Options",
}

// walletsAddCmd represents the add command
var walletsAddCmd = &cli.Command{
	Use:   "add",
	Short: "Add a Wallet Payment Option",
	Long: `
		A method to quickly add a Wallet Payment Option to Buyte.

		ie. APPLE_PAY or GOOGLE_PAY
	`,
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		stage, _ := cmd.Flags().GetString("stage")
		region, _ := cmd.Flags().GetString("region")
		apiGatewayId, _ := cmd.Flags().GetString("api-gateway-id")
		apiGatewayUsagePlanId, _ := cmd.Flags().GetString("api-gateway-usage-plan-id")
		cognitoUserPoolId, _ := cmd.Flags().GetString("cognito-user-pool-id")

		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		providers, _ := cmd.Flags().GetStringArray("providers")

		AddWallet(&keymanager.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
		}, name, image, providers)
	},
}

var walletsDeleteCmd = &cli.Command{
	Use:   "delete",
	Short: "Delete a Wallet Payment Option",
	Long: `
		A method to quickly delete a Wallet Payment Option to Buyte.

		ie. APPLE_PAY or GOOGLE_PAY
	`,
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		stage, _ := cmd.Flags().GetString("stage")
		region, _ := cmd.Flags().GetString("region")
		apiGatewayId, _ := cmd.Flags().GetString("api-gateway-id")
		apiGatewayUsagePlanId, _ := cmd.Flags().GetString("api-gateway-usage-plan-id")
		cognitoUserPoolId, _ := cmd.Flags().GetString("cognito-user-pool-id")

		name, _ := cmd.Flags().GetString("name")

		DeleteWallet(&keymanager.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
		}, name)
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(walletsCmd)

	envConfig := keymanager.NewEnvConfig()
	walletsCmd.PersistentFlags().StringP("region", "r", envConfig.Region, "The region of the environment.")
	walletsCmd.PersistentFlags().StringP("stage", "s", envConfig.APIGatewayStage, "The stage environment.")
	walletsCmd.PersistentFlags().StringP("api-gateway-id", "a", envConfig.APIGatewayId, "The API Gateway ID to use.")
	walletsCmd.PersistentFlags().StringP("api-gateway-usage-plan-id", "x", envConfig.APIGatewayUsagePlanId, "The API Gateway Usage Plan ID to associate the new API keys to.")
	walletsCmd.PersistentFlags().StringP("cognito-user-pool-id", "c", envConfig.CognitoUserPoolId, "The Cognito User Pool ID that the User belongs to.")

	// Add "add" to "wallets"
	walletsCmd.AddCommand(walletsAddCmd)
	walletsCmd.AddCommand(walletsDeleteCmd)

	walletsAddCmd.MarkFlagRequired("name")
	walletsDeleteCmd.MarkFlagRequired("name")
}

func AddWallet(awsConfig *keymanager.AWSConfig, name, image string, providers []string) {
	// Authenticate with Cognito
}

func DeleteWallet(awsConfig *keymanager.AWSConfig, userId string) {
	// Authenticate with Cognito
}
