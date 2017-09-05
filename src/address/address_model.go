package address

import (
	"common/appconstant"
	"fmt"
	"strconv"

	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	"github.com/jabong/florest-core/src/components/sqldb"
)

func getAddressList(params *RequestParams, addressId string, debug *Debug) (address []AddressResponse, err error) {
	db, err := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_model-getAddressList")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_model-getAddressList"})
	}()

	rc := params.RequestContext
	customerId := rc.UserID
	if customerId == nil {
		return nil, "CustomerID not present"
	}

	sql := `SELECT CAST(ca.id_customer_address as SIGNED INT) as id,ca.first_name, ca.last_name, CAST(ca.phone as CHAR), CAST(IFNULL(ca.alternate_phone, "") as CHAR), ca.address1, ca.address2, ca.city, CAST(ca.is_default_billing as SIGNED INT), CAST(ca.is_default_shipping as SIGNED INT), r.name AS region, CAST(r.id_customer_address_region as SIGNED INT), CAST(postcode as SIGNED INT), CAST(country.id_country as SIGNED INT) as country, CAST(adi.sms_opt as SIGNED INT), CAST(IFNULL(ca.address_type, 0) as SIGNED INT)
			FROM customer_address ca JOIN country ON fk_country = id_country 
			LEFT JOIN customer_address_region r ON fk_customer_address_region=r.id_customer_address_region
			LEFT JOIN customer_additional_info adi ON adi.fk_customer=ca.fk_customer
			WHERE ca.fk_customer=` + customerId

	if addressId != "" {
		sql = sql + ` AND id_customer_address = ` + addressId
	}
	limits := strconv.Itoa(params.QueryParams.Limit)
	offsets := strconv.Itoa(params.QueryParams.Offset)
	sql = sql + `limit ` + limits + ` offset ` + offsets
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "SelectAddressSql", Value: sql})

	rows, err := db.Query(sql)
	if err != nil {
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address table |%s|%s|%s", appconstant.MYSQLError, err.Error(), "customer_address"))
		fmt.Println("Mysql Error while getting data from customer_address table", err)
		return nil, err
	}

	addresses := make([]AddressResponse, 0)
	encryptedFields := make([]EncryptedFields, 0)

	for rows.Next() {
		var (
			fname, lname, address1, address2, city, region, phone, altPhone string
			id, isBilling, isShipping, customerAddressRegionId, country     uint32
			postcode, isOffice                                              int
			smsOpt                                                          []byte
		)
		ad := AddressResponse{}
		encFields := EncryptedFields{}

		err = rows.Scan(&id, &fname, &lname, &phone, &altPhone, &address1, &address2, &city, &isBilling, &isShipping, &region, &customerAddressRegionId, &postcode, &country, &smsOpt, &isOffice)
		if err != nil {
			logger.Warning(fmt.Println("Mysql Row Error while getting row from customer_address table", err))
			continue
		}

		ad.Id = id
		ad.FirstName = fname
		ad.LastName = lname
		ad.Address1 = address1
		ad.Address2 = address2
		ad.City = city
		ad.RegionName = region
		ad.AddressRegion = customerAddressRegionId
		ad.PostCode = postcode
		ad.Country = country
		ad.IsOffice = isOffice

		if isBilling == 0 && isShipping == 0 {
			ad.AddressType = appconstant.Other
		} else if isBilling == 1 {
			ad.AddressType = appconstant.Billing
		} else if isShipping == 1 {
			ad.AddressType = appconstant.Shipping
		} else {
			ad.AddressType = appconstant.Other
		}

		sms, err := strconv.Atoi(string(smsOpt))
		if err != nil {
			logger.Warning(fmt.Println("Can not convert string to int while getting customer_address list column 'sms_opt'", err))
		}
		ad.SmsOpt = sms

		encFields.Id = id
		encFields.EncryptedPhone = phone
		encFields.EncryptedAlternatePhone = altPhone
		encryptedFields = append(encryptedFields, encFields)

		addresses = append(addresses, ad)
	}

	if len(encryptedFields) != 0 {
		res := decryptEncryptedFields(encryptedFields, params, debug)
		mergeDecryptedFieldsWithAddressResult(res, &addresses)
	}

	if addressId == "" {
		if len(addresses) != 0 {
			err = saveDataInCache(customerId, "address", addresses)
			if err != nil {
				logger.Error(fmt.Println("getAddressList:Could not update addressList in cache. ", err.Error()))
			}
		}
	}
	return addresses, nil
}
