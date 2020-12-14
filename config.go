package main

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	configPath     string
	TelegramAPI    string   `yaml:"telegram_api"`
	BotToken       string   `yaml:"bot_token"`
	BotAdmins      []string `yaml:"bot_admins"`
	BotAdminGroups []string `yaml:"bot_admin_groups"`
	EtcdCA         string   `yaml:"etcd_ca"`
	EtcdCert       string   `yaml:"etcd_cert"`
	EtcdKey        string   `yaml:"etcd_key"`
	EtcdHostKey    string   `yaml:"etcd_host_key"`
	EtcdEndpoints  []string `yaml:"etcd_endpoints"`
}

func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfig Config
	raw := rawConfig{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if raw.TelegramAPI == "" {
		raw.TelegramAPI = "https://api.telegram.org"
	}
	*cfg = Config(raw)
	return nil
}

func (cfg *Config) MarshalYAML() (interface{}, error) {
	if cfg.TelegramAPI == "" {
		cfg.TelegramAPI = "https://api.telegram.org"
	}
	return cfg, nil
}

func (cfg *Config) SetConfigPath(configPath string) {
	cfg.configPath = configPath
}

func (cfg *Config) Write() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.configPath, out, 0644)
}

func (cfg *Config) WriteTo(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Write()
}

func (cfg *Config) Load() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	buf, err := ioutil.ReadFile(cfg.configPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, cfg)
}

func (cfg *Config) LoadFrom(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Load()
}
