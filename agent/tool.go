package agent

type ToolArgument struct {
	Name, Type string
}

type Tool struct {
	description string
	arguments   []ToolArgument
	function    func(args any) string
}
