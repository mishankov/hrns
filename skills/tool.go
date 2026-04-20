package skills

import (
	"os"

	"github.com/mishankov/hrns/loop"
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
	// TODO: make safe type assertions
	name := args["name"].(string)

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
