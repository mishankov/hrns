package tools

import "fmt"

func StringArg(args map[string]any, name string) (string, error) {
	value, ok := args[name]
	if !ok {
		return "", fmt.Errorf("missing %q argument", name)
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("argument %q must be a string", name)
	}

	return stringValue, nil
}
