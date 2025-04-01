package chat_utils

import (
	"fmt"
	"regexp"
	"strings"
)

func ExtractTagContent(content, tag string) string {
	// 移除空标签
	content = strings.ReplaceAll(content, fmt.Sprintf("<%s></%s>", tag, tag), "")
	// 正则匹配
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

func TestExtractTagContent() {
	content := `<question></question>  
<answers></answers>  
<explanation></explanation>  

请按照要求重新提供题目内容，我将为您生成符合规范的填空题。  

（注：由于您未提供具体题目内容，以下为示例模板）  

<question>水在标准大气压下的沸点是______℃，其化学式是______。</question>  
<answers>["100","H2O"]</answers>  
<explanation>1. 标准大气压下水的沸点为100℃，这是物理常识；2. 水的化学式由2个氢原子和1个氧原子组成，写作H2O。</explanation>  

请补充具体题目要求或主题方向，我将为您定制更精准的题目。`
	var result string
	result = ExtractTagContent(content, "question")
	fmt.Println("Extracted content:", result)
	result = ExtractTagContent(content, "answers")
	fmt.Println("Extracted content:", result)
	result = ExtractTagContent(content, "explanation")
	fmt.Println("Extracted content:", result)
}
