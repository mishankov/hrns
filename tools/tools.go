package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/mishankov/hrns/agent"
)

var ReadFileTool = agent.NewSimpleTool(
	"Reads file from filesystem",
	[]agent.ToolArgument{{Name: "fileName", Type: "string"}},
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

var ListFilesTool = agent.NewSimpleTool(
	"Lists files in directory using glob pattern",
	[]agent.ToolArgument{{Name: "dir", Type: "string"}, {Name: "globPattern", Type: "string"}},
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

var WriteFileTool = agent.NewSimpleTool(
	"Replaces first occurence of oldString with newString in a file",
	[]agent.ToolArgument{{Name: "fileName", Type: "string"}, {Name: "oldString", Type: "string"}, {Name: "newString", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		fileName := args["fileName"].(string)
		oldString := args["oldString"].(string)
		newString := args["newString"].(string)

		dat, err := os.ReadFile(fileName)
		if errors.Is(err, fs.ErrNotExist) {
			// Create file if it doesn't exist
			err = os.WriteFile(fileName, []byte(newString), 0644)
			if err != nil {
				return "ERROR: tools calling error: " + err.Error()
			}
			return "OK"
		} else if err != nil {
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

var CommandTool = agent.NewSimpleTool(
	"Runs shell command",
	[]agent.ToolArgument{{Name: "command", Type: "string"}},
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

var WebFetchTool = agent.NewSimpleTool(
	"Fetches content from URL",
	[]agent.ToolArgument{{Name: "url", Type: "string"}},
	func(args map[string]any) string {
		// TODO: make safe type assertions
		url := args["url"].(string)

		resp, err := http.Get(url)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		return string(body)
	},
)
