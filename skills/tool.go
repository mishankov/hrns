package skills

import (
	"os"

	"github.com/mishankov/hrns/loop"
	"github.com/mishankov/hrns/tools"
)

type LoadSkillTool struct {
	skills []Skill
}

func NewLoadSkillTool(skills []Skill) *LoadSkillTool {
	return &LoadSkillTool{skills: skills}
}

func (t *LoadSkillTool) Description() string {
	return "Loads full skill body by its name"
}

func (t *LoadSkillTool) Arguments() []loop.ToolArgument {
	return []loop.ToolArgument{
		{Name: "name", Type: "string"},
	}
}

func (t *LoadSkillTool) Call(args map[string]any) string {
	name, err := tools.StringArg(args, "name")
	if err != nil {
		return "ERROR: loading skill argument: " + err.Error()
	}

	for _, skill := range t.skills {
		if skill.Name == name {
			data, err := os.ReadFile(skill.Path)
			if err != nil {
				return "ERROR: loading skill file:" + err.Error()
			}
			return string(data)
		}
	}

	return "ERROR: unknown skill name"
}
