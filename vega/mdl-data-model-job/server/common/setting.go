package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"
)

// server配置项
type ServerSetting struct {
	RunMode              string        `mapstructure:"runMode"`
	HttpPort             int           `mapstructure:"httpPort"`
	Language             string        `mapstructure:"language"`
	ReadTimeOut          time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout         time.Duration `mapstructure:"writeTimeOut"`
	RetryIntervalMs      int           `mapstructure:"retryIntervalMs"`
	FlushMiB             int           `mapstructure:"flushMiB"`
	FlushItems           int           `mapstructure:"flushItems"`
	FlushIntervalSec     int           `mapstructure:"flushIntervalSec"`
	FailureThreshold     int           `mapstructure:"failureThreshold"`
	PackagePoolSize      int           `mapstructure:"packagePoolSize"`
	ConnectCheckTimeout  time.Duration `mapstructure:"connectCheckTimeout"`
	WatchJobsIntervalMin time.Duration `mapstructure:"watchJobsIntervalMin"`
}

// Kafka配置
type KafkaSetting struct {
	AutoOffsetReset               string `mapstructure:"autoOffsetReset"`
	RetentionMs                   int    `mapstructure:"retentionMs"`
	RetentionBytes                int    `mapstructure:"retentionBytes"`
	SessionTimeoutMs              int    `mapstructure:"sessionTimeoutMs"`
	FetchWaitMaxMs                int    `mapstructure:"fetchWaitMaxMs"`
	SocketTimeoutMs               int    `mapstructure:"socketTimeoutMs"`
	MaxPollIntervalMs             int    `mapstructure:"maxPollIntervalMs"`
	HeartbeatIntervalMs           int    `mapstructure:"heartbeatIntervalMs"`
	Retries                       int    `mapstructure:"retries"`
	RetryBackoffMs                int    `mapstructure:"retryBackoffMs"`
	TransactionTimeoutMs          int    `mapstructure:"transactionTimeoutMs"`
	AdminClientRequestTimeoutMs   int    `mapstructure:"adminClientRequestTimeoutMs"`
	AdminClientOperationTimeoutMs int    `mapstructure:"adminClientOperationTimeoutMs"`
}

// app配置项
type AppSetting struct {
	ServerSetting ServerSetting             `mapstructure:"server"`
	LogSetting    logger.LogSetting         `mapstructure:"log"`
	DepServices   map[string]map[string]any `mapstructure:"depServices"`
	KafkaSetting  KafkaSetting              `mapstructure:"kafka"`

	DBSetting         libdb.DBSetting
	MQSetting         libmq.MQSetting
	HydraAdminSetting rest.HydraAdminSetting

	// 索引库 url
	IndexBaseUrl string
	// metric-model url
	MetricTaskUrl string
	// uniquery url
	UniQueryUrl string
	//event-model task url
	EventTaskUrl string
	//event-model url
	EventModelUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "data-model-job-config"
	configType string = "yaml"

	rdsServiceName        string = "rds"
	mqServiceName         string = "mq"
	indexBaseServiceName  string = "index-base"
	dataModelServiceName  string = "data-model"
	uniQueryServiceName   string = "uniquery"
	hydraAdminServiceName string = "hydra-admin"

	DATA_BASE_NAME string = "adp"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once
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

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()

	SetMQSetting()

	SetHydraAdminSetting()

	SetIndexBaseSetting()

	SetUniQuerySetting()

	SetMetricTaskSetting()

	SetEventTaskSetting()

	SetEventModelSetting()

	s, _ := sonic.MarshalString(appSetting)
	logger.Debug(s)
}

func SetDBSetting() {
	setting, ok := appSetting.DepServices[rdsServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", rdsServiceName)
	}

	appSetting.DBSetting = libdb.DBSetting{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
		DBName:   DATA_BASE_NAME,
	}
}

func SetMQSetting() {
	setting, ok := appSetting.DepServices[mqServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", mqServiceName)
	}
	authSetting, ok := setting["auth"].(map[string]any)
	if !ok {
		logger.Fatalf("service %s auth not found in depServices", mqServiceName)
	}

	appSetting.MQSetting = libmq.MQSetting{
		MQType: setting["mqtype"].(string),
		MQHost: setting["mqhost"].(string),
		MQPort: setting["mqport"].(int),
		Tenant: setting["tenant"].(string),
		Auth: libmq.MQAuthSetting{
			Username:  authSetting["username"].(string),
			Password:  authSetting["password"].(string),
			Mechanism: authSetting["mechanism"].(string),
		},
	}
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

func SetIndexBaseSetting() {
	setting, ok := appSetting.DepServices[indexBaseServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", indexBaseServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.IndexBaseUrl = fmt.Sprintf("%s://%s:%d/api/mdl-index-base/in/v1/index_bases", protocol, host, port)
}

func SetUniQuerySetting() {
	setting, ok := appSetting.DepServices[uniQueryServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", uniQueryServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.UniQueryUrl = fmt.Sprintf("%s://%s:%d/api/mdl-uniquery/in/v1", protocol, host, port)
}

func SetMetricTaskSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.MetricTaskUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1/metric-tasks", protocol, host, port)
}

func SetEventTaskSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.EventTaskUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1/event-task", protocol, host, port)
}

func SetEventModelSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.EventModelUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1/event-models", protocol, host, port)
}
