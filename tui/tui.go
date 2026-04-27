package tui

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/mishankov/hrns/agent"
	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/openai"
	"github.com/mishankov/hrns/skills"
)

type TUIApp struct {
	systemPrompt string
	tools        map[string]loop.Tool
	agents       map[string]agent.Agent
	skills       []skills.Skill
	commands     []Command

	config      *Config
	client      *openai.Client
	agenticLoop *loop.Loop
	messages    []openai.Message
	model       string
}

type CommandHandler func(ctx context.Context, args string) error

type Command struct {
	Name        string
	Usage       string
	Description string
	Handler     CommandHandler
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

func WithSkills(skills []skills.Skill) Option {
	return func(app *TUIApp) {
		app.skills = skills
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

	app.initCommands()

	return app
}

func (app *TUIApp) Run(ctx context.Context) {
	if err := app.init(); err != nil {
		PrintError(err.Error())
		return
	}

	switch detectMode(os.Args) {
	case "exec":
		app.runExec(ctx, os.Args[2:])
	default:
		app.runInteractive(ctx)
	}
}

func (app *TUIApp) initCommands() {
	app.commands = []Command{
		{
			Name:        "/new",
			Usage:       "/new",
			Description: "start new session",
			Handler:     app.handleNewCommand,
		},
		{
			Name:        "/models",
			Usage:       "/models",
			Description: "list available models",
			Handler:     app.handleModelsCommand,
		},
		{
			Name:        "/model",
			Usage:       "/model <model>",
			Description: "change model",
			Handler:     app.handleModelCommand,
		},
		{
			Name:        "/providers",
			Usage:       "/providers",
			Description: "list connected providers",
			Handler:     app.handleProvidersCommand,
		},
		{
			Name:        "/provider",
			Usage:       "/provider <name>",
			Description: "change provider",
			Handler:     app.handleProviderCommand,
		},
		{
			Name:        "/agents",
			Usage:       "/agents",
			Description: "list available agents",
			Handler:     app.handleAgentsCommand,
		},
		{
			Name:        "/agent",
			Usage:       "/agent <agent>",
			Description: "change agent",
			Handler:     app.handleAgentCommand,
		},
		{
			Name:        "/connect",
			Usage:       "/connect",
			Description: "connect a new provider",
			Handler:     app.handleConnectCommand,
		},
		{
			Name:        "/help",
			Usage:       "/help",
			Description: "show this help",
			Handler:     app.handleHelpCommand,
		},
	}
}

func (app *TUIApp) init() error {
	if err := app.loadConfig(); err != nil {
		return err
	}

	if err := app.selectInitialAgent(); err != nil {
		return err
	}

	app.initMessages()
	app.initLLM()

	return nil
}

func (app *TUIApp) loadConfig() error {
	config, err := LoadConfig()
	if errors.Is(err, os.ErrNotExist) || (err == nil && len(config.Providers) == 0) {
		PrintHarnessMessage("config file not found, running onboarding now")
		var onboardingErr error
		config, onboardingErr = app.onboarding()
		if onboardingErr != nil {
			return errors.New("failed to run onboarding: " + onboardingErr.Error())
		}
	} else if err != nil {
		return errors.New("failed to load config: " + err.Error())
	}

	app.config = config
	return nil
}

func (app *TUIApp) onboarding() (*Config, error) {
	config := &Config{
		Providers: map[string]ProviderConfig{},
	}

	app.config = config
	err := app.connectProvider()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (app *TUIApp) selectInitialAgent() error {
	if app.config.CurrentAgent != "" {
		if _, ok := app.agents[app.config.CurrentAgent]; ok {
			return nil
		}

		PrintHarnessMessage("configured agent not found: " + app.config.CurrentAgent)
		app.config.CurrentAgent = ""
	}

	for name := range app.agents {
		app.config.CurrentAgent = name
		if err := app.config.Save(); err != nil {
			return errors.New("failed to save config: " + err.Error())
		}
		PrintHarnessMessage("Agent changed to " + name)
		return nil
	}

	return app.config.Save()
}

func (app *TUIApp) initMessages() {
	app.messages = []openai.Message{openai.SystemMessage(app.buildSystemPrompt())}
}

func (app *TUIApp) buildSystemPrompt() string {
	return app.buildSystemPromptForAgent(app.config.CurrentAgent)
}

func (app *TUIApp) buildSystemPromptForAgent(agentName string) string {
	parts := []string{}

	if agentName != "" {
		if selectedAgent, ok := app.agents[agentName]; ok && selectedAgent.Prompt != "" {
			parts = append(parts, selectedAgent.Prompt)
		}
	}

	if len(parts) == 0 && app.systemPrompt != "" {
		parts = append(parts, app.systemPrompt)
	}

	if skillsPrompt := app.buildSkillsPrompt(); skillsPrompt != "" {
		parts = append(parts, skillsPrompt)
	}

	return strings.Join(parts, "\n\n")
}

func (app *TUIApp) buildSkillsPrompt() string {
	if len(app.skills) == 0 {
		return ""
	}

	var prompt strings.Builder
	prompt.WriteString("You have access to the following skills:")
	for _, skill := range app.skills {
		prompt.WriteString("\n- ")
		prompt.WriteString(skill.Name)
		prompt.WriteString(": ")
		prompt.WriteString(skill.Description)
	}
	prompt.WriteString("\nYou can load them with the `load_skill` tool.")

	return prompt.String()
}

func (app *TUIApp) initLLM() {
	currentProvider := app.config.Providers[app.config.CurrentProvider]
	app.model = currentProvider.Model
	app.rebuildAgenticLoop(currentProvider)
}

func detectMode(args []string) string {
	if len(args) > 1 && args[1] == "exec" {
		return "exec"
	}

	return "interactive"
}

func (app *TUIApp) runInteractive(ctx context.Context) {
	PrintWelcomeMessage(app.config.CurrentProvider, app.model)

	for {
		PrintAgent(app.config.CurrentAgent)
		messageText := GetUserInput()

		if app.handleCommand(ctx, messageText) {
			continue
		}

		app.messages = append(app.messages, openai.UserMessage(messageText))

		app.runAgent(ctx)
	}
}

func (app *TUIApp) handleCommand(ctx context.Context, input string) bool {
	commandName, args, ok := parseCommand(input)
	if !ok {
		return false
	}

	command, ok := app.findCommand(commandName)
	if !ok {
		PrintError("unknown command: " + commandName)
		return true
	}

	if err := command.Handler(ctx, args); err != nil {
		PrintError(err.Error())
	}

	return true
}

func parseCommand(input string) (string, string, bool) {
	if !strings.HasPrefix(input, "/") {
		return "", "", false
	}

	commandName, args, _ := strings.Cut(input, " ")
	return commandName, strings.TrimSpace(args), true
}

func (app *TUIApp) findCommand(name string) (Command, bool) {
	for _, command := range app.commands {
		if command.Name == name {
			return command, true
		}
	}

	return Command{}, false
}

func (app *TUIApp) handleNewCommand(ctx context.Context, args string) error {
	app.messages = []openai.Message{openai.SystemMessage(app.buildSystemPrompt())}
	PrintHarnessMessage("New session started")
	PrintWelcomeMessage(app.config.CurrentProvider, app.model)

	return nil
}

func (app *TUIApp) handleModelsCommand(ctx context.Context, args string) error {
	models, err := app.client.ListModels(ctx)
	if err != nil {
		return errors.New("failed to list models: " + err.Error())
	}

	for _, model := range models {
		PrintHarnessMessage(model)
	}

	return nil
}

func (app *TUIApp) handleModelCommand(ctx context.Context, args string) error {
	if args == "" {
		return errors.New("usage: /model <model>")
	}

	previousModel := app.model
	app.model = args

	provider := app.config.Providers[app.config.CurrentProvider]
	provider.Model = app.model
	app.config.Providers[app.config.CurrentProvider] = provider

	if err := app.config.Save(); err != nil {
		app.model = previousModel
		provider.Model = previousModel
		app.config.Providers[app.config.CurrentProvider] = provider
		return errors.New("failed to save config: " + err.Error())
	}

	PrintHarnessMessage("Model changed to " + args)

	return nil
}

func (app *TUIApp) handleProvidersCommand(ctx context.Context, args string) error {
	PrintHarnessMessage("Providers:")
	for name := range app.config.Providers {
		PrintHarnessMessage(name)
	}

	return nil
}

func (app *TUIApp) handleProviderCommand(ctx context.Context, args string) error {
	if args == "" {
		return errors.New("usage: /provider <name>")
	}

	provider, ok := app.config.Providers[args]
	if !ok {
		return errors.New("provider not found: " + args)
	}

	previousProvider := app.config.CurrentProvider
	app.config.CurrentProvider = args

	if err := app.config.Save(); err != nil {
		app.config.CurrentProvider = previousProvider
		return errors.New("failed to save config: " + err.Error())
	}

	app.model = provider.Model
	app.rebuildAgenticLoop(provider)

	PrintHarnessMessage("Provider changed to " + app.config.CurrentProvider)
	PrintHarnessMessage("Model changed to " + app.model)

	return nil
}

func (app *TUIApp) handleAgentsCommand(ctx context.Context, args string) error {
	for name, agent := range app.agents {
		PrintHarnessMessage(name + " - " + agent.Description)
	}

	return nil
}

func (app *TUIApp) handleAgentCommand(ctx context.Context, args string) error {
	if args == "" {
		return errors.New("usage: /agent <agent>")
	}

	if _, ok := app.agents[args]; !ok {
		return errors.New("agent not found: " + args)
	}

	previousAgent := app.config.CurrentAgent
	app.config.CurrentAgent = args
	if err := app.config.Save(); err != nil {
		app.config.CurrentAgent = previousAgent
		return errors.New("failed to save config: " + err.Error())
	}

	app.messages[0] = openai.SystemMessage(app.buildSystemPrompt())

	PrintHarnessMessage("Agent changed to " + args)

	return nil
}

func (app *TUIApp) handleConnectCommand(ctx context.Context, args string) error {
	return app.connectProvider()
}

func (app *TUIApp) handleHelpCommand(ctx context.Context, args string) error {
	PrintHarnessMessage("Available commands:")
	for _, command := range app.commands {
		PrintHarnessMessage(fmt.Sprintf("%-16s - %s", command.Usage, command.Description))
	}

	return nil
}

func (app *TUIApp) connectProvider() error {
	PrintHarnessMessage("Input new provider name:")
	name := GetUserInput()

	PrintHarnessMessage("Input provider API URL:")
	url := GetUserInput()

	PrintHarnessMessage("Input provider API key:")
	key := GetUserInput()

	PrintHarnessMessage("Input default model:")
	model := GetUserInput()

	PrintHarnessMessage("Skip SSL verify? (y/n)")
	skipVerify := GetUserInput() == "y"

	provider := ProviderConfig{
		Url:        url,
		Key:        key,
		Model:      model,
		SkipVerify: skipVerify,
	}
	app.config.Providers[name] = provider
	app.config.CurrentProvider = name

	return app.config.Save()
}

func (app *TUIApp) runExec(ctx context.Context, args []string) {
	execCmd := flag.NewFlagSet("exec", flag.ContinueOnError)
	flagModel := execCmd.String("model", app.model, "Model")
	flagProvider := execCmd.String("provider", app.config.CurrentProvider, "Provider")
	flagMessage := execCmd.String("message", "", "Message")
	flagAgent := execCmd.String("agent", app.config.CurrentAgent, "Agent")

	err := execCmd.Parse(args)
	if err != nil {
		PrintError(err.Error())
		return
	}

	modelProvided := false
	agentProvided := false
	execCmd.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "model":
			modelProvided = true
		case "agent":
			agentProvided = true
		}
	})

	provider, ok := app.config.Providers[*flagProvider]
	if !ok {
		PrintError("provider not found: " + *flagProvider)
		return
	}

	_, ok = app.agents[*flagAgent]
	if *flagAgent != "" && !ok {
		if agentProvided {
			PrintError("agent not found: " + *flagAgent)
			return
		}

		*flagAgent = ""
	}

	app.rebuildAgenticLoop(provider)
	app.model = provider.Model
	if modelProvided {
		app.model = *flagModel
	}
	app.messages = []openai.Message{
		openai.SystemMessage(app.buildSystemPromptForAgent(*flagAgent)),
		openai.UserMessage(*flagMessage),
	}

	PrintWelcomeMessage(*flagProvider, app.model)

	app.runAgent(ctx)
}

func (app *TUIApp) createLLMClient(provider ProviderConfig) *openai.Client {
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

func (app *TUIApp) rebuildAgenticLoop(provider ProviderConfig) {
	app.client = app.createLLMClient(provider)
	app.agenticLoop = loop.New(
		app.client,
		app.tools,
	)
}

func (app *TUIApp) runAgent(ctx context.Context) {
	go app.agenticLoop.RunLoop(ctx, app.messages, app.model)

	lastChunk := loop.Chunk{}
	for chunk := range app.agenticLoop.Chunks() {
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

	app.messages = app.agenticLoop.Messages()
}
