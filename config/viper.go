package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config ...
type Config interface {
	ReadLogLevelConfig() string
	ReadP2PConfig() P2PConfig
	ReadDBConfig() PostgresDbConfig
	GetString(key string) string
	GetStringMap(key string) map[string]string
	GetInt64(key string) int64
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetStringSlice(key string) []string
	Init()
}

type viperConfig struct{}

func (v *viperConfig) Init() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(`.`, `_`)
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigType(`json`)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
	}
}

func (v *viperConfig) GetString(key string) string {
	return viper.GetString(key)
}

func (v *viperConfig) GetInt64(key string) int64 {
	return viper.GetInt64(key)
}

func (v *viperConfig) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (v *viperConfig) GetFloat64(key string) float64 {
	return viper.GetFloat64(key)
}

func (v *viperConfig) GetStringMap(key string) map[string]string {
	return viper.GetStringMapString(key)
}

func (v *viperConfig) GetStringSlice(key string) []string {
	data := viper.GetString(key)

	var slice []string
	err := json.Unmarshal([]byte(data), &slice)
	if err != nil {
		return viper.GetStringSlice(key)
	}

	return slice
}

func (v *viperConfig) GetStruct(key string) interface{} {
	return viper.Get(key)
}

// NewViperConfig creates new viper for reading config.json
func NewViperConfig() Config {
	v := &viperConfig{}
	v.Init()
	return v
}
