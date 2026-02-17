package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"github.com/user/subscriptions-monitor/internal/provider"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Subscriptions []provider.SubscriptionEntry `yaml:"subscriptions" mapstructure:"subscriptions"`
	Settings      Settings                     `yaml:"settings" mapstructure:"settings"`
}

type Settings struct {
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
	APIPort int           `yaml:"api_port" mapstructure:"api_port"`
}

func Load(configFile string) (*Config, error) {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "sub-mon"))
		}
		v.AddConfigPath(".")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := DefaultConfig()

	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
	)

	if err := v.Unmarshal(cfg, viper.DecodeHook(decodeHook)); err != nil {
		return nil, err
	}

	for i := range cfg.Subscriptions {
		cfg.Subscriptions[i].Auth.Key = ExpandEnvVars(cfg.Subscriptions[i].Auth.Key)
	}

	return cfg, nil
}

func Save(cfg *Config, configFile string) error {
	path := configFile
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".config", "sub-mon", "config.yaml")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func DefaultConfig() *Config {
	return &Config{
		Settings: Settings{
			Timeout: 10 * time.Second,
			APIPort: 3456,
		},
	}
}

func ExpandEnvVars(s string) string {
	return os.ExpandEnv(s)
}
