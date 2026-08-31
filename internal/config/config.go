package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const envPrefix = "NOTRUST"

// Config holds all daemon settings. Add new fields here as the daemon
// grows (docker connection, idle thresholds, proxy, notify, logging).
type Config struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
	PauseAfter   time.Duration `mapstructure:"pause_after"`
	StopAfter    time.Duration `mapstructure:"stop_after"`
}

// Load reads configuration in order of increasing precedence: built in
// defaults, then a config file, then environment variables prefixed
// with NOTRUST_. If path is empty, ./config.yaml is tried first, then
// ~/.config/notrust/config.yaml, then /etc/notrust/config.yaml.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/notrust")
		v.AddConfigPath("/etc/notrust")
	}

	// defaults registered before AutomaticEnv so env vars reliably
	// override every known key during Unmarshal, see note above
	v.SetDefault("poll_interval", 3*time.Second)
	v.SetDefault("pause_after", 10*time.Minute)
	v.SetDefault("stop_after", 30*time.Minute)

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// no file found is fine, defaults and env vars still apply
	}

	var cfg Config
	hook := viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())
	if err := v.Unmarshal(&cfg, hook); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}
