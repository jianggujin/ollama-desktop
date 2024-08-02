package app

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
	configs map[string]string
}

func (c *Config) update(forceUpdate bool) {
	if c.configs != nil && !forceUpdate {
		return
	}
	c.configs, _ = dao.configs()
}

func (c *Config) Get(key string) (string, bool) {
	c.update(false)
	v, ok := c.configs[key]
	return v, ok
}

func (c *Config) Set(key, value string) error {
	if err := dao.saveOrUpdateConfig(key, value); err != nil {
		return err
	}
	return nil
}
