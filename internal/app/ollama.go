package app

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"ollama-desktop/internal/config"
	"ollama-desktop/internal/log"
	"ollama-desktop/internal/ollama"
	"ollama-desktop/internal/ollama/api"
	"ollama-desktop/internal/ollama/cmd"
	"os"
	gorun "runtime"
	"strings"
)

func (a *App) OllamaHost() string {
	return config.Config.Ollama.Host.String()
}

func (a *App) OllamaEnvs() []OllamaEnvVar {
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

func (a *App) OllamaVersion() (string, error) {
	return api.ClientFromConfig().Version(a.ctx)
}

func (a *App) OllamaHeartbeat() {
	var installed, started bool
	client := api.ClientFromConfig()

	started = client.Heartbeat(a.ctx) == nil

	if !started {
		installed, _ = cmd.CheckInstalled(a.ctx)
	} else {
		installed = true
	}
	os := gorun.GOOS
	runtime.EventsEmit(a.ctx, "ollamaHeartbeat", installed, started, !started && installed && (os == "windows" || os == "darwin"))
}

func (a *App) StartOllama() error {
	client := api.ClientFromConfig()

	err := cmd.StartApp(a.ctx, client)
	if err != nil {
		log.Error().Err(err).Msg("Ollama StartApp")
		return err
	}
	a.OllamaHeartbeat()
	return nil
}

func (a *App) OllamaList() (*ollama.ListResponse, error) {
	return api.ClientFromConfig().List(a.ctx)
}

func (a *App) OllamaListRunning() (*ollama.ProcessResponse, error) {
	return api.ClientFromConfig().ListRunning(a.ctx)
}

func (a *App) OllamaGenerate(request *ollama.GenerateRequest, fn api.GenerateResponseFunc) error {
	return api.ClientFromConfig().Generate(a.ctx, request, fn)
}

func (a *App) OllamaChat(request *ollama.ChatRequest, fn api.ChatResponseFunc) error {
	return api.ClientFromConfig().Chat(a.ctx, request, fn)
}

func (a *App) OllamaPull(model string) error {
	return api.ClientFromConfig().Pull(a.ctx, &ollama.PullRequest{
		Model: model,
	}, func(response ollama.ProgressResponse) error {
		runtime.EventsEmit(a.ctx, "ollamaPull", response)
		return nil
	})
}

func (a *App) OllamaDelete(model string) error {
	return api.ClientFromConfig().Delete(a.ctx, &ollama.DeleteRequest{
		Model: model,
	})
}

func (a *App) OllamaShow(request *ollama.ShowRequest) (*ollama.ShowResponse, error) {
	return api.ClientFromConfig().Show(a.ctx, request)
}

func (a *App) OllamaEmbed(request *ollama.EmbedRequest) (*ollama.EmbedResponse, error) {
	return api.ClientFromConfig().Embed(a.ctx, request)
}

func (a *App) OllamaEmbeddings(request *ollama.EmbeddingRequest) (*ollama.EmbeddingResponse, error) {
	return api.ClientFromConfig().Embeddings(a.ctx, request)
}
