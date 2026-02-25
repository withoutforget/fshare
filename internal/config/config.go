package config

import (
	"bytes"
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"ssl_mode"`
	MaxConns int32  `toml:"max_conns"`
	MinConns int32  `toml:"min_conns"`
}

type HTTPConfig struct {
	Host             string   `toml:"host"`
	Port             int      `toml:"port"`
	AllowOrigins     []string `toml:"allow_origins"`
	AllowMethods     []string `toml:"allow_methods"`
	AllowHeaders     []string `toml:"allow_headers"`
	AllowCredentials bool     `toml:"allow_credentials"`
	MaxAgeSecs       int      `toml:"max_age_secs"`
}

type LoggerConfig struct {
	Level string `toml:"level"`
	JSON  bool   `toml:"json"`
}

type S3Config struct {
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Bucket          string `toml:"bucket"`
	UseSSL          bool   `toml:"use_ssl"`
	Region          string `toml:"region"`
}

type Config struct {
	Postgres PostgresConfig `toml:"postgres"`
	HTTP     HTTPConfig     `toml:"http"`
	Logger   LoggerConfig   `toml:"logger"`
	S3       S3Config       `toml:"s3"`
}

func NewConfig(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		panic("Couldn't open config file")
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic("Couldn't close config file")
		}
	}()

	var buf bytes.Buffer
	io.Copy(&buf, file)

	var config Config
	_, err = toml.Decode(buf.String(), &config)
	if err != nil {
		panic("Couldn't read config")
	}
	return config
}
