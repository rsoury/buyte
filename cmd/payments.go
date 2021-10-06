package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/logrusorgru/aurora"
	cli "github.com/spf13/cobra"

	"github.com/rsoury/buyte/buyte"
	"github.com/rsoury/buyte/pkg/user"
	store "github.com/rsoury/buyte/store/graphql"
)

type DeletePaymentParams struct {
	Id string `json:"id"`
}

type CreatePaymentParams struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type UpdatePaymentParams struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type ResponseProviderParams struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type ListResponsePaymentParams struct {
	Items []*ResponseProviderParams `json:"items"`
}

type PaymentsCommand struct {
	AWSConfig       *buyte.AWSConfig
	User            *user.SuperUser
	Cognito         *cognito.CognitoIdentityProvider
	AuthAccessToken *string
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

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		command, err := NewPaymentsCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.AddPayment(name, image)
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

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		id, _ := cmd.Flags().GetString("id")

		command, err := NewPaymentsCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.DeletePayment(id)
	},
}

var paymentsUpdateCmd = &cli.Command{
	Use:   "update",
	Short: "Update a Payment Option",
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		command, err := NewPaymentsCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		command.UpdatePayment(id, name, image)
	},
}

var paymentsListCmd = &cli.Command{
	Use:   "list",
	Short: "List Payment Options",
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		command, err := NewPaymentsCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.ListPayments()
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(paymentsCmd)

	AssignAWSFlags(paymentsCmd)

	userEnvConfig := user.NewSuperUserEnvConfig()
	paymentsCmd.PersistentFlags().StringP("email", "e", userEnvConfig.Username, "The User Username/Email.")
	paymentsCmd.PersistentFlags().StringP("password", "p", userEnvConfig.Password, "The User Password.")

	// Add "add" to "wallets"
	paymentsCmd.AddCommand(paymentsAddCmd)
	paymentsCmd.AddCommand(paymentsUpdateCmd)
	paymentsCmd.AddCommand(paymentsDeleteCmd)
	paymentsCmd.AddCommand(paymentsListCmd)

	paymentsAddCmd.PersistentFlags().String("name", "", "The name of the payment option")
	paymentsAddCmd.PersistentFlags().String("image", "", "The URL of the image for the payment option")
	paymentsAddCmd.MarkFlagRequired("name")
	paymentsDeleteCmd.PersistentFlags().String("id", "", "The database id of the payment option")
	paymentsDeleteCmd.MarkFlagRequired("id")
	paymentsUpdateCmd.PersistentFlags().String("id", "", "The database id of the payment option")
	paymentsUpdateCmd.PersistentFlags().String("name", "", "The name of the payment option")
	paymentsUpdateCmd.PersistentFlags().String("image", "", "The URL of the image for the payment option")
	paymentsUpdateCmd.MarkFlagRequired("id")
	paymentsUpdateCmd.MarkFlagRequired("name")
}

func NewPaymentsCommand(awsConfig *buyte.AWSConfig, user *user.SuperUser) (*PaymentsCommand, error) {
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

	input := &cognito.AdminInitiateAuthInput{
		ClientId:   &awsConfig.CognitoClientId,
		UserPoolId: &awsConfig.CognitoUserPoolId,
		AuthFlow:   aws.String("ADMIN_USER_PASSWORD_AUTH"),
	}
	input.SetAuthParameters(authParameters)

	auth, err := svc.AdminInitiateAuth(input)

	if err != nil {
		return nil, err
	}

	return &PaymentsCommand{
		AWSConfig:       awsConfig,
		User:            user,
		Cognito:         svc,
		AuthAccessToken: auth.AuthenticationResult.AccessToken,
	}, nil
}

func (c *PaymentsCommand) AddPayment(name, image string) {
	params := &CreatePaymentParams{
		Name:  name,
		Image: image,
	}

	req := graphql.NewRequest(`
		mutation CreateMobileWebPayment($input: CreateMobileWebPaymentInput!) {
			createMobileWebPayment(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", params)
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute creation of new payment option"))
	}

	response := &ResponseProviderParams{}
	err = mapstructure.Decode(respData["createMobileWebPayment"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment option response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment option response has no ID"))
	}

	fmt.Println(aurora.Green("New payment option " + name + " (" + response.Id + ") has been created!"))
}

func (c *PaymentsCommand) UpdatePayment(id, name, image string) {
	params := &UpdatePaymentParams{
		Id:    id,
		Name:  name,
		Image: image,
	}

	req := graphql.NewRequest(`
		mutation UpdateMobileWebPayment($input: UpdateMobileWebPaymentInput!) {
			updateMobileWebPayment(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", params)
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute update of payment option"))
	}

	response := &ResponseProviderParams{}
	err = mapstructure.Decode(respData["updateMobileWebPayment"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment option response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment option response has no ID"))
	}

	fmt.Println(aurora.Green("Payment option " + name + " (" + response.Id + ") has been update!"))
}

func (c *PaymentsCommand) DeletePayment(id string) {
	req := graphql.NewRequest(`
		mutation DeleteMobileWebPayment($input: DeleteMobileWebPaymentInput!) {
			deleteMobileWebPayment(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", &DeletePaymentParams{
		Id: id,
	})
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute deletion of payment option"))
	}

	response := &ResponseProviderParams{}
	err = mapstructure.Decode(respData["deleteMobileWebPayment"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment option response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment option response has no ID"))
	}

	fmt.Println(aurora.Green("Payment option " + response.Name + " (" + response.Id + ") has been deleted"))
}

func (c *PaymentsCommand) ListPayments() {
	req := graphql.NewRequest(`
		query ListMobileWebPayments {
			listMobileWebPayments {
				items {
					id
					name
					image
				}
			}
		}
	`)
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute list of payment option"))
	}

	response := &ListResponsePaymentParams{}
	err = mapstructure.Decode(respData["listMobileWebPayments"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment option response parameters"))
	}

	fmt.Println("")
	if len(response.Items) > 0 {
		for _, item := range response.Items {
			fmt.Println(aurora.Yellow("Id: " + item.Id))
			fmt.Println("Name: " + item.Name)
			fmt.Println("Image URL: " + item.Image)
			fmt.Println("-----------")
		}
	} else {
		fmt.Println("No payment options exist")
	}
}
