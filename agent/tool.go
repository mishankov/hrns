package agent

type ToolArgument struct {
	Name, Type string
}

type Tool interface {
	Description() string
	Arguments() []ToolArgument
	Call(args map[string]any) string
}

type SimpleTool struct {
	description string
	arguments   []ToolArgument
	function    func(args map[string]any) string
}

func NewSimpleTool(description string, arguments []ToolArgument, function func(args map[string]any) string) SimpleTool {
	return SimpleTool{description, arguments, function}
}

func (t SimpleTool) Description() string {
	return t.description
}

func (t SimpleTool) Arguments() []ToolArgument {
	return t.arguments
}

func (t SimpleTool) Call(args map[string]any) string {
	return t.function(args)
}
