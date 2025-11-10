package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"megin/library/files"
	rpcx "megin/library/rpc/config"
	"path"
)

const (
	ConfigDir        = "./resources/"
	ConfigFileFormat = "config-%s.yaml"
)

const (
	EnvProd = "prod"
	EnvTest = "test"
	EnvDev  = "dev"
)

type Database struct {
	Dsn string `yaml:"dsn"`
}

type Redis struct {
	Addr     string `yaml:"addr"` //127.0.0.1:6379
	Password string `yaml:"password"`
}

type Yapi struct {
	Enable bool   `yaml:"enable"`
	Url    string `yaml:"url"`
	Token  string `yaml:"token"`
	CateId int    `yaml:"cate_id"`
}
type Listening struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Elasticsearch struct {
	Addrs    []string `yaml:"addrs"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

// 服务端配置
type ServiceConfig struct {
	ServiceName   string        `yaml:"service_name"`
	Port          string        `yaml:"port"`
	Debug         bool          `yaml:"debug"`
	Env           string        `yaml:"env"`
	Version       string        `yaml:"version"`
	KafkaBroker   string        `yaml:"kafka_broker"`
	Database      Database      `yaml:"database"`
	JwtSecret     string        `yaml:"jwt_secret"`
	Elasticsearch Elasticsearch `yaml:"elasticsearch"`
	Redis         Redis         `yaml:"redis"`
	Yapi          Yapi          `yaml:"yapi"`
	Rpcx          rpcx.Config   `yaml:"rpcx"`
	Listening     Listening     `yaml:"listening"`

	ProjectPath string
}

func (config *ServiceConfig) IsProdEnv() bool {
	if config.Env != EnvProd {
		return false
	}
	return true
}

func (config *ServiceConfig) IsDevEnv() bool {
	if config.Env != EnvDev {
		return false
	}
	return true
}

func (config *ServiceConfig) IsDebug() bool {
	if config.IsProdEnv() {
		return false
	}
	if config.Debug == true {
		return true
	}
	return false
}

var config = new(ServiceConfig)

func GetConfig() *ServiceConfig {
	return config
}

func GetConfigPath(env string) string {
	return path.Join(ConfigDir, fmt.Sprintf(ConfigFileFormat, env))
}

func InitConfig(path string) *ServiceConfig {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("InitConfig: ", path, ".  config err: ", err.Error())
		return nil
	}

	if err = yaml.Unmarshal(content, config); err != nil {
		log.Fatal("InitConfig: ", path, ".  config err: ", err.Error())
		return nil
	}
	config.ProjectPath = files.ProjectDir()
	return config
}
