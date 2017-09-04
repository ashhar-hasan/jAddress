package address

//URL Params for address service
const (
	URL_PARAM_LIMIT        = "limit"
	URL_PARAM_OFFSET       = "offset"
	URL_PARAM_ADDRESS_ID   = "id"
	URL_PARAM_POSTCODE     = "postcode"
	URL_PARAM_ADDRESS_TYPE = "type"
)
const (
	BILLING        string = "billing"
	SHIPPING       string = "shipping"
	OTHER          string = "other"
	ALL            string = "all"
	UPDATE_TYPE    string = "update_type"
	DEFAULT_LIMIT  int    = 10
	DEFAULT_OFFSET int    = 0
	MAX_LIMIT      int    = 50
)

const (
	ADDRESS_ID      string = "id"
	FIRST_NAME      string = "firstname"
	LAST_NAME       string = "lastname"
	PHONE           string = "phone"
	ALTERNATE_PHONE string = "alt_phone"
	ADDRESS1        string = "address1"
	ADDRESS2        string = "address2"
	CITY            string = "city"
	REGION          string = "state"
	ADDRESS_REGION  string = "address_region"
	ADDRESS_TYPE    string = "address_type"
	POSTCODE        string = "postcode"
	COUNTRY         string = "country"
	SMS_OPT         string = "sms_opt"
	IS_OFFICE       string = "is_office"
	PARAM_REQ       string = "req"
)
