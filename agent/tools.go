package agent

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
)

var ReadFileTool = NewTool(
	"Reads file from filesystem",
	[]ToolArgument{{Name: "fileName", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		dat, err := os.ReadFile(args["fileName"].(string))
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		} else {
			return string(dat)
		}
	},
)

var ListFilesTool = NewTool(
	"Lists files in directory using glob pattern",
	[]ToolArgument{{Name: "dir", Type: "string"}, {Name: "globPattern", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		root := os.DirFS(args["dir"].(string))

		mdFiles, err := fs.Glob(root, args["globPattern"].(string))

		if err != nil {
			log.Fatal(err)
		}

		var files []string
		for _, v := range mdFiles {
			files = append(files, path.Join(args["dir"].(string), v))
		}

		data, err := json.Marshal(files)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		} else {
			return string(data)
		}
	},
)

var WriteFileTool = NewTool(
	"Replaces first occurence of oldString with newString in a file",
	[]ToolArgument{{Name: "fileName", Type: "string"}, {Name: "oldString", Type: "string"}, {Name: "newString", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		fileName := args["fileName"].(string)
		oldString := args["oldString"].(string)
		newString := args["newString"].(string)

		dat, err := os.ReadFile(fileName)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		newDat := bytes.Replace(dat, []byte(oldString), []byte(newString), 1)

		err = os.WriteFile(fileName, newDat, 0644)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		return "OK"
	},
)

var CommandTool = NewTool(
	"Runs shell command",
	[]ToolArgument{{Name: "command", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		command := args["command"].(string)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", command)
		} else {
			cmd = exec.Command("/bin/sh", "-c", command)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		return string(output)
	},
)
