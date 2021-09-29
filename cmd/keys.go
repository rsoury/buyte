package cmd

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	. "github.com/logrusorgru/aurora"
	cli "github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/keymanager"
)

// keysCmd represents the keys command
var keysCmd = &cli.Command{
	Use:   "keys",
	Short: "Manage Buyte API Keys",
}

// keysAddCmd represents the add command
var keysAddCmd = &cli.Command{
	Use:   "add",
	Short: "Manage Buyte API Keys: Add an API Key",
	Long: `
		A method to quickly add API keys to Buyte.

		It uses the same process that occurs when a user is successfully authenticated and confirmed.
	`,
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		userId, _ := cmd.Flags().GetString("user-id")
		email, _ := cmd.Flags().GetString("email")
		isPublic, _ := cmd.Flags().GetBool("public")

		stage, _ := cmd.Flags().GetString("stage")
		region, _ := cmd.Flags().GetString("region")
		apiGatewayId, _ := cmd.Flags().GetString("api-gateway-id")
		apiGatewayUsagePlanId, _ := cmd.Flags().GetString("api-gateway-usage-plan-id")
		cognitoUserPoolId, _ := cmd.Flags().GetString("cognito-user-pool-id")

		if email == "" {
			logger.Fatal("Email flag required. ie. hello@buytecheckout.com")
		}
		if userId == "" {
			logger.Fatal("User ID flag required. ie. e7b859c1-81d0-4d0f-b839-6b4510304f1c")
		}

		// logger.Infow("Arguments list",
		// 	`stage`, stage,
		// 	`userId`, userId,
		// 	`email`, email,
		// 	`public`, isPublic,
		// )

		AddKey(&buyte.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
		}, userId, email, isPublic)
	},
}

var keysDeleteCmd = &cli.Command{
	Use:   "delete",
	Short: "Delete a user's API Keys",
	Long: `
		A method to quickly delete a user's Buyte API keys.
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

		userId, _ := cmd.Flags().GetString("user-id")

		if userId == "" {
			logger.Fatal("User ID flag required. ie. e7b859c1-81d0-4d0f-b839-6b4510304f1c")
		}

		DeleteKeys(&buyte.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
		}, userId)
	},
}

func init() {
	// Add "keys" to "root"
	rootCmd.AddCommand(keysCmd)

	envConfig := buyte.NewEnvConfig()
	keysCmd.PersistentFlags().StringP("region", "r", envConfig.Region, "The region of the environment.")
	keysCmd.PersistentFlags().StringP("stage", "s", envConfig.APIGatewayStage, "The stage environment.")
	keysCmd.PersistentFlags().String("api-gateway-id", envConfig.APIGatewayId, "The API Gateway ID to use.")
	keysCmd.PersistentFlags().String("api-gateway-usage-plan-id", envConfig.APIGatewayUsagePlanId, "The API Gateway Usage Plan ID to associate the new API keys to.")
	keysCmd.PersistentFlags().String("cognito-user-pool-id", envConfig.CognitoUserPoolId, "The Cognito User Pool ID that the User belongs to.")

	// Add "add" to "keys"
	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysDeleteCmd)

	keysAddCmd.PersistentFlags().BoolP("public", "p", false, "Is key public? False by default.")
	keysAddCmd.PersistentFlags().StringP("email", "e", "", "What is the email for this user?")
	keysAddCmd.PersistentFlags().StringP("user-id", "u", "", "What is the user ID for this user? Leave empty to use email as user ID")
	keysAddCmd.MarkFlagRequired("email")
	keysAddCmd.MarkFlagRequired("user-id")
	keysDeleteCmd.PersistentFlags().StringP("user-id", "u", "", "What is the user ID for this user? Leave empty to use email as user ID")
	keysDeleteCmd.MarkFlagRequired("user-id")
}

func AddKey(awsConfig *buyte.AWSConfig, userId, email string, isPublic bool) {
	// Initiate KeyManager
	manager := keymanager.NewKeyManager(userId, email, awsConfig)

	key := manager.GenerateKey(isPublic)
	apiKey, err := manager.CreateApiKey(key, isPublic)
	if err != nil {
		logger.Fatal("Could not create API key for user.",
			zap.String("user", userId),
			zap.Bool("isPublic", isPublic),
		)
		fmt.Println(Red("Failed"))
		return
	}

	err = manager.AssociateApiKeyIdsWithCognito([]*keymanager.CognitoAssociateData{
		&keymanager.CognitoAssociateData{
			IsPublic: isPublic,
			ApiKeyId: *apiKey.Id,
		},
	})
	if err != nil {
		logger.Fatal("Could not associate API Keys with Cognito Account for user", zap.String(`user`, manager.UserId))
		fmt.Println(Red("Failed"))
		return
	}

	fmt.Println(Green("SUCCESS!"))
}

func DeleteKeys(awsConfig *buyte.AWSConfig, userId string) {
	// Initiate KeyManager
	manager := keymanager.NewKeyManager(userId, "", awsConfig)

	err := manager.DeleteUserKeys()
	if err != nil {
		logger.Fatalw("Failed to delete API keys", `user`, manager.UserId)
		fmt.Println(Red("Failed"))
		return
	}

	fmt.Println(Green("SUCCESS!"))
}
