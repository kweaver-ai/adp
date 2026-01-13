package common

import (
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"
	"os"
	"sync"
	"time"
	"vega-gateway-pro/version"
)

// server配置项
type ServerSetting struct {
	RunMode  string `mapstructure:"runMode"`
	HttpPort int    `mapstructure:"httpPort"`
	Language string `mapstructure:"language"`
}

// 协程池配置项
type PoolSetting struct {
	QueryPoolSize int `mapstructure:"queryPoolSize"`
}

// 查询配置项
type QuerySetting struct {
	CleanIntervalTime time.Duration `mapstructure:"cleanIntervalTime"`
	MaxIntervalTime   time.Duration `mapstructure:"maxIntervalTime"`
	DataQuerySize     int           `mapstructure:"dataQuerySize"`
	DataCacheSize     int           `mapstructure:"dataCacheSize"`
}

type KafkaSetting struct {
	BatchSize                     int `mapstructure:"batchSize"`
	SessionTimeoutMs              int `mapstructure:"sessionTimeoutMs"`
	FetchWaitMaxMs                int `mapstructure:"fetchWaitMaxMs"`
	SocketTimeoutMs               int `mapstructure:"socketTimeoutMs"`
	MaxPollIntervalMs             int `mapstructure:"maxPollIntervalMs"`
	HeartbeatIntervalMs           int `mapstructure:"heartbeatIntervalMs"`
	Retries                       int `mapstructure:"retries"`
	RetryBackoffMs                int `mapstructure:"retryBackoffMs"`
	TransactionTimeoutMs          int `mapstructure:"transactionTimeoutMs"`
	ProduceFlushTimeoutMs         int `mapstructure:"produceFlushTimeoutMs"`
	AdminClientRequestTimeoutMs   int `mapstructure:"adminClientRequestTimeoutMs"`
	AdminClientOperationTimeoutMs int `mapstructure:"adminClientOperationTimeoutMs"`
	QueryOffsetTimeoutMs          int `mapstructure:"queryOffsetTimeoutMs"`
	TransactionOperationTimeoutMs int `mapstructure:"transactionOperationTimeoutMs"`
}

type RSASetting struct {
	PrivateKeyPath string `mapstructure:"privateKeyPath"`
}

// app配置项
type AppSetting struct {
	ServerSetting        ServerSetting             `mapstructure:"server"`
	LogSetting           logger.LogSetting         `mapstructure:"log"`
	PoolSetting          PoolSetting               `mapstructure:"pool"`
	QuerySetting         QuerySetting              `mapstructure:"query"`
	KafkaSetting         KafkaSetting              `mapstructure:"kafka"`
	ObservabilitySetting o11y.ObservabilitySetting `mapstructure:"observability"`
	DepServices          map[string]map[string]any `mapstructure:"depServices"`
	RSASetting           RSASetting                `mapstructure:"rsa"`

	HydraAdminSetting rest.HydraAdminSetting

	DataConnectionUrl           string
	VegaCalculateCoordinatorUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "vega-gateway-pro-config"
	configType string = "yaml"

	hydraAdminServiceName         string = "hydra-admin"
	vegaDataConnectionServiceName string = "data-connection"
	vegaCalculateCoordinatorName  string = "vega-calculate-coordinator"

	DATA_BASE_NAME string = "adp"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once

	// 当前系统时区
	APP_LOCATION *time.Location
)

// NewSetting 读取服务配置
func NewSetting() *AppSetting {
	settingOnce.Do(func() {
		appSetting = &AppSetting{}
		vp = viper.New()
		initSetting(vp)
	})

	return appSetting
}

// 初始化配置
func initSetting(vp *viper.Viper) {
	logger.Infof("Init Setting From File %s%s.%s", configPath, configName, configType)

	vp.AddConfigPath(configPath)
	vp.SetConfigName(configName)
	vp.SetConfigType(configType)

	loadSetting(vp)

	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		logger.Infof("Config file changed:%s", e)
		loadSetting(vp)
	})
}

// 读取配置文件
func loadSetting(vp *viper.Viper) {
	logger.Infof("Load Setting File %s%s.%s", configPath, configName, configType)

	if err := vp.ReadInConfig(); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	if err := vp.Unmarshal(appSetting); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	// 加载时区
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		loc = time.Local
		logger.Warnf("WARNING: Failed to load timezone from env, using Local[%v] as default. Error: %v\n", time.Local, err)
	}
	APP_LOCATION = loc

	SetLogSetting(appSetting.LogSetting)

	SetHydraAdminSetting()

	SetDataConnectionSetting()

	SetVegaCalculateCoordinatorSetting()

	serverInfo := o11y.ServerInfo{
		ServerName:    version.ServerName,
		ServerVersion: version.ServerVersion,
		Language:      version.LanguageGo,
		GoVersion:     version.GoVersion,
		GoArch:        version.GoArch,
	}
	logger.Infof("ServerName: %s, ServerVersion: %s, Language: %s, GoVersion: %s, GoArch: %s, POD_NAME: %s",
		version.ServerName, version.ServerVersion, version.LanguageGo,
		version.GoVersion, version.GoArch, o11y.POD_NAME)

	o11y.Init(serverInfo, appSetting.ObservabilitySetting)

	s, _ := sonic.MarshalString(appSetting)
	logger.Debug(s)
}

func SetHydraAdminSetting() {
	setting, ok := appSetting.DepServices[hydraAdminServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", hydraAdminServiceName)
	}
	appSetting.HydraAdminSetting = rest.HydraAdminSetting{
		HydraAdminProcotol: setting["protocol"].(string),
		HydraAdminHost:     setting["host"].(string),
		HydraAdminPort:     setting["port"].(int),
	}
}

func SetDataConnectionSetting() {
	setting, ok := appSetting.DepServices[vegaDataConnectionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaDataConnectionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataConnectionUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetVegaCalculateCoordinatorSetting() {
	setting, ok := appSetting.DepServices[vegaCalculateCoordinatorName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaCalculateCoordinatorName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaCalculateCoordinatorUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}
