package appconstant

//Time constants
const (
	DATETIME_FORMAT = "2006-01-02 15:04:05"
)

//Expiration Time of cache key
const (
	EXPIRATION_TIME = 3600
)

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
	URLPARAM_ADDRESSID   = "addressId"
	URLPARAM_POSTCODE    = "postcode"
	URLPARAM_ADDRESSTYPE = "addressType"
	URLPARAM_DEFAULT     = "default"
)

const (
	MYSQL_ERROR string = "MysqlError"
)

//Redis constants
const (
	ADDRESS_CACHE_KEY string = "address_list_key_%s"
	ORDER_CACHE_KEY   string = "order_list_key_%s"
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
	FIRST_NAME      = "FirstName"
	LAST_NAME       = "LastName"
	PHONE           = "Phone"
	ALTERNATE_PHONE = "AlternatePhone"
	ADDRESS1        = "Address1"
	ADDRESS2        = "Address2"
	CITY            = "City"
	REGION          = "RegionName"
	ADDRESS_REGION  = "AddressRegion"
	IS_OFFICE       = "AddressType"
	POSTCODE        = "PostCode"
	COUNTRY         = "Country"
	SMS_OPT         = "Sms_opt"
)
