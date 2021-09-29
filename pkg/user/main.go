package user

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	config "github.com/spf13/viper"

	"go.uber.org/zap"
)

type User struct {
	logger          *zap.SugaredLogger
	ID              string
	Token           string
	BareToken       string
	IsPublic        bool
	IsAuthenticated bool
	UserAttributes  *UserAttributes
	AccessToken     string
}

type UserAttributes struct {
	Currency       string  `json:"custom:currency"`
	Country        string  `json:"custom:country"`
	StoreName      string  `json:"custom:store_name"`
	Website        string  `json:"website"`
	Logo           string  `json:"custom:logo"`
	CoverImage     string  `json:"custom:cover_image"`
	ShippingModule int8    `json:"custom:shipping_module,string"`
	FeeMultiplier  float64 `json:"custom:fee_multiplier,string"`
	AccountBalance int     `json:"custom:account_balance,string"`
	CustomCSS      string  `json:"custom:custom_css"`
}

type SuperUser struct {
	Username string `env:"ADMIN_USERNAME"`
	Password string `env:"ADMIN_PASSWORD"`
}

func NewSuperUserEnvConfig() *SuperUser {
	config := &SuperUser{}
	err := env.Parse(config)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot Marshal Environment into Config"))
	}
	return config
}

func Setup(get func(key string) string) (*User, error) {
	logger := zap.S().With("package", "user")

	IsPublic, _ := strconv.ParseBool(get("IsPublic"))
	IsAuthenticated, _ := strconv.ParseBool(get("IsAuthenticated"))

	userAttributes := &UserAttributes{}
	err := json.Unmarshal([]byte(get("UserAttributes")), userAttributes)
	if err != nil {
		return nil, err
	}

	// Unsure how to handle user attributes at the moment...
	user := &User{
		logger:          logger,
		ID:              get("UserId"),
		Token:           get("Token"),
		BareToken:       get("BareToken"),
		IsPublic:        IsPublic,
		IsAuthenticated: IsAuthenticated,
		UserAttributes:  userAttributes,
		AccessToken:     get("AccessToken"),
	}

	// logger.Debugw("Authenticated User", "user", user)

	return user, nil
}

func FromContext(ctx context.Context) *User {
	return ctx.Value("user").(*User)
}

func (u *User) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "user", u)
}

func (u *User) IncrementAccountBalance(amount int) error {
	balance := u.UserAttributes.AccountBalance + amount
	sess, _ := session.NewSession()
	svc := cognito.New(sess)

	// Update user attribute
	balanceAttrName := "custom:account_balance"
	balanceAttrValue := strconv.Itoa(balance)
	userPoolId := config.GetString("cognito.userPoolId")
	username := u.ID
	var updatedUserAttributes []*cognito.AttributeType
	updatedUserAttributes = append(updatedUserAttributes, &cognito.AttributeType{
		Name:  &balanceAttrName,
		Value: &balanceAttrValue,
	})
	_, err := svc.AdminUpdateUserAttributes(&cognito.AdminUpdateUserAttributesInput{
		UserAttributes: updatedUserAttributes,
		UserPoolId:     &userPoolId,
		Username:       &username,
	})

	if err != nil {
		return err
	}

	return nil
}
