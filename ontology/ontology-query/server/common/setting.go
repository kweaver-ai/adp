package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"

	"ontology-query/version"
)

// server配置项
type ServerSetting struct {
	RunMode                  string        `mapstructure:"runMode"`
	HttpPort                 int           `mapstructure:"httpPort"`
	Language                 string        `mapstructure:"language"`
	ReadTimeOut              time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout             time.Duration `mapstructure:"writeTimeOut"`
	ViewDataTimeout          string        `mapstructure:"viewDataTimeout"`
	DefaultSmallModelEnabled bool          `mapstructure:"defaultSmallModelEnabled"`
}

// app配置项
type AppSetting struct {
	ServerSetting        ServerSetting             `mapstructure:"server"`
	LogSetting           logger.LogSetting         `mapstructure:"log"`
	ObservabilitySetting o11y.ObservabilitySetting `mapstructure:"observability"`
	DepServices          map[string]map[string]any `mapstructure:"depServices"`

	OpenSearchSetting rest.OpenSearchClientConfig
	HydraAdminSetting rest.HydraAdminSetting

	OntologyManagerUrl string
	UniQueryUrl        string
	// 算子执行 url
	AgentOperatorUrl string
	// 工具箱执行 url
	ToolBoxUrl string
	// MCP 执行 url
	MCPUrl string
	// model factory url
	ModelFactoryManagerUrl string
	// model factory api url
	ModelFactoryAPIUrl string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "ontology-query-config"
	configType string = "yaml"

	opensearchServiceName          string = "opensearch"
	hydraAdminServiceName          string = "hydra-admin"
	modelFactoryManagerServiceName string = "mf-model-manager"
	modelFactoryAPIServiceName     string = "mf-model-api"
	ontologyManagerServiceName     string = "ontology-manager"
	uniQueryServiceName            string = "uniquery"
	agentOperatorServiceName       string = "agent-operator-integration"
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

	SetOpenSearchSetting()
	SetHydraAdminSetting()

	SetOntologyManagerSetting()
	SetModelFactoryManagerSetting()

	SetModelFactoryAPISetting()

	SetUniQuerySetting()

	SetAgentOperatorSetting()

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
func SetModelFactoryManagerSetting() {
	setting, ok := appSetting.DepServices[modelFactoryManagerServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", modelFactoryManagerServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.ModelFactoryManagerUrl = fmt.Sprintf("%s://%s:%d/api/private/mf-model-manager/v1", protocol, host, port)
}

func SetModelFactoryAPISetting() {
	setting, ok := appSetting.DepServices[modelFactoryAPIServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", modelFactoryAPIServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.ModelFactoryAPIUrl = fmt.Sprintf("%s://%s:%d/api/private/mf-model-api/v1", protocol, host, port)
}

func SetOntologyManagerSetting() {
	setting, ok := appSetting.DepServices[ontologyManagerServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", ontologyManagerServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.OntologyManagerUrl = fmt.Sprintf("%s://%s:%d/api/ontology-manager/in/v1/knowledge-networks", protocol, host, port)
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

func SetAgentOperatorSetting() {
	setting, ok := appSetting.DepServices[agentOperatorServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", agentOperatorServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.AgentOperatorUrl = fmt.Sprintf("%s://%s:%d/api/agent-operator-integration/internal-v1/operator", protocol, host, port)
	// ToolBox URL: /api/agent-operator-integration/internal-v1/tool-box/{box_id}/proxy/{tool_id}
	appSetting.ToolBoxUrl = fmt.Sprintf("%s://%s:%d/api/agent-operator-integration/internal-v1/tool-box", protocol, host, port)
	// MCP URL: /api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_id}/tool/call
	appSetting.MCPUrl = fmt.Sprintf("%s://%s:%d/api/agent-operator-integration/internal-v1/mcp", protocol, host, port)
}
