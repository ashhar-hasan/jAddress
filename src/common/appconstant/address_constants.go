package appconstant

//URL Params for address service
const (
	SessionID        = "X-Jabong-SessionId"
	IoQuery          = "QUERY"
	IoAddressResult  = "RESULT"
	IoLocalityResult = "RESULT"
	IoHttpRequest    = "REQUEST"
	IoRequestParams  = "QUERYPARAMS"
)

const (
	UrlParamLimit       = "limit"
	UrlParamOffset      = "offset"
	UrlParamAddressId   = "id"
	UrlParamPostcode    = "postcode"
	UrlParamAddressType = "type"
)

const (
	MYSQLError string = "MysqlError"
)

const (
	Billing       = "billing"
	Shipping      = "shipping"
	Other         = "other"
	All           = "all"
	UpdateType    = "update_type"
	DefaultLimit  = 10
	DefaultOffset = 0
	MaxLimit      = 50
)

const (
	AddressId      = "id"
	FirstName      = "firstname"
	LastName       = "lastname"
	Phone          = "phone"
	AlternatePhone = "alt_phone"
	Address1       = "address1"
	Address2       = "address2"
	City           = "city"
	Region         = "state"
	AddressRegion  = "address_region"
	AddressType    = "address_type"
	Postcode       = "postcode"
	Country        = "country"
	SmsOpt         = "sms_opt"
	IsOffice       = "is_office"
	ParamReq       = "req"
)
