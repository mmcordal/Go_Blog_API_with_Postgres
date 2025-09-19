package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

var config *Config

type Config struct {
	Database DBConfig
	Server   ServerConfig
	Secret   JWTConfig
}

type DBConfig struct {
	Name     string
	Username string
	Password string
	Host     string
	Port     string
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	JWTSecret string
}

func setDefaults() {

	viper.SetDefault("database.name", "cleanarch_blog")
	viper.SetDefault("database.username", "mcordal")
	viper.SetDefault("database.password", "157595355")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")

	viper.SetDefault("server.port", "3000") // hata: port string olması gerekirken integer değer girmişim

	viper.SetDefault("secret.jwtsecret", "mcordal123")

}

func Setup() (*Config, error) {
	setDefaults()

	// .env dosyasını yükler (hata olsa env'den devam eder)
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error loading .env file: %v, loading environment variables instead.", err)
	}

	// env değişkenlerini okumak için noktaları alt çizgiye çevir
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if config == nil {
		config = &Config{}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err // değişti ---> return config, err
	}

	if config.Server.Port == "" {
		if p := os.Getenv("SERVER_PORT"); p != "" {
			config.Server.Port = p
		} else {
			config.Server.Port = "3000"
		}
	}

	// JWT Secret için yedek bir değer ayarlar, eğer env'de yoksa default kullanır
	if config.Secret.JWTSecret == "" {
		config.Secret.JWTSecret = os.Getenv("JWT_SECRET")
		if config.Secret.JWTSecret == "" {
			config.Secret.JWTSecret = "default-secret-key" // Varsayılan bir değer
		}
	}

	return config, nil
}

func Get() *Config {
	if config == nil {
		panic("Conifg gelemedi")
	}

	return config
}
