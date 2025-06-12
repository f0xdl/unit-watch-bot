package configs

import (
	"errors"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	ChatID        int64  `env:"CHAT_ID"`
	MqttServer    string `env:"MQTT_SERVER"`
	MqttTopic     string `env:"MQTT_TOPIC"`
	BotDb         string `env:"BOT_DB"`
	TemplateDir   string `env:"TEMPLATE_DIR"`
}

func LoadConfig() (Config, error) {
	cfg := &Config{}
	if err := godotenv.Load(); err != nil {
		return Config{}, errors.New("error loading .env file: " + err.Error())
	}
	if err := env.Parse(cfg); err != nil {
		return Config{}, errors.New("error parsing .env file: " + err.Error())
	}
	return *cfg, nil
}
