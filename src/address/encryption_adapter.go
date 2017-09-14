package address

import (
	"common/appconstant"
	"errors"
	"fmt"
	"strconv"
	"time"

	logger "github.com/jabong/florest-core/src/common/logger"
	utilhttp "github.com/jabong/florest-core/src/common/utils/http"
)

type EncryptionService struct {
	Host           string
	RequestTimeOut string
}

func InitEncryptionService(host string, timeout string) (ret *EncryptionService, err error) {
	ret = new(EncryptionService)

	if host == "" || timeout == "" {
		return ret, errors.New("encription service configuration missing")
	}
	ret.Host = host
	ret.RequestTimeOut = timeout
	return ret, err
}

//EncryptData encrypt a string using the encryption service
func (obj *EncryptionService) EncryptData(data []string, debugInfo *Debug) (body []byte, err error) {
	reqURL := obj.Host + appconstant.ENCRYPT_ENDPOINT
	reqURL = urlEncode(reqURL, data)
	Timeout, err := strconv.Atoi(obj.RequestTimeOut)

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "EncryptionUrl", Value: reqURL})
	response, err := utilhttp.Get(reqURL, nil, time.Duration(Timeout)*time.Millisecond)
	body = response.Body
	if err != nil {
		logger.Debug(fmt.Sprintf("Encryption Err. InputData: %s, Service Response %s, Error: %s", data, body, err))
	}
	return body, err
}

//DecryptData decrypt a string using the decryption service
func (obj *EncryptionService) DecryptData(data []string, debugInfo *Debug) (body []byte, err error) {
	reqURL := obj.Host + appconstant.DECRYPT_ENDPOINT
	reqURL = urlEncode(reqURL, data)
	Timeout, err := strconv.Atoi(obj.RequestTimeOut)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "DecryptionUrl", Value: reqURL})
	response, err := utilhttp.Get(reqURL, nil, time.Duration(Timeout)*time.Millisecond)
	body = response.Body
	if err != nil {
		logger.Debug(fmt.Sprintf("Decryption Err. InputData: %s, Service Response %s, Error: %s", data, body, err))
	}
	return body, err
}
