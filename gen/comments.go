package gen

import (
	"regexp"
	"strings"
)

var (
	stripComments = regexp.MustCompile(`^\/\/\s+(.+)$`)
)

func parseComment(lines []string) (map[string]string, error) {
	ret := map[string]string{}

	key := ""
	value := ""

	for _, line := range lines {
		matches := stripComments.FindStringSubmatch(line)
		if len(matches) > 0 {

			if strings.HasPrefix(matches[1], "@") {
				if key != "" {
					ret[key] = value
				}

				key = strings.TrimLeft(matches[1], "@")
				value = ""

			} else {
				value += matches[1] + "\n"
			}
		}
	}

	if value != "" {
		ret[key] = value
	}

	return ret, nil
}
