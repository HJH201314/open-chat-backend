package langchain_utils

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func Test() {
	llm, err := openai.New(
		openai.WithBaseURL("https://ark.cn-beijing.volces.com/api/v3/"),
		openai.WithModel("deepseek-v3-250324"),
		openai.WithToken("276b37d9-312d-407d-8b83-1e9444cc2a43"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(
		ctx, "The first man to walk on the moon",
		llms.WithTemperature(0.6),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
