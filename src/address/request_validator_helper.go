package address

import (
	"common/appconstant"
	"errors"
	"net/http"
	"strconv"

	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
)

func validateAndSetURLParams(params *RequestParams, httpReq *http.Request) error {
	var (
		limit  = appconstant.DEFAULT_LIMIT
		offset = appconstant.DEFAULT_OFFSET
		err    error
	)
	if httpReq.FormValue("limit") != "" {
		limit, err = utilHttp.GetIntParamFields(httpReq, appconstant.URLPARAM_LIMIT)
		if err != nil {
			return errors.New("Limit must be a valid number")
		}
	}
	if limit > appconstant.MAX_LIMIT {
		limit = appconstant.DEFAULT_LIMIT
	}
	params.QueryParams.Limit = limit
	if httpReq.FormValue(appconstant.URLPARAM_OFFSET) != "" {
		offset, err = utilHttp.GetIntParamFields(httpReq, appconstant.URLPARAM_OFFSET)
		if err != nil {
			return errors.New("Offset must be a number")
		}
	}
	params.QueryParams.Offset = offset
	return nil
}

func validateAddressType(str string) (addressType string, err error) {
	if str == appconstant.BILLING {
		addressType = appconstant.BILLING
	} else if str == appconstant.SHIPPING {
		addressType = appconstant.SHIPPING
	} else if str == appconstant.OTHER {
		addressType = appconstant.OTHER
	} else if str == appconstant.ALL {
		addressType = appconstant.ALL
	} else {
		return addressType, errors.New("Invalid address type. Possible types are billing, shipping, other, all")
	}
	return addressType, err
}

func validateAndSetParams(params *RequestParams, httpReq *utilHttp.Request) error {
	if httpReq.HTTPVerb == "GET" {
		val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSTYPE)
		if val == "" {
			val = appconstant.ALL
		}
		if val != appconstant.BILLING && val != appconstant.SHIPPING && val != appconstant.OTHER && val != appconstant.ALL {
			return errors.New("AddressType should be one of all, billing, shipping or other")
		}
		params.QueryParams.AddressType = val
		return nil
	}
	val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSID)
	addressID, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("Id is missing or not a number")
	}
	params.QueryParams.AddressId = addressID
	return nil
}

func validateAndSetParamsForUpdate(params *RequestParams, httpReq *utilHttp.Request) error {
	val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSID)
	addressID, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("Id is missing or not a number")
	}
	params.QueryParams.AddressId = addressID
	val = httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSTYPE)
	if val == "" {
		return errors.New("Address Type is missing")
	}
	addressType, _ := validateAddressType(val)
	if addressType == appconstant.ALL || addressType == appconstant.OTHER {
		return errors.New("Address Type can be only be billing or shipping")
	}
	params.QueryParams.AddressType = addressType
	return nil
}
