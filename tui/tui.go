package tui

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/openai"
)

type TUIApp struct {
	systemPrompt string
	tools        map[string]loop.Tool
	agents       map[string]agent.Agent
}

type Option func(*TUIApp)

func WithTool(name string, tool loop.Tool) Option {
	return func(app *TUIApp) {
		app.tools[name] = tool
	}
}

func WithTools(tools map[string]loop.Tool) Option {
	return func(app *TUIApp) {
		maps.Copy(app.tools, tools)
	}
}

func WithAgent(agent agent.Agent) Option {
	return func(app *TUIApp) {
		app.agents[agent.Name] = agent
	}
}

func WithAgents(agents []agent.Agent) Option {
	return func(app *TUIApp) {
		for _, agent := range agents {
			app.agents[agent.Name] = agent
		}
	}
}

func WithSystemPrompt(systemPrompt string) Option {
	return func(app *TUIApp) {
		app.systemPrompt = systemPrompt
	}
}

func New(opts ...Option) *TUIApp {
	app := &TUIApp{
		tools:  map[string]loop.Tool{},
		agents: map[string]agent.Agent{},
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *TUIApp) Run(ctx context.Context) {
	// Loading config
	config, err := LoadConfig()
	if errors.Is(err, os.ErrNotExist) || len(config.Providers) == 0 {
		PrintHarnessMessage("config file not found, running onboarding now")
		var onboardingErr error
		config, onboardingErr = Onboarding()
		if onboardingErr != nil {
			PrintError("failed to run onboarding: " + onboardingErr.Error())
			return
		}
	} else if err != nil {
		PrintError("failed to load config: " + err.Error())
		return
	}

	systemPrompt := app.systemPrompt
	if config.CurrentAgent != "" {
		if agent, ok := app.agents[config.CurrentAgent]; ok {
			systemPrompt = agent.Prompt
		}
	} else {
		for name, agent := range app.agents {
			config.CurrentAgent = name
			err := config.Save()
			if err != nil {
				PrintError("failed to save config: " + err.Error())
				return
			}

			systemPrompt = agent.Prompt
			break
		}
	}

	// Initializing messages with system prompt
	messages := []openai.Message{openai.SystemMessage(systemPrompt)}

	// Setting up default provider and model
	currentProvider := config.Providers[config.CurrentProvider]
	model := config.Providers[config.CurrentProvider].Model

	// Detecting mode
	mode := "interactive"
	if len(os.Args) > 1 {
		if os.Args[1] == "exec" {
			mode = "exec"
		}
	}

	switch mode {
	case "interactive":

		client := CreateLLMClient(currentProvider)

		agnt := loop.New(
			client,
			app.tools,
		)

		PrintWelcomeMessage(config.CurrentProvider, model)

		for {
			PrintAgent(config.CurrentAgent)
			messageText := GetUserInput()

			if strings.HasPrefix(messageText, "/") {
				commandSplited := strings.Split(messageText, " ")

				switch commandSplited[0] {
				case "/model":
					previousModel := model
					model = strings.TrimSpace(commandSplited[1])

					provider := config.Providers[config.CurrentProvider]
					provider.Model = model
					config.Providers[config.CurrentProvider] = provider

					err := config.Save()
					if err != nil {
						PrintError("failed to save config: " + err.Error())
						model = previousModel
						break
					}

					PrintHarnessMessage("Model changed to " + model)
				case "/provider":
					newProvider := strings.TrimSpace(strings.Join(commandSplited[1:], " "))
					if _, ok := config.Providers[newProvider]; !ok {
						PrintError("provider not found: " + newProvider)
						break
					}
					previousProvider := config.CurrentProvider
					config.CurrentProvider = newProvider

					err := config.Save()
					if err != nil {
						PrintError("failed to save config: " + err.Error())
						config.CurrentProvider = previousProvider
						break
					}
					model = config.Providers[config.CurrentProvider].Model

					client = CreateLLMClient(config.Providers[config.CurrentProvider])
					agnt = loop.New(
						client,
						app.tools,
					)

					PrintHarnessMessage("Provider changed to " + config.CurrentProvider)
					PrintHarnessMessage("Model changed to " + model)
				case "/new":
					messages = []openai.Message{openai.SystemMessage(app.systemPrompt)}
					PrintHarnessMessage("New session started")
					PrintWelcomeMessage(config.CurrentProvider, model)
				case "/providers":
					PrintHarnessMessage("Providers:")
					for name := range config.Providers {
						PrintHarnessMessage(name)
					}
				case "/models":
					models, err := client.ListModels(ctx)
					if err != nil {
						PrintError("failed to list models: " + err.Error())
						break
					}

					for _, m := range models {
						PrintHarnessMessage(m)
					}

				case "/connect":
					err := ConnectProvider(config)
					if err != nil {
						PrintError("failed to connect provider: " + err.Error())
						break
					}
				case "/help":
					PrintHarnessMessage("Available commands:")
					PrintHarnessMessage("/new           - start new session")
					PrintHarnessMessage("/models        - list available model")
					PrintHarnessMessage("/model <model> - change model")
					PrintHarnessMessage("/providers     - list connected providers")
					PrintHarnessMessage("/provider      - change provider")
					PrintHarnessMessage("/connect       - connect a new provider")
					PrintHarnessMessage("/help          - show this help")

				default:
					PrintError("unknown command: " + commandSplited[0])
				}

				continue
			}

			messages = append(messages, openai.UserMessage(messageText))

			RunAgent(ctx, agnt, messages, model)

			messages = agnt.Messages()
		}
	case "exec":
		execCmd := flag.NewFlagSet("exec", flag.ContinueOnError)
		flagModel := execCmd.String("model", model, "Model")
		flagProvider := execCmd.String("provider", config.CurrentProvider, "Provider")
		flagMessage := execCmd.String("message", "", "Message")
		err := execCmd.Parse(os.Args[2:])
		if err != nil {
			PrintError(err.Error())
			return
		}

		client := CreateLLMClient(config.Providers[*flagProvider])

		agnt := loop.New(
			client,
			app.tools,
		)

		PrintWelcomeMessage(*flagProvider, *flagModel)

		messages = append(messages, openai.UserMessage(*flagMessage))

		RunAgent(ctx, agnt, messages, *flagModel)
	}
}

func CreateLLMClient(provider ProviderConfig) *openai.Client {
	return openai.NewClient(
		openai.WithBaseURL(provider.Url),
		openai.WithAPIKey(provider.Key),
		openai.WithHTTPClient(
			&http.Client{Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: provider.SkipVerify,
				},
			}}),
	)
}

func RunAgent(ctx context.Context, agnt *loop.Loop, messages []openai.Message, model string) {
	go agnt.RunLoop(ctx, messages, model)

	lastChunk := loop.Chunk{}
	for chunk := range agnt.Chunks() {
		toBreak := false

		if chunk.Type != lastChunk.Type {
			if !lastChunk.IsToolChunk() || !chunk.IsToolChunk() {
				PrintNewLine()
			}

			if slices.Contains([]loop.ChunkType{loop.ChunkTypeMessage, loop.ChunkTypeReasoning}, lastChunk.Type) {
				PrintNewLine()
			}
		}

		switch chunk.Type {
		case loop.ChunkTypeMessage:
			PrintResponseChunc(chunk.Text)
		case loop.ChunkTypeReasoning:
			PrintReasoningChunc(chunk.Text)
		case loop.ChunkTypeToolCallStart:
			PrintToolCall(chunk.ToolName, chunk.ToolArgs)
		case loop.ChunkTypeToolCallResult, loop.ChunkTypeToolCallError:
			func() {}()
		case loop.ChunkTypeError:
			PrintError(chunk.Text)
		case loop.ChunkTypeEnd:
			toBreak = true
		default:
			PrintHarnessMessage("other chunk " + string(chunk.Type))
		}

		lastChunk = chunk

		if toBreak {
			break
		}
	}
}
