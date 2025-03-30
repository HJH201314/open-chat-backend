package chat_utils

import (
	"fmt"
	"regexp"
)

func ExtractTagContent(content, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`<%s>(.*?)</%s>(?s:.*)$`, tag, tag))
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		fmt.Println("Last answer:", matches, matches[1])
		return matches[1]
	} else {
		fmt.Println("No match found")
		return ""
	}
}
