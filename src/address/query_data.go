package address

import (
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
)

type RequestParams struct {
	RequestId      string
	QueryParams    QueryParams
	Buckets        map[string]string
	RequestContext utilHttp.RequestContext
}

type QueryParams struct {
	Limit       int
	Offset      int
	AddressType string
	AddressId   int
	Postcode    int
	Address     AddressRequest
}
