package configs

import "github.com/spf13/viper"

type RDBConfig struct {
	Address  string
	Port     uint
	DBName   string
	Login    string
	Password string
}

type RedisConfig struct {
	Address  string
	Port     uint
	Password string
}

func NewRDBConfig (v *viper.Viper) RDBConfig {
	return RDBConfig{
		Address: v.GetString("postgres.address"),
		Port: v.GetUint("postgres.port"),
		DBName: v.GetString("postgres.database"),
		Login: v.GetString("postgres.user"),
		Password: v.GetString("postgres.password"),
	}
}
func NewRedisConfig (v *viper.Viper) RedisConfig {
	return RedisConfig {
		Address: v.GetString("redis.address"),
		Port: v.GetUint("redis.port"),
		Password: v.GetString("redis.password"),
	}
}
