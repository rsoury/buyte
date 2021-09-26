package buyte

type CustomerAddress struct {
	AddressLines          []string `json:"addressLines,omitempty"`
	AdministrativeArea    string   `json:"administrativeArea,omitempty"`
	Country               string   `json:"country,omitempty"`
	CountryCode           string   `json:"countryCode,omitempty"`
	Locality              string   `json:"locality,omitempty"`
	PostalCode            string   `json:"postalCode,omitempty"`
	SubAdministrativeArea string   `json:"subAdministrativeArea,omitempty"`
	SubLocality           string   `json:"subLocality,omitempty"`
}

type Customer struct {
	Name            string           `json:"name,omitempty"`
	GivenName       string           `json:"givenName,omitempty"`
	FamilyName      string           `json:"familyName,omitempty"`
	EmailAddress    string           `json:"emailAddress,omitempty"`
	PhoneNumber     string           `json:"phoneNumber,omitempty"`
	ShippingAddress *CustomerAddress `json:"shippingAddress,omitempty"`
	BillingAddress  *CustomerAddress `json:"billingAddress,omitempty"`
}

func (c *Customer) SetShippingAddress(address *CustomerAddress) {
	if !address.IsEmpty() {
		address.Filter()
		c.ShippingAddress = address
	}
}
func (c *Customer) SetBillingAddress(address *CustomerAddress) {
	if !address.IsEmpty() {
		address.Filter()
		c.BillingAddress = address
	}
}

// Check if all values are empty
func (c *CustomerAddress) IsEmpty() bool {
	if len(c.AddressLines) == 0 &&
		c.AdministrativeArea == "" &&
		c.Country == "" &&
		c.CountryCode == "" &&
		c.Locality == "" &&
		c.PostalCode == "" &&
		c.SubAdministrativeArea == "" &&
		c.SubLocality == "" {
		return true
	}
	return false
}

// Filter address lines and other other array of empty values.
func (c *CustomerAddress) Filter() {
	addressLines := []string{}
	for _, addressLine := range c.AddressLines {
		if addressLine != "" {
			addressLines = append(addressLines, addressLine)
		}
	}
	c.AddressLines = addressLines
}
