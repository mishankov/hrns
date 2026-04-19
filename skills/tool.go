package skills

import (
	"os"

	"github.com/mishankov/hrns/agent"
)

type LoadSkillTool struct {
	skills []Skill
}

func NewLoadSkillTool(skills []Skill) *LoadSkillTool {
	return &LoadSkillTool{skills: skills}
}

func (t *LoadSkillTool) Description() string {
	return "Loads full skill body"
}

func (t *LoadSkillTool) Arguments() []agent.ToolArgument {
	return []agent.ToolArgument{
		{Name: "pathToSkill", Type: "string"},
	}
}

func (t *LoadSkillTool) Call(args map[string]any) string {
	// TODO: make safe type assertions
	pathToSkill := args["fileName"].(string)

	data, err := os.ReadFile(pathToSkill)
	if err != nil {
		return "ERROR: loading skill file:" + err.Error()
	}

	return string(data)
}
