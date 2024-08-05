package app

import (
	"crypto/tls"
	"encoding/json"
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

	started = o.newApiClient().Heartbeat(app.ctx) == nil

	if !started {
		installed, _ = cmd.CheckInstalled(app.ctx)
	} else {
		installed = true
	}
	os := gorun.GOOS
	runtime.EventsEmit(app.ctx, "ollamaHeartbeat", installed, started, !started && installed && (os == "windows" || os == "darwin"))
}

func (o *Ollama) Start() error {
	err := cmd.StartApp(app.ctx, o.newApiClient())
	if err != nil {
		log.Error().Err(err).Msg("Ollama StartApp")
		return err
	}
	o.Heartbeat()
	return nil
}

func (o *Ollama) List() (*olm.ListResponse, error) {
	return o.newApiClient().List(app.ctx)
}

func (o *Ollama) ListRunning() (*olm.ProcessResponse, error) {
	return o.newApiClient().ListRunning(app.ctx)
}

func (o *Ollama) Generate(requestId, requestStr string) error {
	request := &olm.GenerateRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go o.newApiClient().Generate(app.ctx, request, func(response olm.GenerateResponse) error {
		runtime.EventsEmit(app.ctx, requestId, response)
		return nil
	})
	return nil
}

func (o *Ollama) Chat(requestId, requestStr string) error {
	request := &olm.ChatRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go o.newApiClient().Chat(app.ctx, request, func(response olm.ChatResponse) error {
		runtime.EventsEmit(app.ctx, requestId, response)
		return nil
	})
	return nil
}

func (o *Ollama) Delete(requestStr string) error {
	request := &olm.DeleteRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	return o.newApiClient().Delete(app.ctx, request)
}

func (o *Ollama) Show(requestStr string) (*olm.ShowResponse, error) {
	request := &olm.ShowRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return o.newApiClient().Show(app.ctx, request)
}

func (o *Ollama) Embed(requestStr string) (*olm.EmbedResponse, error) {
	request := &olm.EmbedRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return o.newApiClient().Embed(app.ctx, request)
}

func (o *Ollama) Embeddings(requestStr string) (*olm.EmbeddingResponse, error) {
	request := &olm.EmbeddingRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return o.newApiClient().Embeddings(app.ctx, request)
}

func (o *Ollama) Pull(requestId, requestStr string) error {
	request := &olm.PullRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go o.newApiClient().Pull(app.ctx, request, func(response olm.ProgressResponse) error {
		runtime.EventsEmit(app.ctx, requestId, response)
		return nil
	})
	return nil
}

func (o *Ollama) Push(requestId, requestStr string) error {
	request := &olm.PushRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go o.newApiClient().Push(app.ctx, request, func(response olm.ProgressResponse) error {
		runtime.EventsEmit(app.ctx, requestId, response)
		return nil
	})
	return nil
}

func (o *Ollama) Create(requestId, requestStr string) error {
	request := &olm.CreateRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go o.newApiClient().Create(app.ctx, request, func(response olm.ProgressResponse) error {
		runtime.EventsEmit(app.ctx, requestId, response)
		return nil
	})
	return nil
}

func (o *Ollama) Copy(requestStr string) error {
	request := &olm.CopyRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	return o.newApiClient().Copy(app.ctx, request)
}

func (o *Ollama) SearchOnline(requestStr string) (*olm.SearchResponse, error) {
	request := &olm.SearchRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return o.newOllamaClient().Search(app.ctx, request)
}

func (o *Ollama) LibraryOnline(requestStr string) ([]*olm.ModelInfo, error) {
	request := &olm.LibraryRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return o.newOllamaClient().Library(app.ctx, request)
}

func (o *Ollama) ModelInfoOnline(modelTag string) (*olm.ModelInfoResponse, error) {
	return o.newOllamaClient().ModelInfo(app.ctx, modelTag)
}

func (o *Ollama) newApiClient() *api.Client {
	ollamaHost := config.Config.Ollama.Host

	scheme := ollamaHost.Scheme
	if value, err := configStore.get(configOllamaScheme); err != nil && value != "" {
		scheme = value
	}
	host := ollamaHost.Host
	if value, err := configStore.get(configOllamaHost); err != nil && value != "" {
		host = value
	}
	port := ollamaHost.Port
	if value, err := configStore.get(configOllamaPort); err != nil && value != "" {
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
	if value, err := configStore.get(configProxyScheme); err != nil && value != "" {
		scheme = value
	}
	if value, err := configStore.get(configProxyHost); err != nil && value != "" {
		host = value
	}
	if value, err := configStore.get(configProxyPort); err != nil && value != "" {
		port = value
	}
	if value, err := configStore.get(configProxyUsername); err != nil && value != "" {
		username = value
	}
	if value, err := configStore.get(configProxyPassword); err != nil && value != "" {
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
