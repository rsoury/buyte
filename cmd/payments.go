package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	g "github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/logrusorgru/aurora"
	cli "github.com/spf13/cobra"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/user"
	"github.com/rsoury/buyte/store/graphql"
)

type CreatePaymentParams struct {
	Name string `json:"name"`
	Image string `json:"image"`
}

type RepsonsePaymentParams struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Image string `json:"image"`
}

// paymentsCmd represents the wallets command
var paymentsCmd = &cli.Command{
	Use:   "payments",
	Short: "Manage Buyte Payment Options",
}

// paymentsAddCmd represents the add command
var paymentsAddCmd = &cli.Command{
	Use:   "add",
	Short: "Add a Payment Option",
	Long: `
		A method to quickly add a Payment Option to Buyte.

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
		cognitoClientId, _ := cmd.Flags().GetString("cognito-client-id")

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		AddPayment(&buyte.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
			CognitoClientId: 			 cognitoClientId,
		}, &user.SuperUser{
			Username: email,
			Password: password,
		}, name, image)
	},
}

var paymentsDeleteCmd = &cli.Command{
	Use:   "delete",
	Short: "Delete a Payment Option",
	Long: `
		A method to quickly delete a Payment Option from Buyte.

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

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		name, _ := cmd.Flags().GetString("name")

		DeletePayment(&buyte.AWSConfig{
			Region:                region,
			APIGatewayId:          apiGatewayId,
			APIGatewayStage:       stage,
			APIGatewayUsagePlanId: apiGatewayUsagePlanId,
			CognitoUserPoolId:     cognitoUserPoolId,
		}, &user.SuperUser{
			Username: email,
			Password: password,
		}, name)
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(paymentsCmd)

	envConfig := buyte.NewEnvConfig()
	paymentsCmd.PersistentFlags().StringP("region", "r", envConfig.Region, "The region of the environment.")
	paymentsCmd.PersistentFlags().StringP("stage", "s", envConfig.APIGatewayStage, "The stage environment.")
	paymentsCmd.PersistentFlags().String("api-gateway-id", envConfig.APIGatewayId, "The API Gateway ID to use.")
	paymentsCmd.PersistentFlags().String("api-gateway-usage-plan-id", envConfig.APIGatewayUsagePlanId, "The API Gateway Usage Plan ID to associate the new API keys to.")
	paymentsCmd.PersistentFlags().String("cognito-user-pool-id", envConfig.CognitoUserPoolId, "The Cognito User Pool ID that the User belongs to.")
	paymentsCmd.PersistentFlags().String("cognito-client-id", envConfig.CognitoClientId, "The Cognito Client ID that the User belongs to.")

	userEnvConfig := user.NewSuperUserEnvConfig()
	paymentsCmd.PersistentFlags().StringP("email", "e", userEnvConfig.Username, "The User Username/Email.")
	paymentsCmd.PersistentFlags().StringP("password", "p", userEnvConfig.Password, "The User Password.")

	// Add "add" to "wallets"
	paymentsCmd.AddCommand(paymentsAddCmd)
	paymentsCmd.AddCommand(paymentsDeleteCmd)

	paymentsAddCmd.PersistentFlags().String("name", "", "The name of the payment option")
	paymentsAddCmd.PersistentFlags().String("image", "", "The URL of the image for the payment option")
	paymentsAddCmd.MarkFlagRequired("name")
	paymentsDeleteCmd.PersistentFlags().String("name", "", "The name of the payment option")
	paymentsDeleteCmd.MarkFlagRequired("name")
}

func AddPayment(awsConfig *buyte.AWSConfig, user *user.SuperUser, name, image string) {
	// Authenticate with Cognito
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(awsConfig.Region)},
	)
	// Create an APIGateway client from a aws session
	svc := cognito.New(sess)

	authParameters := map[string]*string{
		"USERNAME": aws.String(user.Username),
		"PASSWORD": aws.String(user.Password),
	}

	auth, err := svc.AdminInitiateAuth(&cognito.AdminInitiateAuthInput{
		ClientId: &awsConfig.CognitoClientId,
		UserPoolId: &awsConfig.CognitoUserPoolId,
		AuthFlow: aws.String("ADMIN_USER_PASSWORD_AUTH"),
		AuthParameters: authParameters,
	})

	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot authenticate with user"))
	}

	params := &CreatePaymentParams{
		Name: name,
		Image: image,
	}

	req := g.NewRequest(`
		mutation CreateMobileWebPayment($input: CreateMobileWebPaymentInput!) {
			createMobileWebPayment(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", params)
	req.Header.Set("Authorization", *auth.AuthenticationResult.AccessToken)

	client := graphql.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err = client.Run(ctx, req, &respData);
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot execute creation of new payment option"))
	}

	response := &RepsonsePaymentParams{}
	err = mapstructure.Decode(respData["createMobileWebPayment"], response);
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot destructure new payment option response parameters"))
	}

	if response.Id == "" {
		log.Fatal(errors.Wrap(err, "New payment option response has not ID"))
	}

	fmt.Println(aurora.Green("New payment option " + name+" has been created!"))
}

func DeletePayment(awsConfig *buyte.AWSConfig, user *user.SuperUser, userId string) {
	// Authenticate with Cognito
}
