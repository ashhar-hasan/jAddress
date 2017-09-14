package address

import (
	"common/appconstant"
	"fmt"

	constants "github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type DataEncryptor struct {
	id string
}

func (n *DataEncryptor) SetID(id string) {
	n.id = id
}

func (n DataEncryptor) GetID() (id string, err error) {
	return n.id, nil
}

func (a DataEncryptor) Name() string {
	return "DataEncryptor"
}

func (a DataEncryptor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("DataEncryptor")

	defer func() {
		prof.EndProfileWithMetric([]string{"DataEncryptor_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("DataEncryptor_rc")
	io.ExecContext.SetDebugMsg("Data Encryptor", "Data Encryptor-Execute")
	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("DataEncryptor. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	var phoneStr []string
	str := fmt.Sprintf("%d", params.QueryParams.Address.Phone)
	phoneStr = append(phoneStr, str)

	str = fmt.Sprintf("%d", params.QueryParams.Address.AlternatePhone)
	if str != "0" {
		phoneStr = append(phoneStr, str)
	}

	debugInfo := new(Debug)
	res, err := encryptServiceObj.EncryptData(phoneStr, debugInfo)
	if err != nil {
		logger.Error("PhoneEncryption: Data Encryption Error", err)
	}
	data, err := getDataFromServiceResponse(res)
	if err != nil {
		logger.Error("PhoneEncryption: Error while parsing Encryption Service Response", rc)
		return io, &constants.AppError{Code: constants.ResourceErrorCode, Message: "DataEncryptor: Error while parsing Encryption Service Response"}
	}

	var enPh, enAltPh string
	for i, v := range data {
		if i == 0 {
			enPh = v
		} else {
			enAltPh = v
		}
	}
	params.QueryParams.Address.EncryptedPhone = enPh
	params.QueryParams.Address.EncryptedAlternatePhone = enAltPh

	return io, nil
}
