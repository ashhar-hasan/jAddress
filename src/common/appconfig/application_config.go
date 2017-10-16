package appconfig

import (
	"errors"
	"fmt"
	"os"

	"github.com/jabong/florest-core/src/common/config"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/components/cache"
	"github.com/jabong/florest-core/src/components/sqldb"
)

type AddressServiceConfig struct {
	MySqlConfig             *MySqlConfig             `json:"MySql",omitempty`
	EncryptionServiceConfig *EncryptionServiceConfig `json:"EncryptionService,omitempty"`
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

// MapEnvVariables -> Returns map of config values to be replaced by environment variables
func MapEnvVariables() map[string]string {
	overrideVar := make(map[string]string)
	// mysql master configuration
	overrideVar["ApplicationConfig.MySql.Master.Username"] = "MYSQL_MASTER_USERNAME"
	overrideVar["ApplicationConfig.MySql.Master.Password"] = "MYSQL_MASTER_PASSWORD"
	overrideVar["ApplicationConfig.MySql.Master.Host"] = "MYSQL_MASTER_HOST"
	overrideVar["ApplicationConfig.MySql.Master.Port"] = "MYSQL_MASTER_PORT"
	overrideVar["ApplicationConfig.MySql.Master.Dbname"] = "MYSQL_MASTER_DBNAME"
	overrideVar["ApplicationConfig.MySql.Master.MaxOpenCon"] = "MYSQL_MASTER_MAX_OPEN_CONN"

	// mysql slave configuration
	overrideVar["ApplicationConfig.MySql.Slave.Username"] = "MYSQL_SLAVE_USERNAME"
	overrideVar["ApplicationConfig.MySql.Slave.Password"] = "MYSQL_SLAVE_PASSWORD"
	overrideVar["ApplicationConfig.MySql.Slave.Host"] = "MYSQL_SLAVE_HOST"
	overrideVar["ApplicationConfig.MySql.Slave.Port"] = "MYSQL_SLAVE_PORT"
	overrideVar["ApplicationConfig.MySql.Slave.Dbname"] = "MYSQL_SLAVE_DBNAME"
	overrideVar["ApplicationConfig.MySql.Slave.MaxOpenCon"] = "MYSQL_SLAVE_MAX_OPEN_CONN"

	overrideVar["ApplicationConfig.EncryptionService.ReqTimeout"] = "ENCRYPTION_SERVICE_REQ_TIMEOUT"
	overrideVar["ApplicationConfig.EncryptionService.Endpoint"] = "ENCRYPTION_SERVICE_ENDPOINT"
	overrideVar["ApplicationConfig.EncryptionService.EndpointDecrypt"] = "ENCRYPTION_SERVICE_ENDPOINT_DECRYPT"
	overrideVar["ApplicationConfig.EncryptionService.Host"] = "ENCRYPTION_SERVICE_HOST"

	overrideVar["ApplicationConfig.Cache.Redis.ConnStr"] = "REDIS_CONN_STR"
	overrideVar["ApplicationConfig.Cache.Redis.Cluster"] = "IS_CLUSTER"

	checkEnv(overrideVar)
	return overrideVar
}

// checkEnv -> Checks environment variable availability in map, deletes entry if doesn't exist.
func checkEnv(override map[string]string) {
	for key, value := range override {
		if os.Getenv(value) == "" {
			delete(override, key)
		}
	}
}
