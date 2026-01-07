// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package config

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/telemetry"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
	"github.com/creasty/defaults"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config configuration
type Config struct {
	Project             Project               `yaml:"project"`
	OAuth               OAuthConfig           `yaml:"oauth"`
	UserMgnt            PrivateBaseConfig     `yaml:"user_management"`
	OntologyManager     PrivateBaseConfig     `yaml:"ontology_manager"`
	OntologyQuery       PrivateBaseConfig     `yaml:"ontology_query"`
	AgentApp            PrivateBaseConfig     `yaml:"agent_app"`
	OperatorIntegration PrivateBaseConfig     `yaml:"operator_integration"` // Operator integration service configuration
	RedisConfig         RedisConfig           `yaml:"redis"`
	Logger              interfaces.Logger     `yaml:"-"`
	DeployAgent         DeployAgentConfig     `yaml:"deploy_agent"`          // Dependent agent configuration
	ConceptSearchConfig KnConceptSearchConfig `yaml:"concept_search_config"` // Knowledge network concept search configuration
	DataRetrieval       PrivateBaseConfig     `yaml:"data_retrieval"`        // Data retrieval configuration
	Observability       ObservabilityConfig   `yaml:"-"`
}

// ObservabilityConfig trace configuration
type ObservabilityConfig struct {
	TraceType                 telemetry.ExporterType `mapstructure:"traceType"`
	o11y.ObservabilitySetting `mapstructure:",squash"`
}

// Project configuration
type Project struct {
	Host        string              `yaml:"host"`
	Port        int                 `yaml:"port"`
	Language    string              `yaml:"language"`
	LoggerLevel int                 `yaml:"logger_level"`
	Name        string              `yaml:"name" default:"agent-retrieval"`
	MachineID   string              `yaml:"machine_id"`
	PodID       string              `yaml:"pod_id" default:"DEFAULT_POD_ID"`
	Debug       bool                `yaml:"debug"`
	CommitInfo  utils.GitCommitInfo `yaml:"-"`
}

// GetLogger gets logger
func (conf *Config) GetLogger() interfaces.Logger {
	if conf.Logger == nil {
		return logger.DefaultLogger()
	}
	return conf.Logger
}

// OAuthConfig OAuth connection info
type OAuthConfig struct {
	PublicBaseConfig `yaml:",inline"`
	AdminHost        string `yaml:"admin_host"`
	AdminPort        int    `yaml:"admin_port"`
	AdminProtocol    string `yaml:"admin_protocol"`
	AdminPrefix      string `yaml:"admin_prefix"`
}

// PublicBaseConfig public base configuration
type PublicBaseConfig struct {
	PublicHost     string `yaml:"public_host"`
	PublicPort     int    `yaml:"public_port"`
	PublicProtocol string `yaml:"public_protocol"`
}

// PrivateBaseConfig private base configuration
type PrivateBaseConfig struct {
	PrivateHost     string `yaml:"private_host"`
	PrivatePort     int    `yaml:"private_port"`
	PrivateProtocol string `yaml:"private_protocol"`
}

// OpenSearchConfig OpenSearch configuration
type OpenSearchConfig struct {
	Protocol string `yaml:"protocol"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"user"`
	Password string `yaml:"password"`
}

// KnConceptSearchConfig knowledge network concept search configuration
type KnConceptSearchConfig struct {
	ConceptRecallSize int `yaml:"concept_recall_size"` // Concept rough recall size
	KnnKValue         int `yaml:"knn_k"`               // knn k value
}

// SetMachineID sets machine ID
func (conf *Project) SetMachineID() {
	// Generate MachineID
	if conf.MachineID == "" {
		mid := os.Getenv(conf.PodID)
		if mid == "" {
			mid, _ = os.Hostname()
			// Empty is allowed
			mid = utils.MD5(mid)
			mid = mid[:8]
		}
		conf.MachineID = mid
	}
}

// GetMachineID gets machine ID
func (conf *Project) GetMachineID() string {
	return conf.MachineID
}

// DeployAgentConfig dependent agent configuration
type DeployAgentConfig struct {
	ConceptIntentionAnalysisAgentKey   string `yaml:"concept_intention_analysis_agent_key"`   // Concept intention analysis agent Key
	ConceptRetrievalStrategistAgentKey string `yaml:"concept_retrieval_strategist_agent_key"` // Concept retrieval strategist agent Key
	MetricDynamicParamsGeneratorKey    string `yaml:"metric_dynamic_params_generator_key"`    // Metric dynamic params generator Key
	OperatorDynamicParamsGeneratorKey  string `yaml:"operator_dynamic_params_generator_key"`  // Operator dynamic params generator Key
}

var (
	once         sync.Once
	configLoader *Config
)

// NewConfigLoader gets configuration
func NewConfigLoader() *Config {
	once.Do(func() {
		configFilePath := "/sysvol/config/agent-retrieval.yaml"
		secretFilePath := "/sysvol/secret/agent-retrieval-secret.yaml"

		// Set default configuration
		configLoader = &Config{}
		err := configLoader.localConfig(configFilePath)
		if err != nil {
			log.Panicln("Error: load local config failed: ", err)
			return
		}
		err = configLoader.localConfig(secretFilePath)
		if err != nil {
			log.Panicln("Error: load local secret failed: ", err)
			return
		}
		// Initialize observability related configuration
		configLoader.initO11yAndLog()
		// Set machine ID
		configLoader.Project.SetMachineID()
		overrideWithEnv(configLoader)
	})
	return configLoader
}

func (conf *Config) localConfig(path string) (err error) {
	err = defaults.Set(conf)
	if err != nil {
		return
	}

	file, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return
	}
	err = yaml.Unmarshal(file, conf)
	return
}

// overrideWithEnv automatically traverses struct, using reflection to override with env variables based on tags
func overrideWithEnv(cfg interface{}) {
	v := reflect.ValueOf(cfg).Elem() // Get pointer to struct
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct {
			// Recursively handle nested struct
			overrideWithEnv(field.Addr().Interface())
			continue
		}

		// Get env tag of field
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			continue // Skip if env tag is not defined
		}

		// Check if env variable exists
		envValue, exists := os.LookupEnv(envTag)
		if !exists {
			continue // Skip if env key does not exist
		}

		// If key exists but value is empty, set field to zero value of type
		if envValue == "" {
			field.Set(reflect.Zero(field.Type()))
			continue
		}

		// Use reflection to set field value directly, type match required
		switch field.Kind() {
		case reflect.String:
			field.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(envValue, 10, 64)
			if err == nil {
				field.SetInt(intValue)
			}
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(envValue)
			if err == nil {
				field.SetBool(boolValue)
			}
		default:
			panic("Unsupported field type for env override")
		}
	}
}

// Load & Initialize observability related configuration
func (conf *Config) initO11yAndLog() {
	// Initialize logger
	level := logger.Level(configLoader.Project.LoggerLevel)
	if configLoader.Project.Debug {
		level = logger.LevelDebug
	}

	// Load configuration file
	viper.SetConfigName("observability")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/sysvol/config/")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&conf.Observability); err != nil {
		panic(err)
	}
	serverInfo := o11y.ServerInfo{
		ServerName:    conf.Project.Name,
		ServerVersion: "1.0.0",
		Language:      "Go",
		GoVersion:     runtime.Version(),
		GoArch:        runtime.GOARCH,
	}

	// Initialize observability
	if conf.Observability.TraceEnabled && conf.Observability.TraceType == telemetry.ExporterTypeJaeger {
		_, err := telemetry.InitJaegerExporter(conf.Project.Name, conf.Observability.TraceProvider, conf.Observability.GrpcTraceFeedIngesterUrl)
		if err != nil {
			panic(err)
		}
		conf.Observability.TraceEnabled = false
	}
	o11y.Init(serverInfo, conf.Observability.ObservabilitySetting)
	// Initialize logger
	if conf.Observability.LogEnabled {
		configLoader.Logger = telemetry.NewSamplerLogger(logger.NewLogger(level, logger.MaxCalldepth))
		return
	}
	configLoader.Logger = logger.NewLogger(level, logger.DefaultCalldepth)
}
