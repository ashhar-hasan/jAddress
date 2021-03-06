package address

import (
	"common/appconstant"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "GetRegionSql", Value: sql + regionId})
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

func getAddressList(params *RequestParams, addressId string, debug *Debug) (address map[string]*AddressResponse, order []string, err error) {
	db, err := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_model-getAddressList")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_model-getAddressList"})
	}()

	rc := params.RequestContext
	customerId := rc.UserID
	if customerId == "" {
		return nil, nil, errors.New("CustomerID not present")
	}

	sql := `SELECT DISTINCT(ca.id_customer_address) as id,ca.first_name, ca.last_name, ca.phone, IFNULL(ca.alternate_phone, ""), ca.address1, ca.address2, ca.city, ca.is_default_billing, ca.is_default_shipping, ca.fk_customer, ca.created_at, ca.updated_at, r.name AS region, r.id_customer_address_region, postcode, country.id_country as country, adi.sms_opt, IFNULL(ca.address_type, 0)
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
		return nil, nil, e
	}

	addresses := make(map[string]*AddressResponse, 0)
	encryptedFields := make([]EncryptedFields, 0)
	for rows.Next() {
		var (
			fname, lname, address1, address2, city, region, phone, altPhone, smsOpt                     []byte
			id, isBilling, isShipping, fkCustomer, customerAddressRegionId, country, postcode, isOffice []byte
			createdAt                                                                                   []byte
			updatedAt                                                                                   time.Time
		)
		encFields := EncryptedFields{}

		err = rows.Scan(&id, &fname, &lname, &phone, &altPhone, &address1, &address2, &city, &isBilling, &isShipping, &fkCustomer, &createdAt, &updatedAt, &region, &customerAddressRegionId, &postcode, &country, &smsOpt, &isOffice)
		if err != nil {
			logger.Warning(fmt.Sprintf("Mysql Row Error while getting row from customer_address table", err))
			continue
		}

		resp := new(AddressResponse)

		index := string(id)
		resp.Id = index
		resp.FirstName = sanitize(string(fname), true)
		resp.LastName = sanitize(string(lname), true)
		resp.Address1 = sanitize(string(address1), false)
		resp.Address2 = sanitize(string(address2), false)
		resp.City = sanitize(string(city), false)
		resp.RegionName = string(region)
		resp.AddressRegion = string(customerAddressRegionId)
		resp.PostCode = string(postcode)
		resp.Country = string(country)
		resp.IsOffice = string(isOffice)
		resp.IsDefaultBilling = string(isBilling)
		resp.IsDefaultShipping = string(isShipping)
		resp.FkCustomer = string(fkCustomer)
		createdAtStr := string(createdAt)
		if createdAtStr == "" {
			resp.CreatedAt = ""
		} else {
			createdAtTime, _ := time.Parse(time.RFC3339, createdAtStr)
			resp.CreatedAt = createdAtTime.Format(appconstant.DATETIME_FORMAT)
		}
		resp.UpdatedAt = updatedAt.Format(appconstant.DATETIME_FORMAT)
		resp.SmsOpt = string(smsOpt)

		encFields.Id = string(id)
		encFields.EncryptedPhone = string(phone)
		encFields.EncryptedAlternatePhone = string(altPhone)
		encryptedFields = append(encryptedFields, encFields)
		addresses[index] = resp
		order = append(order, index)
	}
	if len(encryptedFields) != 0 {
		res, err := decryptEncryptedFields(encryptedFields, params, debug)
		if err != nil {
			logger.Error("PhoneDecryption: Error while parsing Decryption Service Response")
			return nil, nil, &constants.AppError{Code: constants.ResourceErrorCode, Message: "DecryptEncryptedFields: Error while parsing Decryption Service Response"}
		}
		mergeDecryptedFieldsWithAddressResult(res, &addresses)
	}

	if addressId == "" {
		if len(addresses) != 0 {
			err = saveOrderInCache(customerId, order)
			if err != nil {
				logger.Error("getAddressList:Could not update Order in cache. ", err.Error())
			}
			err = saveDataInCache(customerId, addresses)
			if err != nil {
				logger.Error("getAddressList:Could not update addressList in cache. ", err.Error())
			}
		}
	}
	return addresses, order, nil
}

func addAddress(userID string, a AddressRequest, debug *Debug) (int64, error) {
	db, _ := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("AddressModel#addAddress")

	defer func() {
		prof.EndProfileWithMetric([]string{"AddressModel#addAddress"})
	}()

	sql := `INSERT INTO customer_address SET first_name=?, address1=?, phone=?, postcode=?, city=?, fk_customer_address_region=?, fk_country=?, fk_customer=?, created_at=?, validation_flag=?, last_name=?`
	if a.Address2 != "" {
		sql = sql + `, address2='` + a.Address2 + `'`
	}
	if a.AlternatePhone != "" {
		sql = sql + `, alternate_phone='` + a.EncryptedAlternatePhone + `'`
	}
	if a.IsOffice != "" {
		sql = sql + `, address_type='` + a.IsOffice + `'`
	}
	// Check if the user has any other addresses, if not, mark this as default
	flag, err := isFirstAddress(userID, debug)
	if flag == true {
		sql = sql + `, is_default_shipping = 1, is_default_billing = 1`
	} else if err != nil {
		return 0, err
	}

	customerAddressRegion, countryID, err1 := getRegionId(a.AddressRegion, debug)
	if err1 != nil {
		logger.Error(fmt.Sprintf("|%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"))
		return 0, err1
	}
	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "InsertAddressSql", Value: sql + fmt.Sprintf("%+v", a)})

	// start and commit one txn: insert one row in table
	txObj, terr := db.GetTxnObj()
	if terr != nil {
		logger.Error(fmt.Sprintf("|%s|%s|%s", appconstant.MYSQL_ERROR, terr.Error(), "customer_address"))
		return 0, terr
	}
	validationFlag := validateAddress(a.Address1 + a.Address2)
	rows, err1 := txObj.Exec(sql, a.FirstName, a.Address1, a.EncryptedPhone, a.PostCode, a.City, customerAddressRegion, countryID, userID, time.Now().Format(appconstant.DATETIME_FORMAT), validationFlag, a.LastName)
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
	var query string
	sql := `UPDATE customer_address SET first_name = '%s', address1 = '%s', phone = '%s', city = '%s', postcode = '%s', fk_customer_address_region = '%s', fk_country = '%s' , address_type = '%s', validation_flag = '%s'`
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
	addressId := strconv.Itoa(params.QueryParams.AddressId)
	logger.Info(fmt.Sprintf("Update Address query: %s", query), rc)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "updateAddressInDb:Sql", Value: query + "fk_customer: " + userId + "id_customer_address: " + addressId})

	var err1, err2 error
	txObj, terr := db.GetTxnObj()
	if terr == nil {
		_, err1 = txObj.Exec(query, userId, addressId)
		if err1 != nil {
			logger.Error(fmt.Sprintf("Error while updating user address |%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"), rc)
		}

		updateSmsOptSql := getUpdateSmsOptOfUserQuery()
		_, err2 = txObj.Exec(updateSmsOptSql, a.SmsOpt, userId)
		if err2 != nil {
			logger.Error(fmt.Sprintf("Error while updating customer_additional_info for sms_opt |%s|%s", appconstant.MYSQL_ERROR, err2.Error()), rc)
		}

		if err1 != nil || err2 != nil {
			txObj.Rollback()
			key := GetAddressListCacheKey(userId)
			invalidateCache(key)
			key = GetAddressOrderCacheKey(userId)
			invalidateCache(key)
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
	addressId := strconv.Itoa(params.QueryParams.AddressId)

	sql := `DELETE FROM customer_address WHERE id_customer_address=? AND fk_customer=?`

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "deleteAddress:Sql", Value: sql + "id_customer_address: " + addressId + "fk_customer: " + userId})

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
		key = GetAddressOrderCacheKey(userId)
		invalidateCache(key)
	}
	e <- nil
	return
}

func getAddressTypeSql(ty string) string {
	var updateTypeField string
	if ty == appconstant.BILLING {
		updateTypeField = ` is_default_billing = 1`
	} else if ty == appconstant.SHIPPING {
		updateTypeField = ` is_default_shipping = 1`
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

func updateType(params *RequestParams, debugInfo *Debug, e chan error) {
	db, _ := sqldb.Get("mysdb")
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_model-updateType")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_model-updateType"})
	}()
	rc := params.RequestContext
	userId := rc.UserID
	updateTypeField := getAddressTypeSql(params.QueryParams.AddressType)
	query := `UPDATE customer_address SET ` + updateTypeField + ` WHERE fk_customer = ? and id_customer_address= ?`
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "updateType:Sql", Value: query + "fk_customer: " + userId + "id_customer_address: " + string(params.QueryParams.AddressId)})
	txObj, _ := db.GetTxnObj()
	updateTypeResult, err1 := txObj.Exec(query, userId, params.QueryParams.AddressId)
	if err1 != nil {
		txObj.Rollback()
		key := GetAddressListCacheKey(userId)
		invalidateCache(key)
		key = GetAddressOrderCacheKey(userId)
		invalidateCache(key)
		logger.Error(fmt.Sprintf("Error while updating  address type|%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"), rc)
		e <- err1
	}
	rowsaffected, _ := updateTypeResult.RowsAffected()
	if rowsaffected == 0 {
		txObj.Rollback()
		addressNotFoundError := errors.New("Address not found")
		e <- addressNotFoundError
		return
	}
	err1 = txObj.Commit()
	if err1 != nil {
		txObj.Rollback()
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "UpdateType::CommitTransactionError:", Value: err1.Error()})
		e <- err1
		return
	}

	// Reset defaults for other addresses
	addressTypeSQL := ""
	if params.QueryParams.AddressType == appconstant.BILLING {
		addressTypeSQL = ` is_default_billing = 1`
	} else if params.QueryParams.AddressType == appconstant.SHIPPING {
		addressTypeSQL = ` is_default_shipping = 1`
	}
	query = `SELECT id_customer_address FROM customer_address WHERE ` + addressTypeSQL + ` AND fk_customer = ?`
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "resetDefaultAddress#Sql", Value: query + userId})
	rows, err := db.Query(query, userId)
	if err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "resetDefaultAddress#Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address |%s|%s", appconstant.MYSQL_ERROR, err.Error()))
		e <- err
		return
	}
	var addressID string
	resetAddressTypeSql := ""
	if params.QueryParams.AddressType == appconstant.BILLING {
		resetAddressTypeSql = `is_default_billing = 0`
	} else if params.QueryParams.AddressType == appconstant.SHIPPING {
		resetAddressTypeSql = `is_default_shipping = 0`
	}
	resetQuery := `UPDATE customer_address SET ` + resetAddressTypeSql + ` WHERE id_customer_address = ?`
	for rows.Next() {
		err1 := rows.Scan(&addressID)
		if err1 != nil {
			debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "resetDefaultAddress#Err", Value: err1.Error()})
			logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address |%s|%s", appconstant.MYSQL_ERROR, err1.Error()))
			e <- err1
			return
		}
		txObj, terr := db.GetTxnObj()
		if terr == nil {
			if addressID != strconv.Itoa(params.QueryParams.AddressId) {
				_, err1 = txObj.Exec(resetQuery, addressID)
				if err1 != nil {
					txObj.Rollback()
					key := GetAddressListCacheKey(userId)
					invalidateCache(key)
					key = GetAddressOrderCacheKey(userId)
					invalidateCache(key)
					logger.Error(fmt.Sprintf("Error while updating user address |%s|%s|%s", appconstant.MYSQL_ERROR, err1.Error(), "customer_address"), rc)
				}
			}
		} else {
			logger.Error(fmt.Sprintf("Transaction Error:: Error while updating user address |%s|%s", appconstant.MYSQL_ERROR, terr), rc)
		}
		err1 = txObj.Commit()
		if err != nil {
			txObj.Rollback()
			debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "AddAddress::CommitTransactionError:", Value: err.Error()})
			e <- err
		}
	}

	e <- nil
	return
}

func isFirstAddress(userID string, debug *Debug) (bool, error) {
	// Use cache before using DB
	var e QueryParams
	addressList, _, cacheErr := getAddressListFromCache(userID, e, debug)
	if cacheErr != nil || len(addressList) == 0 {
		db, _ := sqldb.Get("mysdb")
		prof := profiler.NewProfiler()
		prof.StartProfile("AddressModel#isFirstAddress")

		defer func() {
			prof.EndProfileWithMetric([]string{"AddressModel#isFirstAddress"})
		}()

		sql := `SELECT COUNT(id_customer_address) FROM customer_address WHERE fk_customer = ?`
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "isFirstAddressSql", Value: sql + userID})
		rows, err := db.Query(sql, userID)
		if err != nil {
			debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "isFirstAddressSql#Err", Value: err.Error()})
			logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address |%s|%s", appconstant.MYSQL_ERROR, err.Error()))
			return false, err
		}
		count := 1
		if rows.Next() {
			err1 := rows.Scan(&count)
			if err1 != nil {
				debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "isFirstAddressSql#Err", Value: err1.Error()})
				logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address |%s|%s", appconstant.MYSQL_ERROR, err1.Error()))
				return false, err1
			}
		}
		return (count == 0), nil

	}
	return false, nil

}

func checkDefaultAddressInDB(addressID int, userID string, debugInfo *Debug) (int, error) {
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "checkDefaultAddressInDB", Value: "checkDefaultAddressInDB execute"})
	db, _ := sqldb.Get("mysdb")
	sql := "SELECT is_default_shipping,is_default_billing FROM customer_address WHERE id_customer_address=? AND fk_customer=?"
	rows, err := db.Query(sql, addressID, userID)
	if err != nil {
		logger.Error(fmt.Sprintf("Mysql Error while getting data from customer_address table |%s|%s|%s", appconstant.MYSQL_ERROR, err.Error(), "customer_address"))
		return -1, err
	}
	var shipping, billing int

	if rows.Next() {

		err1 := rows.Scan(&shipping, &billing)
		if err1 != nil {
			logger.Warning(fmt.Sprintf("checkDefaultAddressInDB : Mysql Row Error while getting row from customer_address table", err1))
		}
		if billing == 1 {
			return 1, nil
		} else if shipping == 1 {
			return 2, nil
		} else {
			return 0, nil
		}

	}
	return 0, nil
}
