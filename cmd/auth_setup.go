package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/briandowns/spinner"
	. "github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	cli "github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/rsoury/buyte/buyte"
)

// authSetupCmd represents the auth-setup command
var authSetupCmd = &cli.Command{
	Use:   "auth-setup",
	Short: "Set up Cognito Auth",
	Long: `
		Set up the Cognito User Attributes required for Dashboard
	`,
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		AuthSetup(NewAWSConfigFromCmd(cmd))
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(authSetupCmd)

	AssignAWSFlags(authSetupCmd)
}

func AuthSetup(awsConfig *buyte.AWSConfig) {
	// Authenticate with Cognito
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(awsConfig.Region)},
	)
	// Create an APIGateway client from a aws session
	svc := cognito.New(sess)

	customAttributes := []*cognito.SchemaAttributeType{
		{
			AttributeDataType: aws.String("String"),
			Mutable:           aws.Bool(true),
			Name:              aws.String("phone_number"),
		},
		{
			AttributeDataType: aws.String("String"),
			Mutable:           aws.Bool(true),
			Name:              aws.String("store_name"),
		},
		{
			AttributeDataType: aws.String("String"),
			Mutable:           aws.Bool(true),
			Name:              aws.String("currency"),
		},
		{
			AttributeDataType: aws.String("String"),
			Mutable:           aws.Bool(true),
			Name:              aws.String("country"),
		},
	}

	// Set up User Attributes
	_, err := svc.AddCustomAttributes(&cognito.AddCustomAttributesInput{
		CustomAttributes: customAttributes,
		UserPoolId:       &awsConfig.CognitoUserPoolId,
	})

	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot add custom attributes"))
	}

	fmt.Println(Green("Cognito Custom User Attributes have been set up!"))
}
