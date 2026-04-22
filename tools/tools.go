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
	"path/filepath"
	"runtime"
	"time"

	"github.com/mishankov/hrns/loop"
)

var ReadFileTool = loop.NewSimpleTool(
	"Reads file from filesystem",
	[]loop.ToolArgument{{Name: "fileName", Type: "string"}},
	func(args map[string]any) string {
		fileName, err := StringArg(args, "fileName")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		dat, err := os.ReadFile(fileName)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		} else {
			return string(dat)
		}
	},
)

var ListFilesTool = loop.NewSimpleTool(
	"Lists files in directory using glob pattern",
	[]loop.ToolArgument{{Name: "dir", Type: "string"}, {Name: "globPattern", Type: "string"}},
	func(args map[string]any) string {
		dir, err := StringArg(args, "dir")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		globPattern, err := StringArg(args, "globPattern")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		root := os.DirFS(dir)

		mdFiles, err := fs.Glob(root, globPattern)

		if err != nil {
			log.Fatal(err)
		}

		var files []string
		for _, v := range mdFiles {
			files = append(files, filepath.Join(dir, v))
		}

		data, err := json.Marshal(files)
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		} else {
			return string(data)
		}
	},
)

var WriteFileTool = loop.NewSimpleTool(
	"Replaces first occurence of oldString with newString in a file",
	[]loop.ToolArgument{{Name: "fileName", Type: "string"}, {Name: "oldString", Type: "string"}, {Name: "newString", Type: "string"}},
	func(args map[string]any) string {
		fileName, err := StringArg(args, "fileName")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		oldString, err := StringArg(args, "oldString")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		newString, err := StringArg(args, "newString")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

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

var CommandTool = loop.NewSimpleTool(
	"Runs shell command",
	[]loop.ToolArgument{{Name: "command", Type: "string"}},
	func(args map[string]any) string {
		command, err := StringArg(args, "command")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

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

var WebFetchTool = loop.NewSimpleTool(
	"Fetches content from URL",
	[]loop.ToolArgument{{Name: "url", Type: "string"}},
	func(args map[string]any) string {
		url, err := StringArg(args, "url")
		if err != nil {
			return "ERROR: tools calling error: " + err.Error()
		}

		client := http.Client{
			Timeout: 1 * time.Minute,
		}

		resp, err := client.Get(url)
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
