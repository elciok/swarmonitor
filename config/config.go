package config

import "os"

type SMTPConfig struct {
	Address    string
	Port       string
	User       string
	Password   string
	Domain     string
	AuthMethod string
	From       string
	To         string
}

type Config struct {
	ContainerDir string
	SMTP         *SMTPConfig
}

const SWMON_CONTAINER_DIR = "SWMON_CONTAINER_DIR"
const SWMON_SMTP_ADDRESS = "SWMON_SMTP_ADDRESS"
const SWMON_SMTP_PORT = "SWMON_SMTP_PORT"
const SWMON_SMTP_USERNAME = "SWMON_SMTP_USERNAME"
const SWMON_SMTP_PASSWORD = "SWMON_SMTP_PASSWORD"
const SWMON_SMTP_DOMAIN = "SWMON_SMTP_DOMAIN"
const SWMON_SMTP_AUTH = "SWMON_SMTP_AUTH"
const SWMON_SMTP_TO = "SWMON_SMTP_TO"
const SWMON_SMTP_FROM = "SWMON_SMTP_FROM"

func ReadConfig() *Config {
	config := &Config{}

	config.ContainerDir = os.Getenv(SWMON_CONTAINER_DIR)
	if config.ContainerDir == "" {
		config.ContainerDir = "/etc/swarmonitor/containers"
	}

	config.SMTP = &SMTPConfig{}

	config.SMTP.Address = os.Getenv(SWMON_SMTP_ADDRESS)

	config.SMTP.Port = os.Getenv(SWMON_SMTP_PORT)
	if config.SMTP.Port == "" {
		config.SMTP.Port = "587"
	}

	config.SMTP.User = os.Getenv(SWMON_SMTP_USERNAME)
	config.SMTP.Password = os.Getenv(SWMON_SMTP_PASSWORD)
	config.SMTP.Domain = os.Getenv(SWMON_SMTP_DOMAIN)
	config.SMTP.From = os.Getenv(SWMON_SMTP_FROM)
	config.SMTP.To = os.Getenv(SWMON_SMTP_TO)

	config.SMTP.AuthMethod = os.Getenv(SWMON_SMTP_AUTH)
	if config.SMTP.AuthMethod == "" {
		config.SMTP.AuthMethod = "plain"
	}

	return config
}
