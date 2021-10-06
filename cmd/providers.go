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

type DeleteProviderParams struct {
	Id string `json:"id"`
}

type CreateProviderParams struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type UpdateProviderParams struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type RepsonseProviderParams struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Image          string `json:"image"`
	PaymentOptions struct {
		Items []struct {
			ConnectionId  string                  `json:"id" mapstructure:"id"`
			PaymentOption *ResponseProviderParams `json:"paymentOption"`
		} `json:"items"`
	} `json:"paymentOptions"`
}

type ListResponseProviderParams struct {
	Items []*RepsonseProviderParams `json:"items"`
}

type ProviderCommand struct {
	AWSConfig       *buyte.AWSConfig
	User            *user.SuperUser
	Cognito         *cognito.CognitoIdentityProvider
	AuthAccessToken *string
}

type ConnectProviderParams struct {
	ProviderId string `json:"providerPaymentOptionProviderId"`
	PaymentId  string `json:"providerPaymentOptionPaymentOptionId"`
}

type DisconnectProviderParams struct {
	Id string `json:"id"`
}

type ResponseConnectProviderParams struct {
	Id string `json:"id"`
}

// providersCmd represents the wallets command
var providersCmd = &cli.Command{
	Use:   "providers",
	Short: "Manage Buyte Payment Providers",
}

// providersAddCmd represents the add command
var providersAddCmd = &cli.Command{
	Use:   "add",
	Short: "Add a Payment Provider",
	Long: `
		A method to quickly add a Payment Provider to Buyte.

		ie. Stripe or Adyen
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

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.AddProvider(name, image)
	},
}

// providersUpdateCmd represents the add command
var providersUpdateCmd = &cli.Command{
	Use:   "update",
	Short: "Update a Payment Provider",
	Long: `
		A method to update a Payment Provider to Buyte.
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
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.UpdateProvider(id, name, image)
	},
}

var providersDeleteCmd = &cli.Command{
	Use:   "delete",
	Short: "Delete a Payment Provider",
	Long: `
		A method to delete a Payment Provider from Buyte.

		Be sure to disconnect all Payment Options from the Payment Provider before deleting.
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

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.DeleteProvider(id)
	},
}

var providersListCmd = &cli.Command{
	Use:   "list",
	Short: "List Payment Providers",
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		command.ListProviders()
	},
}

var providersConnectCmd = &cli.Command{
	Use:   "connect",
	Short: "Connect a Payment Option to a Payment Providers",
	Long: `
		Specify the Payment Options offered by the Payment Provider.

		Some Payment Providers can only support a select few Payment Options.
	`,
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		paymentId, _ := cmd.Flags().GetString("payment-id")
		providerId, _ := cmd.Flags().GetString("provider-id")

		command.ConnectPaymentOption(providerId, paymentId)
	},
}

var providersDisconnectCmd = &cli.Command{
	Use:   "disconnect",
	Short: "Disconnect a Payment Option from a Payment Providers",
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")

		command, err := NewProviderCommand(NewAWSConfigFromCmd(cmd), &user.SuperUser{
			Username: email,
			Password: password,
		})

		if err != nil {
			zap.S().Fatal(errors.Wrap(err, "Cannot create command and authenticate with user"))
		}

		connectionId, _ := cmd.Flags().GetString("connection-id")

		command.DisconnectPaymentOption(connectionId)
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(providersCmd)

	AssignAWSFlags(providersCmd)

	userEnvConfig := user.NewSuperUserEnvConfig()
	providersCmd.PersistentFlags().StringP("email", "e", userEnvConfig.Username, "The User Username/Email.")
	providersCmd.PersistentFlags().StringP("password", "p", userEnvConfig.Password, "The User Password.")

	// Add "add" to "wallets"
	providersCmd.AddCommand(providersAddCmd)
	providersCmd.AddCommand(providersDeleteCmd)
	providersCmd.AddCommand(providersUpdateCmd)
	providersCmd.AddCommand(providersListCmd)
	providersCmd.AddCommand(providersConnectCmd)
	providersCmd.AddCommand(providersDisconnectCmd)

	providersAddCmd.PersistentFlags().String("name", "", "The name of the payment provider")
	providersAddCmd.PersistentFlags().String("image", "", "The URL of the image for the payment provider")
	providersAddCmd.MarkFlagRequired("name")
	providersDeleteCmd.PersistentFlags().String("id", "", "The database id of the payment provider")
	providersDeleteCmd.MarkFlagRequired("id")
	providersUpdateCmd.PersistentFlags().String("id", "", "The database id of the payment provider")
	providersUpdateCmd.PersistentFlags().String("name", "", "The name of the payment provider")
	providersUpdateCmd.PersistentFlags().String("image", "", "The URL of the image for the payment provider")
	providersUpdateCmd.MarkFlagRequired("id")
	providersUpdateCmd.MarkFlagRequired("name")

	providersConnectCmd.PersistentFlags().String("payment-id", "", "The ID of the payment option to connect to the payment provider")
	providersConnectCmd.PersistentFlags().String("provider-id", "", "The ID of the payment provider to create connections for")
	providersDisconnectCmd.PersistentFlags().String("connection-id", "", "The ID of the payment option to provider connection")
}

func NewProviderCommand(awsConfig *buyte.AWSConfig, user *user.SuperUser) (*ProviderCommand, error) {
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

	return &ProviderCommand{
		AWSConfig:       awsConfig,
		User:            user,
		Cognito:         svc,
		AuthAccessToken: auth.AuthenticationResult.AccessToken,
	}, nil
}

func (c *ProviderCommand) AddProvider(name, image string) {
	params := &CreatePaymentParams{
		Name:  name,
		Image: image,
	}

	req := graphql.NewRequest(`
		mutation CreatePaymentProvider($input: CreatePaymentProviderInput!) {
			createPaymentProvider(input: $input) {
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
		zap.S().Fatal(errors.Wrap(err, "Cannot execute creation of new payment provider"))
	}

	response := &RepsonseProviderParams{}
	err = mapstructure.Decode(respData["createPaymentProvider"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment provider response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment provider response has no ID"))
	}

	fmt.Println(aurora.Green("New payment provider " + name + " (" + response.Id + ") has been created!"))
}

func (c *ProviderCommand) UpdateProvider(id, name, image string) {
	req := graphql.NewRequest(`
		mutation UpdatePaymentProvider($input: UpdatePaymentProviderInput!) {
			updatePaymentProvider(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", &UpdateProviderParams{
		Id:    id,
		Name:  name,
		Image: image,
	})
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute update of payment provider"))
	}

	response := &RepsonseProviderParams{}
	err = mapstructure.Decode(respData["updatePaymentProvider"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment provider response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment provider response has no ID"))
	}

	fmt.Println(aurora.Green("Payment provider " + response.Name + " (" + response.Id + ") has been update"))
}

func (c *ProviderCommand) DeleteProvider(id string) {
	req := graphql.NewRequest(`
		mutation DeletePaymentProvider($input: DeletePaymentProviderInput!) {
			deletePaymentProvider(input: $input) {
				id
				name
				image
			}
		}
	`)
	req.Var("input", &DeleteProviderParams{
		Id: id,
	})
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute deletion of payment provider"))
	}

	response := &RepsonseProviderParams{}
	err = mapstructure.Decode(respData["deletePaymentProvider"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment provider response parameters"))
	}

	if response.Id == "" {
		zap.S().Fatal(errors.Wrap(err, "Payment provider response has no ID"))
	}

	fmt.Println(aurora.Green("Payment provider " + response.Name + " (" + response.Id + ") has been deleted"))
}

func (c *ProviderCommand) ListProviders() {
	req := graphql.NewRequest(`
		query listPaymentProviders {
			listPaymentProviders {
				items {
					id
					name
					image
					paymentOptions {
						items {
							id
							paymentOption {
								id
								name
								image
							}
						}
					}
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
		zap.S().Fatal(errors.Wrap(err, "Cannot execute list of payment provider"))
	}

	response := &ListResponseProviderParams{}
	err = mapstructure.Decode(respData["listPaymentProviders"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure payment provider response parameters"))
	}

	fmt.Println("")
	if len(response.Items) > 0 {
		for _, item := range response.Items {
			fmt.Println(aurora.Yellow("Id: " + item.Id))
			fmt.Println("Name: " + item.Name)
			fmt.Println("Image URL: " + item.Image)
			fmt.Println("Payment Options: ")
			for _, option := range item.PaymentOptions.Items {
				fmt.Println(aurora.Gray(20, "\t Connection Id: "+option.ConnectionId))
				fmt.Println(aurora.Yellow("\t Id: " + option.PaymentOption.Id))
				fmt.Println("\t Name: " + option.PaymentOption.Name)
				fmt.Println("\t Image URL: " + option.PaymentOption.Image)
				fmt.Println("-----------")
			}
			fmt.Println("-----------")
		}
	} else {
		fmt.Println("No providers exist")
	}
}

func (c *ProviderCommand) ConnectPaymentOption(providerId, paymentId string) {
	req := graphql.NewRequest(`
		mutation CreateProviderPaymentOption(
			$input: CreateProviderPaymentOptionInput!
		) {
			createProviderPaymentOption(input: $input) {
				id
			}
		}
	`)
	req.Var("input", &ConnectProviderParams{
		ProviderId: providerId,
		PaymentId:  paymentId,
	})
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute payment option connection to payment provider"))
	}

	response := &ResponseConnectProviderParams{}
	err = mapstructure.Decode(respData["createProviderPaymentOption"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure response parameters"))
	}

	fmt.Println(aurora.Green("Connection (" + response.Id + ") successful"))
}

func (c *ProviderCommand) DisconnectPaymentOption(connectionId string) {
	req := graphql.NewRequest(`
		mutation DeleteProviderPaymentOption(
			$input: DeleteProviderPaymentOptionInput!
		) {
			deleteProviderPaymentOption(input: $input) {
				id
			}
		}
	`)
	req.Var("input", &DisconnectProviderParams{
		Id: connectionId,
	})
	req.Header.Set("Authorization", *c.AuthAccessToken)

	client := store.New()
	ctx := context.Background()

	var respData map[string]interface{}
	err := client.Run(ctx, req, &respData)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot execute payment option disconnection from payment provider"))
	}

	response := &ResponseConnectProviderParams{}
	err = mapstructure.Decode(respData["deleteProviderPaymentOption"], response)
	if err != nil {
		zap.S().Fatal(errors.Wrap(err, "Cannot destructure response parameters"))
	}

	fmt.Println(aurora.Green("Disconnection (" + response.Id + ") successful"))
}
