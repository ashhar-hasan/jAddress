package address

import (
	"common/appconstant"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	"github.com/jabong/florest-core/src/components/sqldb"
)

func getRegionId(regionId string, debug *Debug) (id string, countryId string, err error) {
	db, err := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("AddressModel#getRegionId")
	defer func() {
		prof.EndProfileWithMetric([]string{"AddressModel#getRegionId"})
	}()

	sql := `SELECT id_customer_address_region, fk_country FROM customer_address_region WHERE id_customer_address_region = ?`
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "GetRegionSql", Value: sql})
	rows, err := db.Query(sql, regionId)
	if err.(*sqldb.SDBError) != nil {
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "GetRegionSql;Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address_region  |%s|%s|%s", appconstant.MYSQL_ERROR, err.Error(), "customer_address_region"))
		return "", "", err
	}
	var rid string
	flag := false
	for rows.Next() {
		flag = true
		err = rows.Scan(&rid, &countryId)
		if err != nil {
			logger.Error(fmt.Sprintf("Mysql Row Error while getting row from customer_address_region table %s", err))
			return "", "", err
		}
	}
	if flag == false {
		return "", "", errors.New("Invalid address region Id is given")
	}
	return rid, countryId, nil
}

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

	sql := `SELECT DISTINCT(ca.id_customer_address) as id,ca.first_name, ca.last_name, ca.phone, IFNULL(ca.alternate_phone, ""), ca.address1, ca.address2, ca.city, ca.is_default_billing, ca.is_default_shipping, r.name AS region, r.id_customer_address_region, postcode, country.id_country as country, adi.sms_opt, IFNULL(ca.address_type, 0)
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
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address table |%s|%s|%s", appconstant.MYSQL_ERROR, e.Error(), "customer_address"))
		fmt.Println("Mysql Error while getting data from customer_address table", e.Error())
		return nil, e
	}

	addresses := make([]AddressResponse, 0)
	encryptedFields := make([]EncryptedFields, 0)

	for rows.Next() {
		var (
			fname, lname, address1, address2, city, region, phone, altPhone, smsOpt []byte
			id, isBilling, isShipping, customerAddressRegionId, country             []byte
			postcode, isOffice                                                      []byte
		)
		ad := AddressResponse{}
		encFields := EncryptedFields{}

		err = rows.Scan(&id, &fname, &lname, &phone, &altPhone, &address1, &address2, &city, &isBilling, &isShipping, &region, &customerAddressRegionId, &postcode, &country, &smsOpt, &isOffice)
		if err != nil {
			logger.Warning(fmt.Println("Mysql Row Error while getting row from customer_address table", err))
			continue
		}

		ad.Id = string(id)
		ad.FirstName = string(fname)
		ad.LastName = string(lname)
		ad.Address1 = string(address1)
		ad.Address2 = string(address2)
		ad.City = string(city)
		ad.RegionName = string(region)
		ad.AddressRegion = string(customerAddressRegionId)
		ad.PostCode = string(postcode)
		ad.Country = string(country)
		ad.IsOffice = string(isOffice)

		if string(isBilling) == "0" && string(isShipping) == "0" {
			ad.AddressType = appconstant.OTHER
		} else if string(isBilling) == "1" {
			ad.AddressType = appconstant.BILLING
		} else if string(isShipping) == "1" {
			ad.AddressType = appconstant.SHIPPING
		} else {
			ad.AddressType = appconstant.OTHER
		}

		ad.SmsOpt = string(smsOpt)

		encFields.Id = string(id)
		encFields.EncryptedPhone = string(phone)
		encFields.EncryptedAlternatePhone = string(altPhone)
		encryptedFields = append(encryptedFields, encFields)

		addresses = append(addresses, ad)
	}

	if len(encryptedFields) != 0 {
		res, err := decryptEncryptedFields(encryptedFields, params, debug)
		if err != nil {
			logger.Error("PhoneDecryption: Error while parsing Decryption Service Response")
			return nil, &constants.AppError{Code: constants.ResourceErrorCode, Message: "DecryptEncryptedFields: Error while parsing Decryption Service Response"}
		}
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

func addAddress(userID string, a AddressRequest, debug *Debug) (int64, error) {
	db, _ := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("AddressModel#addAddress")

	defer func() {
		prof.EndProfileWithMetric([]string{"AddressModel#addAddress"})
	}()

	sql := `INSERT INTO customer_address SET first_name = ? , last_name = ? , address1 = ? , address2 = ?, phone = ?, alternate_phone = ?, city = ?, postcode = ?, fk_customer_address_region = ?, fk_country = ?, fk_customer = ?, address_type = ?, created_at = ?, validation_flag= ?`

	var addressTypeField string
	if a.AddressType == appconstant.BILLING {
		addressTypeField = `, is_default_billing = 1`
	} else if a.AddressType == appconstant.SHIPPING {
		addressTypeField = `, is_default_shipping = 1`
	}
	sql = sql + addressTypeField

	customerAddressRegion, countryID, err1 := getRegionId(a.AddressRegion, debug)
	if err1 != nil {
		logger.Error(fmt.Sprintf("|%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"))
		return 0, err1
	}
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "InsertAddressSql", Value: sql})

	// start and commit one txn: insert one row in table
	txObj, terr := db.GetTxnObj()
	if terr != nil {
		logger.Error(fmt.Sprintf("|%s|%s|%s", appconstant.MYSQL_ERROR, terr.Error(), "customer_address"))
		return 0, terr
	}
	validationFlag := validateAddress(a.Address1 + a.Address2)
	rows, err1 := txObj.Exec(sql, a.FirstName, a.LastName, a.Address1, a.Address2, a.EncryptedPhone, a.EncryptedAlternatePhone, a.City, a.PostCode, customerAddressRegion, countryID, userID, a.IsOffice, time.Now().Format("2006-01-02 15:04:05"), validationFlag)
	if err1 != nil {
		txObj.Rollback()
		logger.Error(fmt.Sprintf("|%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"))
		return 0, err1
	}
	err1 = txObj.Commit()
	if err1 != nil {
		txObj.Rollback()
		logger.Error(fmt.Sprintf("AddAddress::CommitError::|%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"))
		return 0, err1
	}
	id, err1 := rows.LastInsertId()
	if err1 != nil {
		logger.Error(fmt.Sprintf("Mysql Error while retrieving last inserted row into customer_address table |%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"))
		return 0, err1
	}
	logger.Info(fmt.Sprintf("Last Insert Id %s", id))

	return id, nil
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
	if params.QueryParams.Address.Req == appconstant.UPDATE_TYPE {
		query = `UPDATE customer_address SET ` + updateTypeField + ` WHERE fk_customer = ? and id_customer_address= ?`
		logger.Info(fmt.Sprintf("Update Address Type query: %s", query), rc)
	} else {
		sql := `UPDATE customer_address SET first_name = '%s', address1 = '%s', phone = '%s', city = '%s', postcode = '%d', fk_customer_address_region = '%d', fk_country = '%d' , address_type = '%d', validation_flag = '%s'`
		if updateTypeField != "" {
			sql = sql + `, ` + updateTypeField
		}
		if a.LastName != "" {
			sql = sql + `, last_name = '` + a.LastName + `'`
		}
		if a.Address2 != "" {
			sql = sql + `, address2 = '` + a.Address2 + `'`
		}
		if a.AlternatePhone != "" {
			sql = sql + `, alternate_phone = '` + a.EncryptedAlternatePhone + `'`
		}

		sql = sql + ` WHERE fk_customer = ? and id_customer_address= ?` // + fmt.Sprintf("%d", uint32(a.Id))

		customerAddressRegion, countryId, err := getRegionId(a.AddressRegion, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("Error while getting Region Info of the user"), rc)
		}
		validationFlag := validateAddress(a.Address1 + a.Address2)
		query = fmt.Sprintf(sql, a.FirstName, a.Address1, a.EncryptedPhone, a.City, a.PostCode, customerAddressRegion, countryId, a.IsOffice, validationFlag)
		logger.Info(fmt.Sprintf("Update Address query: %s", query), rc)
	}
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "updateAddressInDb:Sql", Value: query})

	var err1, err2 error
	txObj, terr := db.GetTxnObj()
	if terr == nil {
		_, err1 = txObj.Exec(query, userId, a.Id)
		if err1 != nil {
			logger.Error(fmt.Sprintf("Error while updating user address |%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"), rc)
		}

		if params.QueryParams.Address.Req != appconstant.UPDATE_TYPE {
			updateSmsOptSql := getUpdateSmsOptOfUserQuery()
			_, err2 = txObj.Exec(updateSmsOptSql, a.SmsOpt, userId)
			if err2 != nil {
				logger.Error(fmt.Sprintf("Error while updating customer_additional_info for sms_opt |%s|%s", appconstant.MYSQL_ERROR, err2.Error()), rc)
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
		logger.Error(fmt.Sprintf("Transaction Error:: Error while updating user address |%s|%+v", appconstant.MYSQL_ERROR, terr), rc)
	}
	return nil
}

func deleteAddress(params *RequestParams, cacheErr error, debugInfo *Debug, e chan error) (err error) {
	db, err := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_model-deleteAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_model-deleteAddress"})
	}()
	rc := params.RequestContext
	userId := rc.UserID
	id := params.QueryParams.AddressId
	addressId := fmt.Sprintf("%d", id)
	sql := `DELETE FROM customer_address WHERE id_customer_address=? AND fk_customer=?`

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "deleteAddress:Sql", Value: sql})

	txObj, _ := db.GetTxnObj()
	deleteResult, err1 := txObj.Exec(sql, addressId, userId)
	if err1 != nil {
		txObj.Rollback()
		logger.Error(fmt.Sprintf("Error while delete user address |%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"), rc)
		e <- err1
	}
	rowsaffected, _ := deleteResult.RowsAffected()
	if rowsaffected == 0 {
		txObj.Rollback()
		addressNotFoundError := errors.New("Address not found")
		e <- addressNotFoundError
		return
	}
	err = txObj.Commit()
	if err != nil {
		txObj.Rollback()
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "Delete::CommitTransactionError:", Value: err.Error()})
		e <- err
		return
	}
	if cacheErr != nil {
		key := GetAddressListCacheKey(userId)
		invalidateCache(key)
	}
	e <- nil
	return
}

func getAddressTypeSql(ty string) string {
	var updateTypeField string
	if ty == appconstant.BILLING {
		updateTypeField = ` is_default_billing = 1, is_default_shipping = 0`
	} else if ty == appconstant.SHIPPING {
		updateTypeField = ` is_default_shipping = 1, is_default_billing = 0`
	}
	return updateTypeField
}

func getUpdateSmsOptOfUserQuery() string {
	sql := `UPDATE customer_additional_info SET sms_opt=? WHERE fk_customer=?`
	return sql
}

func validateAddress(address string) string {

	//To check the same character is not repeated 4 times
	repeatCount := 1
	thresh := 4
	lastChar := ""
	var flag string
	flag = "1"
	for _, r := range address {
		c := string(r)
		if c == lastChar {
			repeatCount++
			if repeatCount == thresh {
				flag = "0"
				break
			}
		} else {
			repeatCount = 1
		}
		lastChar = c
	}

	if flag == "1" {
		count := 0
		for i := 0; i < len(address); i++ {
			if (i == len(address)-1) && count == 0 {
				flag = "0"
				break
			}
			if string(address[i]) == " " {
				count += 1
			}
		}
	}
	specialChars := regexp.MustCompile(`[ \-\,\n]`)
	vowels := regexp.MustCompile(`(?i)[aeiouy]`)
	if flag == "1" && len(vowels.FindStringIndex(address)) == 0 {
		flag = "0"
	}
	if flag == "1" {
		result_array := specialChars.Split(address, -1)
		for _, v := range result_array {
			if len(v) > 20 {
				flag = "0"
				break
			}
		}
	}
	return flag
}
