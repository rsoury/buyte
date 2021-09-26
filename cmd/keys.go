package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/briandowns/spinner"
	. "github.com/logrusorgru/aurora"
	cli "github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/rsoury/buyte/pkg/keymanager"
	"github.com/rsoury/buyte/pkg/util"
)

// TODO: Get email from UserId?

type StackInfo struct {
	ApiGatewayUsagePlan            string `json:"ApiGatewayUsagePlan"`
	CognitoUserPool                string `json:"CognitoUserPool"`
	Stage                          string `json:"Stage"`
	Region                         string `json:"Region"`
	ApiGatewayRestApi              string `json:"ApiGatewayRestApi"`
	ServerlessDeploymentBucketName string `json:"ServerlessDeploymentBucketName"`
}

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

		stage, _ := cmd.Flags().GetString("stage")
		userId, _ := cmd.Flags().GetString("user-id")
		email, _ := cmd.Flags().GetString("email")
		isPublic, _ := cmd.Flags().GetBool("public")

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

		AddKey(stage, userId, email, isPublic)
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
		userId, _ := cmd.Flags().GetString("user-id")

		if userId == "" {
			logger.Fatal("User ID flag required. ie. e7b859c1-81d0-4d0f-b839-6b4510304f1c")
		}

		DeleteKeys(stage, userId)
	},
}

func init() {
	// Add "keys" to "root"
	rootCmd.AddCommand(keysCmd)

	// Add "add" to "keys"
	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysDeleteCmd)

	keysAddCmd.PersistentFlags().BoolP("public", "p", false, "Is key public? False by default.")
	keysAddCmd.PersistentFlags().StringP("email", "e", "", "What is the email for this user?")
	keysAddCmd.PersistentFlags().StringP("user-id", "u", "", "What is the user ID for this user? Leave empty to use email as user ID")
	keysAddCmd.PersistentFlags().StringP("stage", "s", "dev", "The stage environment to create the keys for. (dev | prod)")
	keysAddCmd.MarkFlagRequired("email")
	keysAddCmd.MarkFlagRequired("user-id")
	keysDeleteCmd.PersistentFlags().StringP("user-id", "u", "", "What is the user ID for this user? Leave empty to use email as user ID")
	keysDeleteCmd.PersistentFlags().StringP("stage", "s", "dev", "The stage environment to create the keys for. (dev | prod)")
	keysDeleteCmd.MarkFlagRequired("user-id")
}

func loadStackInfo(stage string) (StackInfo, error) {
	// Read config from deploy stack output file
	// Open our jsonFile from bin folder.
	stackInfoFileName := stage + "-stack-info.json"
	filepath := path.Join(util.DirName(), "../", stackInfoFileName)
	plan, err := ioutil.ReadFile(filepath)
	if err != nil {
		return StackInfo{}, err
	}
	var stackInfo StackInfo
	err = json.Unmarshal(plan, &stackInfo)
	if err != nil {
		return StackInfo{}, err
	}

	return stackInfo, nil
}

func AddKey(stage, userId, email string, isPublic bool) {
	stackInfo, err := loadStackInfo(stage)
	if err != nil {
		logger.Fatalw("Cannot open stack info file",
			"file", stage+"-stack-info.json",
			"err", err,
		)
		return
	}

	// Initiate KeyManager
	manager := keymanager.NewKeyManager(userId, email, &keymanager.AWSConfig{
		Region:                stackInfo.Region,
		APIGatewayId:          stackInfo.ApiGatewayRestApi,
		APIGatewayStage:       stackInfo.Stage,
		APIGatewayUsagePlanId: stackInfo.ApiGatewayUsagePlan,
		CognitoUserPoolId:     stackInfo.CognitoUserPool,
	})

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

func DeleteKeys(stage, userId string) {
	stackInfo, err := loadStackInfo(stage)
	if err != nil {
		logger.Fatalw("Cannot open stack info file",
			"file", stage+"-stack-info.json",
			"err", err,
		)
		return
	}

	// Initiate KeyManager
	manager := keymanager.NewKeyManager(userId, "", &keymanager.AWSConfig{
		Region:                stackInfo.Region,
		APIGatewayId:          stackInfo.ApiGatewayRestApi,
		APIGatewayStage:       stackInfo.Stage,
		APIGatewayUsagePlanId: stackInfo.ApiGatewayUsagePlan,
		CognitoUserPoolId:     stackInfo.CognitoUserPool,
	})

	err = manager.DeleteUserKeys()
	if err != nil {
		logger.Fatalw("Failed to delete API keys", `user`, manager.UserId)
		fmt.Println(Red("Failed"))
		return
	}

	fmt.Println(Green("SUCCESS!"))
}
