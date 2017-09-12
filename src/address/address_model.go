package address

import (
	"common/appconstant"
	"errors"
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
	if customerId == "" {
		return nil, errors.New("CustomerID not present")
	}

	sql := `SELECT DISTINCT(CAST(ca.id_customer_address as SIGNED INT)) as id,ca.first_name, ca.last_name, CAST(ca.phone as CHAR), CAST(IFNULL(ca.alternate_phone, "") as CHAR), ca.address1, ca.address2, ca.city, CAST(ca.is_default_billing as SIGNED INT), CAST(ca.is_default_shipping as SIGNED INT), r.name AS region, CAST(r.id_customer_address_region as SIGNED INT), CAST(postcode as SIGNED INT), CAST(country.id_country as SIGNED INT) as country, CAST(adi.sms_opt as SIGNED INT), CAST(IFNULL(ca.address_type, 0) as SIGNED INT)
            FROM customer_address ca JOIN country ON fk_country = id_country 
            LEFT JOIN customer_address_region r ON fk_customer_address_region=r.id_customer_address_region
            LEFT JOIN customer_additional_info adi ON adi.fk_customer=ca.fk_customer
            WHERE ca.fk_customer=` + customerId

	if addressId != "" {
		sql = sql + ` AND id_customer_address = ` + addressId
	}
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "SelectAddressSql", Value: sql})

	rows, err := db.Query(sql)
	e := err.(*sqldb.SDBError)
	if e != nil {
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address table |%s|%s|%s", appconstant.MYSQLError, e.Error(), "customer_address"))
		fmt.Println("Mysql Error while getting data from customer_address table", e.Error())
		return nil, e
	}

	addresses := make([]AddressResponse, 0)
	encryptedFields := make([]EncryptedFields, 0)

	for rows.Next() {
		var (
			fname, lname, address1, address2, city, region, phone, altPhone, smsOpt []byte
			id, isBilling, isShipping, customerAddressRegionId, country             uint32
			postcode, isOffice                                                      int
		)
		ad := AddressResponse{}
		encFields := EncryptedFields{}

		err = rows.Scan(&id, &fname, &lname, &phone, &altPhone, &address1, &address2, &city, &isBilling, &isShipping, &region, &customerAddressRegionId, &postcode, &country, &smsOpt, &isOffice)
		if err != nil {
			logger.Warning(fmt.Println("Mysql Row Error while getting row from customer_address table", err))
			continue
		}

		ad.Id = id
		ad.FirstName = string(fname)
		ad.LastName = string(lname)
		ad.Address1 = string(address1)
		ad.Address2 = string(address2)
		ad.City = string(city)
		ad.RegionName = string(region)
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
		encFields.EncryptedPhone = string(phone)
		encFields.EncryptedAlternatePhone = string(altPhone)
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

func updateAddressInDb(params *RequestParams, debugInfo *Debug) (err error) {
	db, err := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_model-updateAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_model-updateAddress"})
	}()

	rc := params.RequestContext
	userId := rc.UserID
	a := params.QueryParams.Address
	var updateTypeField, query string
	updateTypeField = getAddressTypeSql(a.AddressType)
	if params.QueryParams.Address.Req == appconstant.UpdateType {
		query = `UPDATE customer_address SET ` + updateTypeField + ` WHERE fk_customer = ? and id_customer_address= ?`
		logger.Info(fmt.Sprintf("Update Address Type query: %s", query), rc)
	} else {
		sql := `UPDATE customer_address SET first_name = '%s', address1 = '%s', phone = '%s', city = '%s', postcode = '%d', fk_customer_address_region = '%d', fk_country = '%d' , address_type = '%d'`
		if updateTypeField != "" {
			sql = sql + `, ` + updateTypeField
		}
		if a.LastName != "" {
			sql = sql + `, last_name = '` + a.LastName + `'`
		}
		if a.Address2 != "" {
			sql = sql + `, address2 = '` + a.Address2 + `'`
		}
		if a.AlternatePhone != 0 {
			sql = sql + `, alternate_phone = '` + a.EncryptedAlternatePhone + `'`
		}

		sql = sql + ` WHERE fk_customer = ? and id_customer_address= ?` // + fmt.Sprintf("%d", uint32(a.Id))

		customerAddressRegion, countryId, err := getRegionId(a.AddressRegion, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("Error while getting Region Info of the user"), rc)
		}
		query = fmt.Sprintf(sql, a.FirstName, a.Address1, a.EncryptedPhone, a.City, a.PostCode, customerAddressRegion, countryId, a.IsOffice)
		logger.Info(fmt.Sprintf("Update Address query: %s", query), rc)
	}

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "updateAddressInDb:Sql", Value: query})

	var err1, err2 error
	txObj, terr := db.GetTxnObj()
	if terr == nil {
		_, err1 = txObj.Exec(query, userId, a.Id)
		if err1 != nil {
			logger.Error(fmt.Sprintf("Error while updating user address |%s|%s|%s", appconstant.MYSQLError, err1.Error(), "customer_address"), rc)
		}

		if params.QueryParams.Address.Req != appconstant.UpdateType {
			updateSmsOptSql := getUpdateSmsOptOfUserQuery()
			_, err2 = txObj.Exec(updateSmsOptSql, a.SmsOpt, userId)
			if err2 != nil {
				logger.Error(fmt.Sprintf("Error while updating customer_additional_info for sms_opt |%s|%s", appconstant.MYSQLError, err2.Error()), rc)
			}
		}

		if err1 != nil || err2 != nil {
			txObj.Rollback()
			key := GetAddressListCacheKey(userId)
			invalidateCache(key)
			//updateAddressListInCache(params, fmt.Sprintf("%s",a.Id), debugInfo) //update address list in cache
		}
		err = txObj.Commit()
		if err != nil {
			txObj.Rollback()
			debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "Update::CommitTransactionError:", Value: err.Error()})
			return err
		}
	} else {
		logger.Error(fmt.Sprintf("Transaction Error:: Error while updating user address |%s|%+v", appconstant.MYSQLError, terr), rc)
	}
	return nil
}

func getAddressTypeSql(ty string) string {
	var updateTypeField string
	if ty == appconstant.Billing {
		updateTypeField = ` is_default_billing = 1, is_default_shipping = 0`
	} else if ty == appconstant.Shipping {
		updateTypeField = ` is_default_shipping = 1, is_default_billing = 0`
	}
	return updateTypeField
}

func getRegionId(regionId uint32, debug *Debug) (id uint32, countryId uint32, err error) {

	db, err := sqldb.Get("mysdb")

	sql := `Select CAST(id_customer_address_region as SIGNED INT), CAST(fk_country as SIGNED INT) from customer_address_region where id_customer_address_region = ?` //+ fmt.Sprintf("%d", regionId)

	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "GetRegionSql", Value: sql})
	rows, err := db.Query(sql, regionId)

	if err != nil {
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "GetRegionSql;Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address_region  |%s|%s|%s", appconstant.MYSQLError, err.Error(), "customer_address_region"))
		return 0, 0, err
	}
	var rid uint32
	flag := false
	for rows.Next() {
		flag = true
		err = rows.Scan(&rid, &countryId)
		if err != nil {
			logger.Error(fmt.Sprintf("Mysql Row Error while getting row from customer_address_region table", err))
			return 0, 0, err
		}
	}
	if flag == false {
		return 0, 0, errors.New("Invalid address region Id is given")
	}
	return rid, countryId, nil
}

func getUpdateSmsOptOfUserQuery() string {
	sql := `UPDATE customer_additional_info SET sms_opt=? WHERE fk_customer=?`
	return sql
}
