package main

import (
	"fmt"

	"github.com/mishankov/hrns/skills"
)

func main() {
	files, err := skills.DisocoverSkillFile(skills.DefaultRootPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		data, err := skills.GetSkillData(file)

		fmt.Println(data, err)
	}
}
