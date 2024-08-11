package app

import (
	"database/sql"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
	"time"
)

const (
	configOllamaScheme = "ollama.scheme"
	configOllamaHost   = "ollama.host"
	configOllamaPort   = "ollama.port"

	configProxyScheme   = "proxy.scheme"
	configProxyHost     = "proxy.host"
	configProxyPort     = "proxy.port"
	configProxyUsername = "proxy.username"
	configProxyPassword = "proxy.password"
)

var configStore = Config{}

type Config struct {
	configCaches map[string]string
}

func (c *Config) configs(forceUpdate bool) (map[string]string, error) {
	if c.configCaches != nil && !forceUpdate {
		return c.configCaches, nil
	}
	sqlStr := `select config_key, config_value from t_config`
	rows, err := dao.db().QueryContext(app.ctx, sqlStr)
	if err != nil {
		log.Error().Err(err).Msg("query config error")
		return nil, err
	}
	defer rows.Close()
	configs := make(map[string]string)
	for rows.Next() {
		var configKey, configValue string
		if err := rows.Scan(&configKey, &configValue); err != nil {
			return nil, err
		}
		configs[configKey] = configValue
	}
	c.configCaches = configs
	return c.configCaches, nil
}

func (c *Config) get(key string) (string, error) {
	configs, err := c.configs(false)
	if err != nil {
		return "", err
	}
	return configs[key], nil
}

func (c *Config) getOrDefault(key, defValue string) (string, error) {
	value, err := c.get(key)
	if err != nil || value == "" {
		return defValue, err
	}
	return value, nil
}

func (c *Config) accept(key string, fn func(value string)) error {
	value, err := c.get(key)
	if err != nil || value == "" {
		return err
	}
	fn(value)
	return nil
}

func (c *Config) set(key, value string) error {
	configs, err := c.configs(false)
	if err != nil {
		return err
	}
	return dao.transaction(func(tx *sql.Tx) (err error) {
		if _, ok := configs[key]; ok {
			sqlStr := `update t_config set config_value = ?, updated_at = ? where config_key = ?`
			_, err = tx.ExecContext(app.ctx, sqlStr, value, time.Now(), key)
		} else {
			sqlStr := `insert into t_config(config_key, config_value, created_at, updated_at) values(?, ?, ?, ?)`
			_, err = tx.ExecContext(app.ctx, sqlStr, key, value, time.Now(), time.Now())
		}
		if err != nil {
			log.Error().Err(err).Msg("set config error")
		}
		return
	})
}

type OllamaConfig struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   string `json:"port"`
}

func (c *Config) OllamaConfigs() (*OllamaConfig, error) {
	configs, err := c.configs(false)
	if err != nil {
		return nil, err
	}
	ollamaHost := config.Config.Ollama.Host
	ollamaConfig := &OllamaConfig{
		Scheme: ollamaHost.Scheme,
		Host:   ollamaHost.Host,
		Port:   ollamaHost.Port,
	}

	for name, value := range configs {
		if value == "" {
			continue
		}
		switch name {
		case configOllamaScheme:
			ollamaConfig.Scheme = value
		case configOllamaHost:
			ollamaConfig.Host = value
		case configOllamaPort:
			ollamaConfig.Port = value
		}
	}
	return ollamaConfig, nil
}

func (c *Config) SaveOllamaConfigs(request *OllamaConfig) error {
	if err := c.set(configOllamaScheme, request.Scheme); err != nil {
		return err
	}
	if err := c.set(configOllamaHost, request.Host); err != nil {
		c.configs(true)
		return err
	}
	if err := c.set(configOllamaPort, request.Port); err != nil {
		c.configs(true)
		return err
	}
	c.configs(true)
	return nil
}

type ProxyConfig struct {
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Config) ProxyConfigs() (*ProxyConfig, error) {
	configs, err := c.configs(false)
	if err != nil {
		return nil, err
	}
	proxyConfig := &ProxyConfig{}

	if config.Config.Proxy != nil {
		proxy := config.Config.Proxy
		proxyConfig.Scheme = proxy.Scheme
		proxyConfig.Host = proxy.Host
		proxyConfig.Port = proxy.Port
		proxyConfig.Username = proxy.Username
		proxyConfig.Password = proxy.Password
	}

	for name, value := range configs {
		if value == "" {
			continue
		}
		switch name {
		case configProxyScheme:
			proxyConfig.Scheme = value
		case configProxyHost:
			proxyConfig.Host = value
		case configProxyPort:
			proxyConfig.Port = value
		case configProxyUsername:
			proxyConfig.Username = value
		case configProxyPassword:
			proxyConfig.Password = value
		}
	}
	return proxyConfig, nil
}

func (c *Config) SaveProxyConfigs(request *ProxyConfig) error {
	if err := c.set(configProxyScheme, request.Scheme); err != nil {
		return err
	}
	if err := c.set(configProxyHost, request.Host); err != nil {
		c.configs(true)
		return err
	}
	if err := c.set(configProxyPort, request.Port); err != nil {
		c.configs(true)
		return err
	}
	if err := c.set(configProxyUsername, request.Username); err != nil {
		c.configs(true)
		return err
	}
	if err := c.set(configProxyPassword, request.Password); err != nil {
		c.configs(true)
		return err
	}
	c.configs(true)
	return nil
}
