package googlepay

type TokenizationData struct {
	Type  string `json:"type,omitempty"`
	Token string `json:"token,omitempty"`
}
type PaymentMethodDataInfo struct {
	CardNetwork    string  `json:"cardNetwork,omitempty"`
	CardDetails    string  `json:"cardDetails,omitempty"`
	BillingAddress Address `json:"billingAddress,omitempty"`
}
type PaymentMethodData struct {
	Description      string                `json:"description,omitempty"`
	TokenizationData TokenizationData      `json:"tokenizationData,omitempty"`
	Type             string                `json:"type,omitempty"`
	Info             PaymentMethodDataInfo `json:"info,omitempty"`
}
type ShippingOptionData struct {
	ID string `json:"id,omitempty"`
}
type Address struct {
	PhoneNumber        string `json:"phoneNumber,omitempty"`
	Address3           string `json:"address3,omitempty"`
	SortingCode        string `json:"sortingCode,omitempty"`
	Address2           string `json:"address2,omitempty"`
	CountryCode        string `json:"countryCode,omitempty"`
	Address1           string `json:"address1,omitempty"`
	PostalCode         string `json:"postalCode,omitempty"`
	Name               string `json:"name,omitempty"`
	Locality           string `json:"locality,omitempty"`
	AdministrativeArea string `json:"administrativeArea,omitempty"`
}
type Response struct {
	ApiVersionMinor    int                `json:"apiVersionMinor,omitempty"`
	ApiVersion         int                `json:"apiVersion,omitempty"`
	PaymentMethodData  PaymentMethodData  `json:"paymentMethodData,omitempty"`
	ShippingOptionData ShippingOptionData `json:"shippingOptionData,omitempty"`
	ShippingAddress    Address            `json:"shippingAddress,omitempty"`
	BillingAddress     Address            `json:"billingAddress,omitempty"`
	Email              string             `json:"email,omitempty"`
}
