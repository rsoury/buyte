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
	"github.com/rsoury/buyte/pkg/user"
)

// createSuperUserCmd represents the create-super-user command
var createSuperUserCmd = &cli.Command{
	Use:   "create-super-user",
	Short: "Create Super User for Buyte",
	Run: func(cmd *cli.Command, args []string) {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()
		defer func() {
			s.Stop()
		}()

		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		cognitoUserPoolId, _ := cmd.Flags().GetString("cognito-user-pool-id")
		region, _ := cmd.Flags().GetString("region")

		CreateSuperUser(&buyte.AWSConfig{
			Region:            region,
			CognitoUserPoolId: cognitoUserPoolId,
		}, email, password)
	},
}

func init() {
	// Add "wallets" to "root"
	rootCmd.AddCommand(createSuperUserCmd)

	envConfig := buyte.NewEnvConfig()
	createSuperUserCmd.PersistentFlags().StringP("region", "r", envConfig.Region, "The region of the environment.")
	createSuperUserCmd.PersistentFlags().String("cognito-user-pool-id", envConfig.CognitoUserPoolId, "The Cognito User Pool ID that the User belongs to.")

	userEnvConfig := user.NewSuperUserEnvConfig()
	createSuperUserCmd.PersistentFlags().StringP("email", "e", userEnvConfig.Username, "The User Username/Email.")
	createSuperUserCmd.PersistentFlags().StringP("password", "p", userEnvConfig.Password, "The User Password.")
}

func CreateSuperUser(awsConfig *buyte.AWSConfig, email, password string) {
	// Authenticate with Cognito
	sess, _ := session.NewSession(
		&aws.Config{Region: aws.String(awsConfig.Region)},
	)
	// Create an APIGateway client from a aws session
	svc := cognito.New(sess)

	// Check if group exist
	groups, err := svc.ListGroups(&cognito.ListGroupsInput{
		UserPoolId: &awsConfig.CognitoUserPoolId,
		Limit: aws.Int64(50),
	});
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot fetch Groups"))
	}
	superUserGroupExists := false
	for _, group := range groups.Groups {
		if group.GroupName == aws.String("SuperUsers") {
			superUserGroupExists = true
			break
		}
	}

	// If not, create it
	if !superUserGroupExists {
		_, err = svc.CreateGroup(&cognito.CreateGroupInput{
			UserPoolId: &awsConfig.CognitoUserPoolId,
			GroupName: aws.String("SuperUsers"),
			Description: aws.String("A group of users who have Admin privileges."),
		})
		if err != nil {
			log.Fatal(errors.Wrap(err, "Cannot create SuperUsers Groups"))
		}
	}

	var attributes []*cognito.AttributeType
	attributes = append(attributes, &cognito.AttributeType{
		Name:  aws.String("email"),
		Value: &email,
	})
	newUser, err := svc.AdminCreateUser(&cognito.AdminCreateUserInput{
		Username:          &email,
		UserAttributes:    attributes,
		UserPoolId:        &awsConfig.CognitoUserPoolId,
		TemporaryPassword: &password,
	})

	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot create super user"))
	}

	_, err = svc.AdminAddUserToGroup(&cognito.AdminAddUserToGroupInput{
		GroupName:  aws.String("SuperUsers"),
		UserPoolId: &awsConfig.CognitoUserPoolId,
		Username:   newUser.User.Username,
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot add user to SuperUsers Group"))
		_, err = svc.AdminDeleteUser(&cognito.AdminDeleteUserInput{
			UserPoolId: &awsConfig.CognitoUserPoolId,
			Username: newUser.User.Username,
		});

		if err != nil {
			log.Fatal(errors.Wrap(err, "Cannot delete the user in error rollback"))
		}
	}

	fmt.Println(Green("Super user created!"))
}
