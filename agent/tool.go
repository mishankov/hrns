package agent

type ToolArgument struct {
	Name, Type string
}

type Tool struct {
	description string
	arguments   []ToolArgument
	function    func(args map[string]any) string
}

func NewTool(description string, arguments []ToolArgument, function func(args map[string]any) string) Tool {
	return Tool{description, arguments, function}
}
