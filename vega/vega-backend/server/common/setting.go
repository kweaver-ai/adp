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

	"vega-backend/version"
)

// ServerSetting server配置项
type ServerSetting struct {
	RunMode      string        `mapstructure:"runMode"`
	HttpPort     int           `mapstructure:"httpPort"`
	Language     string        `mapstructure:"language"`
	ReadTimeOut  time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout time.Duration `mapstructure:"writeTimeOut"`
}

// CryptoSetting RSA 密钥配置项
type CryptoSetting struct {
	Enabled        bool   `mapstructure:"enabled"`
	PrivateKey     string `mapstructure:"-"`              // RSA 私钥 (PEM 格式) - 从文件读取
	PublicKey      string `mapstructure:"-"`              // RSA 公钥 (PEM 格式) - 从文件读取
	PrivateKeyPath string `mapstructure:"privateKeyPath"` // RSA 私钥文件路径
	PublicKeyPath  string `mapstructure:"publicKeyPath"`  // RSA 公钥文件路径
}

// RedisSetting Redis 配置项
type RedisSetting struct {
	Host     string
	Port     int
	Username string
	Password string
}

// AppSetting app配置项
type AppSetting struct {
	ServerSetting        ServerSetting             `mapstructure:"server"`
	LogSetting           logger.LogSetting         `mapstructure:"log"`
	ObservabilitySetting o11y.ObservabilitySetting `mapstructure:"observability"`
	CryptoSetting        CryptoSetting             `mapstructure:"crypto"`
	DepServices          map[string]map[string]any `mapstructure:"depServices"`

	DBSetting         libdb.DBSetting
	MQSetting         libmq.MQSetting
	OpenSearchSetting rest.OpenSearchClientConfig
	RedisSetting      RedisSetting
}

const (
	configPath string = "./config/"
	configName string = "vega-backend-config"
	configType string = "yaml"

	rdsServiceName        string = "rds"
	mqServiceName         string = "mq"
	opensearchServiceName string = "opensearch"
	redisServiceName      string = "redis"

	DATA_BASE_NAME string = "adp"
)

var (
	appSetting  *AppSetting
	vp          *viper.Viper
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

	if err := loadCryptoKeys(); err != nil {
		logger.Fatalf("Failed to load crypto keys: %s\n", err)
	}

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()

	SetMQSetting()

	SetOpenSearchSetting()

	SetRedisSetting()

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
		Username: setting["username"].(string),
		Password: setting["password"].(string),
	}
}

// SetRedisSetting 设置 Redis 配置
func SetRedisSetting() {
	setting, ok := appSetting.DepServices[redisServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", redisServiceName)
	}

	appSetting.RedisSetting = RedisSetting{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Username: setting["username"].(string),
		Password: setting["password"].(string),
	}
}

// GetDBSetting 获取DB配置
func GetDBSetting() libdb.DBSetting {
	return appSetting.DBSetting
}

// GetOpenSearchSetting 获取OpenSearch配置
func GetOpenSearchSetting() rest.OpenSearchClientConfig {
	return appSetting.OpenSearchSetting
}

// GetServerSetting 获取Server配置
func GetServerSetting() ServerSetting {
	return appSetting.ServerSetting
}

// GetHttpPort 获取HTTP端口
func GetHttpPort() string {
	return fmt.Sprintf(":%d", appSetting.ServerSetting.HttpPort)
}

// loadCryptoKeys 从文件加载 RSA 密钥
func loadCryptoKeys() error {
	if !appSetting.CryptoSetting.Enabled {
		return nil
	}

	if appSetting.CryptoSetting.PrivateKeyPath == "" {
		return fmt.Errorf("privateKeyPath is required when crypto is enabled")
	}
	if appSetting.CryptoSetting.PublicKeyPath == "" {
		return fmt.Errorf("publicKeyPath is required when crypto is enabled")
	}

	privateKeyContent, err := os.ReadFile(appSetting.CryptoSetting.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}
	appSetting.CryptoSetting.PrivateKey = string(privateKeyContent)

	publicKeyContent, err := os.ReadFile(appSetting.CryptoSetting.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}
	appSetting.CryptoSetting.PublicKey = string(publicKeyContent)

	return nil
}
