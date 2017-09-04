package appconfig

import (
	"errors"
	"fmt"

	"github.com/jabong/florest-core/src/common/config"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/components/sqldb"
)

type AddressServiceConfig struct {
	MySqlConfig             *MySqlConfig
	EncryptionServiceConfig *EncryptionServiceConfig
}

type MySqlConfig struct {
	MySqlMaster *sqldb.SDBConfig `json:"Master"`
	MySqlSlave  *sqldb.SDBConfig `json:"Slave"`
}

type EncryptionServiceConfig struct {
	ReqTimeout      string
	Endpoint        string
	EndpointDecrypt string
	Host            string
}

func GetAddressServiceConfig() (*AddressServiceConfig, error) {
	c := config.GlobalAppConfig.ApplicationConfig
	appConfig, ok := c.(*AddressServiceConfig)
	if !ok {
		msg := fmt.Sprintf("Example APP Config Not correct %+v", c)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	return appConfig, nil
}
