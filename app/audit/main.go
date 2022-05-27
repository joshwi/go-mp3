package audit

import (
	"regexp"

	"github.com/joshwi/go-pkg/utils"
)

var CONFIG = map[string][]utils.Tag{
	"title": {
		{
			Name:  "_",
			Value: `(?i)\s[\(|\[]feat\..*[\)|\]]`,
		},
	},
}

func Compile(input []utils.Tag) []utils.Match {
	tags := []utils.Match{}
	for _, n := range input {
		r := regexp.MustCompile(n.Value)
		exp := utils.Match{Name: n.Name, Value: *r}
		tags = append(tags, exp)
	}
	return tags
}

func Run(input string, commands []utils.Match) string {
	output := input
	for _, entry := range commands {
		output = entry.Value.ReplaceAllString(output, entry.Name)
	}
	return output
}
