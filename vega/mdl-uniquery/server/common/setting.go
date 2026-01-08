package common

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"

	"uniquery/version"
)

// server配置项
type ServerSetting struct {
	RunMode                  string        `mapstructure:"runMode"`
	HttpPort                 int           `mapstructure:"httpPort"`
	Language                 string        `mapstructure:"language"`
	ReadTimeOut              time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout             time.Duration `mapstructure:"writeTimeOut"`
	FixedQuerySteps          []string      `mapstructure:"fixedQuerySteps"`
	FullCacheRefreshInterval time.Duration `mapstructure:"fullCacheRefreshInterval"`
	IgnoringHcts             bool          `mapstructure:"ignoringHcts"`
	EventSubscribeEnabled    bool          `mapstructure:"eventSubscribeEnabled"`
}

// 协程池配置项
type PoolSetting struct {
	ExecutePoolSize     int `mapstructure:"executePoolSize"`
	MegerPoolSize       int `mapstructure:"megerPoolSize"`
	BatchSubmitPoolSize int `mapstructure:"batchSubmitPoolSize"`
	ViewPoolSize        int `mapstructure:"viewPoolSize"`
	ObjectivePoolSize   int `mapstructure:"objectivePoolSize"`
}

// promql 配置项.
// LookbackDelta: 瞬时查询时, 查询的数据返回往前回退的时间差
// MaxSearchSeriesSize: 配置每个查询能查询的最大序列数
type PromqlSetting struct {
	LookbackDelta       time.Duration `mapstructure:"lookbackDelta"`
	MaxSearchSeriesSize int           `mapstructure:"maxSearchSeriesSize"`
}

type ThirdParty struct {
	TingYunMaxTimePeriod int64 `mapstructure:"tingYunMaxTimePeriod"`
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

// app配置项
type AppSetting struct {
	ServerSetting        ServerSetting             `mapstructure:"server"`
	LogSetting           logger.LogSetting         `mapstructure:"log"`
	PoolSetting          PoolSetting               `mapstructure:"pool"`
	PromqlSetting        PromqlSetting             `mapstructure:"promql"`
	ThirdParty           ThirdParty                `mapstructure:"thirdParty"`
	KafkaSetting         KafkaSetting              `mapstructure:"kafka"`
	ObservabilitySetting o11y.ObservabilitySetting `mapstructure:"observability"`
	DepServices          map[string]map[string]any `mapstructure:"depServices"`

	DBSetting         libdb.DBSetting
	MQSetting         libmq.MQSetting
	OpenSearchSetting rest.OpenSearchClientConfig
	HydraAdminSetting rest.HydraAdminSetting

	DataManagerUrl        string
	SearchUrl             string
	DataModelInUrl        string
	DataViewUrl           string
	IndexBaseUrl          string
	DataDictUrl           string
	DataModelUrl          string
	EventModelUrl         string
	DataConnDataSourceUrl string
	DataConnGatewayUrl    string
	VegaGatewayProUrl     string
	VegaGatewaysUrl       string
	VegaViewUrl           string
	// permission url
	PermissionUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "uniquery-config"
	configType string = "yaml"

	rdsServiceName        string = "rds"
	mqServiceName         string = "mq"
	opensearchServiceName string = "opensearch"

	dataManagerServiceName        string = "data-manager"
	dataModelServiceName          string = "data-model"
	hydraAdminServiceName         string = "hydra-admin"
	indexBaseServiceName          string = "index-base"
	permissionServiceName         string = "authorization-private"
	searchServiceName             string = "search"
	vegaDataConnectionServiceName string = "data-connection"
	vegaGatewayProServiceName     string = "vega-gateway-pro"
	vegaGatewayServiceName        string = "vega-gateway"
	vegaViewServiceName           string = "vega-logic-view"

	DATA_BASE_NAME string = "adp"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once

	// 系统持久化步长集
	FixedStepsMap map[string]string
	FixedSteps    []string

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

	// 系统步长集转成map变成全局变量
	stepsMap := make(map[string]string)
	for _, step := range appSetting.ServerSetting.FixedQuerySteps {
		stepsMap[step] = step
	}
	FixedStepsMap = stepsMap
	FixedSteps = appSetting.ServerSetting.FixedQuerySteps

	// 加载时区
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		loc = time.Local
		logger.Warnf("WARNING: Failed to load timezone from env, using Local[%v] as default. Error: %v\n", time.Local, err)
	}
	APP_LOCATION = loc

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()

	SetMQSetting()

	SetOpenSearchSetting()

	SetDataConnGatewaySetting()

	SetVegaGatewayProSetting()

	SetDataDictSetting()

	SetDataModelSetting()

	SetDataModelInSetting()

	SetDataViewSetting()

	SetEventModelSetting()

	SetHydraAdminSetting()

	SetIndexBaseSetting()

	SetPermissionSetting()

	SetVegaDatSourceSetting()

	SetVegaGatewaySetting()

	SetVegaViewSetting()

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

func SetOpenSearchSetting() {
	setting, ok := appSetting.DepServices[opensearchServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", opensearchServiceName)
	}

	appSetting.OpenSearchSetting = rest.OpenSearchClientConfig{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Protocol: setting["protocol"].(string),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
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

func SetDataManagerSetting() {
	setting, ok := appSetting.DepServices[dataManagerServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataManagerServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataManagerUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetSearchSetting() {
	setting, ok := appSetting.DepServices[searchServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", searchServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.SearchUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetDataModelInSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	// "http://%s/api/mdl-data-model/v1/metric-models/451134087558610171"
	appSetting.DataModelInUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1", protocol, host, port)
}

func SetDataViewSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataViewUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1/data-views", protocol, host, port)
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

func SetDataDictSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	// "http://%s/api/mdl-data-model/v1/metric-models/451134087558610171"
	appSetting.DataDictUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model/in/v1/data-dicts", protocol, host, port)
}

func SetDataModelSetting() {
	setting, ok := appSetting.DepServices[dataModelServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataModelUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model", protocol, host, port)
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

func SetVegaViewSetting() {
	setting, ok := appSetting.DepServices[vegaViewServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaViewServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaViewUrl = fmt.Sprintf("%s://%s:%d/api/internal/vega-logic-view/v1/form-view", protocol, host, port)
}

func SetVegaDatSourceSetting() {
	setting, ok := appSetting.DepServices[vegaDataConnectionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaDataConnectionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataConnDataSourceUrl = fmt.Sprintf("%s://%s:%d/api/internal/data-connection/v1/datasource", protocol, host, port)
}

func SetDataConnGatewaySetting() {
	setting, ok := appSetting.DepServices[vegaDataConnectionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaDataConnectionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataConnGatewayUrl = fmt.Sprintf("%s://%s:%d/api/internal/data-connection/v1/gateway", protocol, host, port)
}

func SetVegaGatewayProSetting() {
	setting, ok := appSetting.DepServices[vegaGatewayProServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaGatewayProServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaGatewayProUrl = fmt.Sprintf("%s://%s:%d/api/internal/vega-gateway/v2", protocol, host, port)
}

func SetVegaGatewaySetting() {
	setting, ok := appSetting.DepServices[vegaGatewayServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaGatewayServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaGatewaysUrl = fmt.Sprintf("%s://%s:%d/api/internal/virtual_engine_service/v1", protocol, host, port)
}

func SetPermissionSetting() {
	setting, ok := appSetting.DepServices[permissionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", permissionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.PermissionUrl = fmt.Sprintf("%s://%s:%d/api/authorization/v1", protocol, host, port)
}
