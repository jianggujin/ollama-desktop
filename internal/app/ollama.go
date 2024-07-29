package app

import (
	"context"
	"encoding/json"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
	olm "ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	"ollama-desktop/internal/ollama/cmd"
	"os"
	gorun "runtime"
	"strings"
)

var ollama = Ollama{}

type Ollama struct {
	ctx context.Context
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
	return api.ClientFromConfig().Version(o.ctx)
}

func (o *Ollama) Heartbeat() {
	var installed, started bool
	client := api.ClientFromConfig()

	started = client.Heartbeat(o.ctx) == nil

	if !started {
		installed, _ = cmd.CheckInstalled(o.ctx)
	} else {
		installed = true
	}
	os := gorun.GOOS
	runtime.EventsEmit(o.ctx, "ollamaHeartbeat", installed, started, !started && installed && (os == "windows" || os == "darwin"))
}

func (o *Ollama) Start() error {
	client := api.ClientFromConfig()

	err := cmd.StartApp(o.ctx, client)
	if err != nil {
		log.Error().Err(err).Msg("Ollama StartApp")
		return err
	}
	o.Heartbeat()
	return nil
}

func (o *Ollama) List() (*olm.ListResponse, error) {
	return api.ClientFromConfig().List(o.ctx)
}

func (o *Ollama) ListRunning() (*olm.ProcessResponse, error) {
	return api.ClientFromConfig().ListRunning(o.ctx)
}

func (o *Ollama) Generate(requestId, requestStr string) error {
	request := &olm.GenerateRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go api.ClientFromConfig().Generate(o.ctx, request, func(response olm.GenerateResponse) error {
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
	go api.ClientFromConfig().Chat(o.ctx, request, func(response olm.ChatResponse) error {
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
	return api.ClientFromConfig().Delete(o.ctx, request)
}

func (o *Ollama) Show(requestStr string) (*olm.ShowResponse, error) {
	request := &olm.ShowRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return api.ClientFromConfig().Show(o.ctx, request)
}

func (o *Ollama) Embed(requestStr string) (*olm.EmbedResponse, error) {
	request := &olm.EmbedRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return api.ClientFromConfig().Embed(o.ctx, request)
}

func (o *Ollama) Embeddings(requestStr string) (*olm.EmbeddingResponse, error) {
	request := &olm.EmbeddingRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return nil, err
	}
	return api.ClientFromConfig().Embeddings(o.ctx, request)
}

func (o *Ollama) Pull(requestId, requestStr string) error {
	request := &olm.PullRequest{}
	if err := json.Unmarshal([]byte(requestStr), request); err != nil {
		return err
	}
	go api.ClientFromConfig().Pull(o.ctx, request, func(response olm.ProgressResponse) error {
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
	go api.ClientFromConfig().Push(o.ctx, request, func(response olm.ProgressResponse) error {
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
	go api.ClientFromConfig().Create(o.ctx, request, func(response olm.ProgressResponse) error {
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
	return api.ClientFromConfig().Copy(o.ctx, request)
}
