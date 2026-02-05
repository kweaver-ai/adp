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
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"

	"data-model/version"
)

// server配置项
type ServerSetting struct {
	RunMode               string        `mapstructure:"runMode"`
	HttpPort              int           `mapstructure:"httpPort"`
	Language              string        `mapstructure:"language"`
	ReadTimeOut           time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout          time.Duration `mapstructure:"writeTimeOut"`
	PersistSteps          []string      `mapstructure:"persistSteps"`
	WatchMetadataInterval time.Duration `mapstructure:"watchMetadataInterval"`
	WatchMetadataEnabled  bool          `mapstructure:"watchMetadataEnabled"`
}

type ThirdParty struct {
	TingYunTokenActiveTime int `mapstructure:"tingYunTokenActiveTime"`
}

type DataModelJob struct {
	ServiceDeploymentEnabled bool `mapstructure:"serviceDeploymentEnabled"`
}

// app配置项
type AppSetting struct {
	ServerSetting        ServerSetting             `mapstructure:"server"`
	LogSetting           logger.LogSetting         `mapstructure:"log"`
	ObservabilitySetting o11y.ObservabilitySetting `mapstructure:"observability"`
	ThirdParty           ThirdParty                `mapstructure:"thirdParty"`
	DataModelJob         DataModelJob              `mapstructure:"dataModelJob"`
	DepServices          map[string]map[string]any `mapstructure:"depServices"`

	DBSetting         libdb.DBSetting
	MQSetting         libmq.MQSetting
	HydraAdminSetting rest.HydraAdminSetting

	// uniquery url
	UniQueryUrl string
	// 索引库 url
	IndexBaseUrl string
	// data-model-job url
	DataModelJobUrl string
	// permission url
	PermissionUrl string
	// vega excel url
	VegaExcelUrl string
	// vega gateway url
	VegaGatewayUrl string
	// vega data source url
	VegaDatSourceUrl string
	// vega metadata url
	VegaMetadataUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "data-model-config"
	configType string = "yaml"

	rdsServiceName                string = "rds"
	mqServiceName                 string = "mq"
	dataModelJobServiceName       string = "data-model-job"
	indexBaseServiceName          string = "index-base"
	uniQueryServiceName           string = "uniquery"
	vegaDataConnectionServiceName string = "data-connection"
	vegaGatewayServiceName        string = "vega-gateway"
	permissionServiceName         string = "authorization-private"
	hydraAdminServiceName         string = "hydra-admin"

	DATA_BASE_NAME string = "adp"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once

	// 系统持久化步长集
	PersistStepsMap map[string]string
	PersistSteps    []string
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
	for _, step := range appSetting.ServerSetting.PersistSteps {
		stepsMap[step] = step
	}
	PersistStepsMap = stepsMap
	PersistSteps = appSetting.ServerSetting.PersistSteps

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()

	SetMQSetting()

	SetHydraAdminSetting()

	SetDataModelJobSetting()

	SetIndexBaseSetting()

	SetUniQuerySetting()

	SetPermissionSetting()

	SetVegaExcelSetting()

	SetVegaGatewaySetting()

	SetVegaDatSourceSetting()

	SetVegaMetadataSetting()

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

func SetDataModelJobSetting() {
	setting, ok := appSetting.DepServices[dataModelJobServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", dataModelJobServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.DataModelJobUrl = fmt.Sprintf("%s://%s:%d/api/mdl-data-model-job/v1/jobs", protocol, host, port)
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

func SetVegaExcelSetting() {
	setting, ok := appSetting.DepServices[vegaGatewayServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaGatewayServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaExcelUrl = fmt.Sprintf("%s://%s:%d/api/vega-data-source/v1/excel/view", protocol, host, port)
}

func SetVegaGatewaySetting() {
	setting, ok := appSetting.DepServices[vegaGatewayServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaGatewayServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaGatewayUrl = fmt.Sprintf("%s://%s:%d/api/internal/virtual_engine_service/v1/view", protocol, host, port)
}

func SetVegaDatSourceSetting() {
	setting, ok := appSetting.DepServices[vegaDataConnectionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaDataConnectionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaDatSourceUrl = fmt.Sprintf("%s://%s:%d/api/internal/data-connection/v1/datasource", protocol, host, port)
}

func SetVegaMetadataSetting() {
	setting, ok := appSetting.DepServices[vegaDataConnectionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", vegaDataConnectionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.VegaMetadataUrl = fmt.Sprintf("%s://%s:%d/api/internal/data-connection/v1/metadata/data-source", protocol, host, port)
}
