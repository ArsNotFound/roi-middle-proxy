package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel      string               `yaml:"log_level" envconfig:"ROI_LOG_LEVEL"`
	Proxy         *ProxyConfig         `yaml:"proxy"`
	UpstreamProxy *UpstreamProxyConfig `yaml:"upstream_proxy"`
	AllowedHosts  []string             `yaml:"allowed_hosts" env:"ROI_ALLOWED_HOSTS" env-required:"true" env-default:""`
}

func (cfg *Config) String() string {
	return fmt.Sprintf("Config:{LogLevel: %s, Proxy: %s, UpstreamProxy: %s, AllowedHosts: %s}", cfg.LogLevel, cfg.Proxy, cfg.UpstreamProxy, cfg.AllowedHosts)
}

type ProxyConfig struct {
	Port int    `yaml:"port" env:"ROI_PROXY_PORT" env-required:"true" env-default:"3128"`
	Host string `yaml:"host" env:"ROI_PROXY_HOST" env-required:"true" env-default:"127.0.0.1"`
}

func (pc *ProxyConfig) String() string {
	return fmt.Sprintf("ProxyConfig{port: %d, host: %s}", pc.Port, pc.Host)
}

type UpstreamProxyConfig struct {
	HttpProxy  string `yaml:"http_proxy" env:"ROI_UPSTREAM_HTTP_PROXY"`
	HttpsProxy string `yaml:"https_proxy" env:"ROI_UPSTREAM_HTTPS_PROXY"`
	NoProxy    string `yaml:"no_proxy" env:"ROI_UPSTREAM_NO_PROXY"`
}

func (upc *UpstreamProxyConfig) String() string {
	return fmt.Sprintf("UpstreamProxyConfig{http_proxy: %s, https_proxy: %s, no_proxy: %s}", upc.HttpProxy, upc.HttpsProxy, upc.NoProxy)
}

func ParseConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("error opening config file '%s': %w", path, err)
	}

	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &cfg, nil
}
