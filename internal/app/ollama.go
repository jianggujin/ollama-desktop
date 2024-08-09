package app

import (
	"crypto/tls"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"net"
	"net/http"
	"net/url"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
	olm "ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	"ollama-desktop/internal/ollama/cmd"
	ollama2 "ollama-desktop/internal/ollama/ollama"
	"os"
	gorun "runtime"
	"strings"
	"time"
)

var ollama = Ollama{}

type Ollama struct {
	started bool
	version string
}

func (o *Ollama) Host() string {
	return config.Config.Ollama.Host.String()
}

func (o *Ollama) Envs() []OllamaEnvVar {
	envs := []OllamaEnvVar{
		{"OLLAMA_DEBUG", cleanEnvValue("OLLAMA_DEBUG"), "Show additional debug information (e.g. OLLAMA_DEBUG=1)"},
		{"OLLAMA_FLASH_ATTENTION", cleanEnvValue(""), "Enabled flash attention"},
		{"OLLAMA_HOST", cleanEnvValue(""), "IP Address for the ollama server (default 127.0.0.1:11434)"},
		{"OLLAMA_KEEP_ALIVE", cleanEnvValue(""), "The duration that models stay loaded in memory (default \"5m\")"},
		{"OLLAMA_LLM_LIBRARY", cleanEnvValue(""), "Set LLM library to bypass autodetection"},
		{"OLLAMA_MAX_LOADED_MODELS", cleanEnvValue(""), "Maximum number of loaded models per GPU"},
		{"OLLAMA_MAX_QUEUE", cleanEnvValue(""), "Maximum number of queued requests"},
		{"OLLAMA_MAX_VRAM", cleanEnvValue(""), "Maximum VRAM"},
		{"OLLAMA_MODELS", cleanEnvValue(""), "The path to the models directory"},
		{"OLLAMA_NOHISTORY", cleanEnvValue(""), "Do not preserve readline history"},
		{"OLLAMA_NOPRUNE", cleanEnvValue(""), "Do not prune model blobs on startup"},
		{"OLLAMA_NUM_PARALLEL", cleanEnvValue(""), "Maximum number of parallel requests"},
		{"OLLAMA_ORIGINS", cleanEnvValue(""), "A comma separated list of allowed origins"},
		{"OLLAMA_RUNNERS_DIR", cleanEnvValue(""), "Location for runners"},
		{"OLLAMA_SCHED_SPREAD", cleanEnvValue(""), "Always schedule model across all GPUs"},
		{"OLLAMA_TMPDIR", cleanEnvValue(""), "Location for temporary files"},
	}
	if gorun.GOOS != "darwin" {
		envs = append(envs, OllamaEnvVar{"CUDA_VISIBLE_DEVICES", cleanEnvValue(""), "Set which NVIDIA devices are visible"})
		envs = append(envs, OllamaEnvVar{"HIP_VISIBLE_DEVICES", cleanEnvValue(""), "Set which AMD devices are visible"})
		envs = append(envs, OllamaEnvVar{"ROCR_VISIBLE_DEVICES", cleanEnvValue(""), "Set which AMD devices are visible"})
		envs = append(envs, OllamaEnvVar{"GPU_DEVICE_ORDINAL", cleanEnvValue(""), "Set which AMD devices are visible"})
		envs = append(envs, OllamaEnvVar{"HSA_OVERRIDE_GFX_VERSION", cleanEnvValue(""), "Override the gfx used for all detected AMD GPUs"})
		envs = append(envs, OllamaEnvVar{"OLLAMA_INTEL_GPU", cleanEnvValue(""), "Enable experimental Intel GPU detection"})
	}
	return envs
}

// Clean quotes and spaces from the value
func cleanEnvValue(key string) string {
	return strings.Trim(os.Getenv(key), "\"' ")
}

type OllamaEnvVar struct {
	Name        string
	Value       string
	Description string
}

func (o *Ollama) Version() (string, error) {
	return o.newApiClient().Version(app.ctx)
}

func (o *Ollama) Heartbeat() {
	var installed, started bool
	client := o.newApiClient()
	started = client.Heartbeat(app.ctx) == nil
	if started != o.started {
		o.started = started
		o.version = ""
	}

	if !started {
		installed, _ = cmd.CheckInstalled(app.ctx)
	} else {
		installed = true
	}
	if started && o.version == "" {
		o.version, _ = client.Version(app.ctx)
	}
	os := gorun.GOOS
	runtime.EventsEmit(app.ctx, "ollamaHeartbeat", installed, started, !started && installed && (os == "windows" || os == "darwin"), o.version)
}

func (o *Ollama) Start() error {
	err := cmd.StartApp(app.ctx, o.newApiClient())
	if err != nil {
		log.Error().Err(err).Msg("start ollama app error")
		return err
	}
	o.Heartbeat()
	return nil
}

func (o *Ollama) List() (*olm.ListResponse, error) {
	resp, err := o.newApiClient().List(app.ctx)
	if err != nil {
		log.Error().Err(err).Msg("list ollama model error")
	}
	return resp, err
}

func (o *Ollama) ListRunning() (*olm.ProcessResponse, error) {
	resp, err := o.newApiClient().ListRunning(app.ctx)
	if err != nil {
		log.Error().Err(err).Msg("list ollama running model error")
	}
	return resp, err
}

func (o *Ollama) Delete(request *olm.DeleteRequest) error {
	err := o.newApiClient().Delete(app.ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("delete ollama model error")
	}
	return err
}

func (o *Ollama) Show(request olm.ShowRequest) (*olm.ShowResponse, error) {
	log.Error().Any("request", request).Msg("SHow")
	resp, err := o.newApiClient().Show(app.ctx, &request)
	if err != nil {
		log.Error().Err(err).Msg("show ollama model error")
	}
	return resp, err
}

func (o *Ollama) Pull(requestId string, request *olm.PullRequest) error {
	go func() {
		err := o.newApiClient().Pull(app.ctx, request, func(response olm.ProgressResponse) error {
			runtime.EventsEmit(app.ctx, requestId, response)
			return nil
		})
		if err != nil {
			log.Error().Err(err).Msg("pull ollama model error")
		}
	}()
	return nil
}

func (o *Ollama) SearchOnline(request *olm.SearchRequest) (*olm.SearchResponse, error) {
	resp, err := o.newOllamaClient().Search(app.ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("search ollama model error")
	}
	return resp, err
}

func (o *Ollama) LibraryOnline(request *olm.LibraryRequest) ([]*olm.ModelInfo, error) {
	resp, err := o.newOllamaClient().Library(app.ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("ollama library error")
	}
	return resp, err
}

func (o *Ollama) ModelInfoOnline(modelTag string) (*olm.ModelInfoResponse, error) {
	resp, err := o.newOllamaClient().ModelInfo(app.ctx, modelTag)
	if err != nil {
		log.Error().Err(err).Msg("ollama model info error")
	}
	return resp, err
}

func (o *Ollama) newApiClient() *api.Client {
	ollamaHost := config.Config.Ollama.Host

	scheme := ollamaHost.Scheme
	if value, err := configStore.get(configOllamaScheme); err == nil && value != "" {
		scheme = value
	}
	host := ollamaHost.Host
	if value, err := configStore.get(configOllamaHost); err == nil && value != "" {
		host = value
	}
	port := ollamaHost.Port
	if value, err := configStore.get(configOllamaPort); err == nil && value != "" {
		port = value
	}

	return &api.Client{
		Base: &url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(host, port),
		},
		Http: http.DefaultClient,
	}
}

func (o *Ollama) newOllamaClient() *ollama2.Client {
	base, _ := url.Parse("https://ollama.com")
	var proxy func(*http.Request) (*url.URL, error)

	var scheme, host, port, username, password string
	if config.Config.Proxy != nil {
		proxy := config.Config.Proxy
		scheme = proxy.Scheme
		host = proxy.Host
		port = proxy.Port
		username = proxy.Username
		password = proxy.Password
	}
	if value, err := configStore.get(configProxyScheme); err == nil && value != "" {
		scheme = value
	}
	if value, err := configStore.get(configProxyHost); err == nil && value != "" {
		host = value
	}
	if value, err := configStore.get(configProxyPort); err == nil && value != "" {
		port = value
	}
	if value, err := configStore.get(configProxyUsername); err == nil && value != "" {
		username = value
	}
	if value, err := configStore.get(configProxyPassword); err == nil && value != "" {
		password = value
	}
	if scheme != "" && host != "" && port != "" {
		proxy = http.ProxyURL(o.proxyUrl(scheme, host, port, username, password))
	}

	return &ollama2.Client{
		Base: base,
		Http: &http.Client{
			Timeout: 30 * time.Second, // 设置超时时间为 30 秒
			Transport: &http.Transport{
				Proxy: proxy,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // 不验证证书
				},
			},
		},
	}
}

func (o *Ollama) proxyUrl(scheme, host, port, username, password string) *url.URL {
	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, port),
	}
	if username != "" {
		if password != "" {
			u.User = url.UserPassword(username, password)
		} else {
			u.User = url.User(username)
		}
	}
	return u
}
