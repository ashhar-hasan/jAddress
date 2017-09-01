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
	Id                      uint32
	EncryptedPhone          string
	EncryptedAlternatePhone string
}

type DecryptedFields struct {
	Id                      uint32
	DecryptedPhone          int64
	DecryptedAlternatePhone int64
}

type AddressRequest struct {
	Id                      uint32
	FirstName               string
	LastName                string
	Address1                string
	Address2                string
	City                    string
	RegionName              string
	EncryptedPhone          string
	EncryptedAlternatePhone string
	AddressType             string
	AddressRegion           uint32
	Country                 uint32
	Phone                   int64
	AlternatePhone          int64
	PostCode                int
	SmsOpt                  int
	IsOffice                int
	Req                     string
}

type AddressResponse struct {
	Id             uint32 `json:"id"`
	FirstName      string `json:"firstname"`
	LastName       string `json:"lastname"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	PostCode       int    `json:"postcode"`
	City           string `json:"city"`
	RegionName     string `json:"region_name"`
	AddressRegion  uint32 `json:"fk_address_region"`
	Country        uint32 `json:"fk_country"`
	Phone          int64  `json:"phone,string"` //trick to unmarshel during test case execution
	AlternatePhone int64  `json:"alternate_phone,string"`
	AddressType    string `json:"address_type"`
	SmsOpt         int    `json:"sms_opt"`
	IsOffice       int    `json:"is_office"`
}
