package address

type AddressResult struct {
	Summery     AddressDetails    `json:"Summery,omitempty"`
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
	Id                      string
	FirstName               string
	LastName                string
	Address1                string
	Address2                string
	City                    string
	RegionName              string
	EncryptedPhone          string
	EncryptedAlternatePhone string
	AddressType             string
	AddressRegion           string
	Country                 string
	Phone                   string
	AlternatePhone          string
	PostCode                string
	SmsOpt                  string
	IsOffice                string
	Req                     string
}

type AddressResponse struct {
	Address1          string `json:"address1"`
	Address2          string `json:"address2"`
	IsOffice          string `json:"address_type"`
	City              string `json:"city"`
	CreatedAt         string `json:"created_at"`
	FirstName         string `json:"first_name"`
	Country           string `json:"fk_country"`
	FkCustomer        string `json:"fk_customer"`
	AddressRegion     string `json:"fk_customer_address_region"`
	Id                string `json:"id_customer_address"`
	IsDefaultBilling  string `json:"is_default_billing"`
	IsDefaultShipping string `json:"is_default_shipping"`
	LastName          string `json:"last_name"`
	Phone             string `json:"phone"`           //trick to unmarshel during test case execution
	AlternatePhone    string `json:"alternate_phone"` //trick to unmarshel during test case execution
	PostCode          string `json:"postcode"`
	RegionName        string `json:"region_name"`
	SmsOpt            string `json:"sms_opt"`
	UpdatedAt         string `json:"updated_at"`
}
