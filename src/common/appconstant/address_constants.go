package appconstant

//URL Params for address service
const (
	SESSION_ID        = "X-Jabong-SessionId"
	USER_ID           = "X-Jabong-UserId"
	IO_QUERY          = "QUERY"
	IO_ADDRESS_RESULT = "RESULT"
	IO_HTTP_REQUEST   = "REQUEST"
	IO_REQUEST_PARAMS = "QUERYPARAMS"
)

const (
	URLPARAM_LIMIT       = "limit"
	URLPARAM_OFFSET      = "offset"
	URLPARAM_ADDRESSID   = "id"
	URLPARAM_POSTCODE    = "postcode"
	URLPARAM_ADDRESSTYPE = "type"
)

const (
	MYSQL_ERROR string = "MysqlError"
)

//Redis constants
const (
	ADDRESS_CACHE_KEY string = "address_list_key_%s"
)

//Encryption service end points
const (
	ENCRYPT_ENDPOINT = "/encryption/v1/encrypt/"
	DECRYPT_ENDPOINT = "/encryption/v1/decrypt/"
	BATCH_SIZE       = float64(50)
)

const (
	BILLING        = "billing"
	SHIPPING       = "shipping"
	OTHER          = "other"
	ALL            = "all"
	UPDATE_TYPE    = "update_type"
	DEFAULT_LIMIT  = 10
	DEFAULT_OFFSET = 0
	MAX_LIMIT      = 50
)

const (
	ADDRESS_ID      = "id"
	FIRST_NAME      = "firstname"
	LAST_NAME       = "lastname"
	PHONE           = "phone"
	ALTERNATE_PHONE = "alt_phone"
	ADDRESS1        = "address1"
	ADDRESS2        = "address2"
	CITY            = "city"
	REGION          = "state"
	ADDRESS_REGION  = "address_region"
	ADDRESS_TYPE    = "address_type"
	POSTCODE        = "postcode"
	COUNTRY         = "country"
	SMS_OPT         = "sms_opt"
	IS_OFFICE       = "is_office"
	PARAM_REQ       = "req"
)
