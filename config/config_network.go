package config

// ReadLogLevelConfig reads logger level from config.json
func (v *viperConfig) ReadLogLevelConfig() string {
	return v.GetString("logger-level")
}

type DBConfig struct {
	Address  string
	Password string
	Db       int64
}

func (v *viperConfig) ReadDbConfig() DBConfig {
	return DBConfig{
		Address:  v.GetString("redis.addresss"),
		Password: v.GetString("redis.password"),
		Db:       v.GetInt64("redis.Db"),
	}
}
