package app

var config = Config{}

type Config struct {
	configs map[string]string
}

func (c *Config) update(forceUpdate bool) {
	if c.configs != nil && !forceUpdate {
		return
	}

}

func (c *Config) Value(key string) (string, bool) {
	c.update(false)
	v, ok := c.configs[key]
	return v, ok
}
