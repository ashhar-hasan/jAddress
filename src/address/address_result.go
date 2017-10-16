package address

type AddressResult struct {
	Summary     AddressDetails    `json:"Summary,omitempty"`
	AddressList []AddressResponse `json:"AddressList,omitempty"`
}

type AddressDetails struct {
	Count int    `json:"Count,omitempty"`
	Type  string `json:"Type,omitempty"`
}

type EncryptedFields struct {
	Id                      string
	EncryptedPhone          string
	EncryptedAlternatePhone string
}

type DecryptedFields struct {
	Id                      string
	DecryptedPhone          string
	DecryptedAlternatePhone string
}

type AddressRequest struct {
	FirstName               string
	LastName                string
	Address1                string
	Address2                string
	City                    string
	RegionName              string
	IsOffice                string `json:"AddressType"`
	AddressRegion           string
	Country                 string
	Phone                   string
	AlternatePhone          string
	EncryptedPhone          string
	EncryptedAlternatePhone string
	PostCode                string
	SmsOpt                  string `json:"Sms_opt"`
}

type AddressResponse struct {
	Id                string `json:"id_customer_address"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Address1          string `json:"address1"`
	Address2          string `json:"address2"`
	City              string `json:"city"`
	PostCode          string `json:"postcode"`
	Country           string `json:"fk_country"`
	IsOffice          string `json:"address_type"`
	RegionName        string `json:"region_name"`
	AddressRegion     string `json:"fk_customer_address_region"`
	Phone             string `json:"phone"`           //trick to unmarshel during test case execution
	AlternatePhone    string `json:"alternate_phone"` //trick to unmarshel during test case execution
	FkCustomer        string `json:"fk_customer"`
	IsDefaultBilling  string `json:"is_default_billing"`
	IsDefaultShipping string `json:"is_default_shipping"`
	SmsOpt            string `json:"sms_opt"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}
