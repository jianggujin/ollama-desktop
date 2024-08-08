package app

import (
	"database/sql"
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
