package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
)

func LoadConfig[TConfig any](path string, config *TConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	k := koanf.New(".")

	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		return fmt.Errorf("load yaml: %w", err)
	}

	convert := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "__", ".")
		return s
	}

	err = k.Load(env.Provider("", ".", convert), nil)
	if err != nil {
		return fmt.Errorf("load env: %w", err)
	}

	conf := koanf.UnmarshalConf{
		Tag: "yaml",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			),
		},
	}
	err = k.UnmarshalWithConf("", config, conf)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}
