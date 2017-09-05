package appconfig

import (
	"errors"
	"fmt"

	"github.com/jabong/florest-core/src/common/config"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/components/cache"
	"github.com/jabong/florest-core/src/components/sqldb"
)

type AddressServiceConfig struct {
	MySqlConfig             *MySqlConfig             `json:"MySqlConfig",omitempty`
	EncryptionServiceConfig *EncryptionServiceConfig `json:"EncryptionServiceConfig,omitempty"`
	Cache                   *CacheConf               `json:"Cache,omitempty"`
}

type MySqlConfig struct {
	MySqlMaster *sqldb.SDBConfig `json:"Master,omitempty"`
	MySqlSlave  *sqldb.SDBConfig `json:"Slave,omitempty"`
}

type EncryptionServiceConfig struct {
	ReqTimeout      string
	Endpoint        string
	EndpointDecrypt string
	Host            string
}

type CacheConf struct {
	Redis        *cache.Config `json:"Redis,omitempty"`
	RedisCluster *cache.Config `json:"RedisCluster,omitempty"`
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
