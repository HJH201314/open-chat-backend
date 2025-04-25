// Package services Question 出题服务
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"github.com/openai/openai-go"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func (s *MakeQuestionService) MakeQuestion(problemType schema.ProblemType, problemDescription string) (*schema.Problem, error) {
	presetName := "tue_make_question_" + string(problemType)
	paramMap := map[string]string{
		"TOPIC": problemDescription,
	}
	// 开始执行生成题目
	preset := GetPresetService().GetBuiltinPreset(presetName)
	if preset == nil {
		return nil, fmt.Errorf("preset %s not found", presetName)
	}
	completion, recordId, err := BuiltinPresetCompletion(presetName, paramMap)
	if err != nil {
		// 生成失败
		return nil, err
	}
	// 解析题目
	problem, err := ParseProblemFromCompletion(completion)
	if err != nil {
		// 解析失败
		return nil, err
	}
	if err := s.Gorm.Create(&problem).Error; err != nil {
		return nil, err
	}
	s.Gorm.Create(
		&schema.ProblemMakeRecord{
			PresetCompletionRecordID: recordId,
			ProblemID:                problem.ID,
		},
	)
	return problem, nil
}

func ParseProblemFromCompletion(completion string) (*schema.Problem, error) {
	// 解析题目、答案、解释
	question := chat_utils.ExtractTagContent(completion, "question")
	if question == "" {
		return nil, fmt.Errorf("failed to extract question")
	}

	answersStr := chat_utils.ExtractTagContent(completion, "answers")
	if answersStr == "" {
		return nil, fmt.Errorf("failed to extract answers")
	}

	explanation := chat_utils.ExtractTagContent(completion, "explanation")
	if explanation == "" {
		return nil, fmt.Errorf("failed to extract explanation")
	}

	// 根据类型进行 parse
	var problemType schema.ProblemType
	var answer schema.ProblemAnswer
	var options []schema.ProblemOption

	//  判断是否为选择题
	var choiceOptions []schema.ProblemOption
	if err := json.Unmarshal([]byte(answersStr), &choiceOptions); err == nil {
		// 是选择题！
		if len(choiceOptions) > 0 {
			var correctAnswers []uint64
			// Convert to ProblemOption format
			for index, opt := range choiceOptions {
				options = append(
					options, schema.ProblemOption{
						ID:      uint(index + 1),
						Content: opt.Content,
						Correct: opt.Correct,
					},
				)
				if opt.Correct {
					correctAnswers = append(correctAnswers, uint64(opt.ID))
				}
			}
			// Set answer
			answer = schema.ProblemAnswer{
				Answer: correctAnswers,
			}
			if len(correctAnswers) > 1 {
				problemType = schema.MultipleChoice
			} else {
				problemType = schema.SingleChoice
			}
		}
	} else {
		// Try to parse as true/false
		if strings.EqualFold(strings.TrimSpace(answersStr), "true") || strings.EqualFold(
			strings.TrimSpace(answersStr),
			"false",
		) {
			problemType = schema.TrueFalse
			answer = schema.ProblemAnswer{
				Answer: strings.EqualFold(answersStr, "true"),
			}
		} else {
			// Try to parse as fill blank (array of strings)
			var fillBlankAnswers []string
			if err := json.Unmarshal([]byte(answersStr), &fillBlankAnswers); err == nil {
				problemType = schema.FillBlank
				answer = schema.ProblemAnswer{
					Answer: fillBlankAnswers,
				}
			} else {
				// Default to short answer
				problemType = schema.ShortAnswer
				answer = schema.ProblemAnswer{
					Answer: answersStr,
				}
			}
		}
	}

	// Create the problem
	problem := &schema.Problem{
		Type:        problemType,
		Description: question,
		Options:     options,
		Answer:      answer,
		Explanation: explanation,
		Difficulty:  3, // Default difficulty
	}

	return problem, nil
}

func GetQuestionTools() []chat_utils.CompletionTool {
	return []chat_utils.CompletionTool{
		MakeExamTool(),
		MakeQuestionTool(schema.SingleChoice),
		MakeQuestionTool(schema.MultipleChoice),
		MakeQuestionTool(schema.TrueFalse),
		MakeQuestionTool(schema.ShortAnswer),
		MakeQuestionTool(schema.FillBlank),
		EveryDayQuestionTool(),
	}
}

// MakeQuestionTool 生成题目 Tool Call 工具
func MakeQuestionTool(problemType schema.ProblemType) chat_utils.CompletionTool {
	return chat_utils.CompletionTool{
		Param: openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name: fmt.Sprintf("gen_%s_question", problemType),
				Description: openai.String(
					fmt.Sprintf(
						"When the user required to generate a %s question, this tool can generate it.",
						strings.Replace(string(problemType), "_", " ", 1),
					),
				),
				Parameters: openai.FunctionParameters{
					"type":        "object",
					"description": "",
					"properties": map[string]interface{}{
						"description": map[string]string{
							"type":        "string",
							"description": "the description of the question",
						},
					},
					"required": []string{"description"},
				},
			},
		},
		UserTip: "生成题目中...",
		Handler: func(args ...interface{}) (*chat_utils.CompletionToolHandlerReturn, error) {
			if len(args) == 0 {
				return nil, nil
			}
			params := struct {
				Topic string `json:"description"`
			}{}
			err := json.Unmarshal([]byte(args[0].(string)), &params)
			if err != nil {
				return nil, err
			}
			question, err := GetMakeQuestionService().MakeQuestion(problemType, params.Topic)
			if err != nil {
				return nil, err
			}
			return &chat_utils.CompletionToolHandlerReturn{
				Data:           question,
				ReplaceMessage: "好的，下面是你要的题目～",
				Type:           "question",
			}, nil
		},
	}
}

// EveryDayQuestionTool 获取每日题目 tool call 工具
func EveryDayQuestionTool() chat_utils.CompletionTool {
	return chat_utils.CompletionTool{
		Param: openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        "get_every_day_problem",
				Description: openai.String("When the user want to do an every day problem, this tool will send it"),
				Parameters: openai.FunctionParameters{
					"type":        "object",
					"description": "",
					"properties":  map[string]interface{}{},
					"required":    []string{},
				},
			},
		},
		UserTip: "获取题目中...",
		Handler: func(args ...interface{}) (*chat_utils.CompletionToolHandlerReturn, error) {
			var problem schema.Problem

			if err := GetMakeQuestionService().Gorm.Order("RANDOM()").Limit(1).First(&problem).Error; err != nil {
				return nil, err
			}
			return &chat_utils.CompletionToolHandlerReturn{
				Data:           problem,
				ReplaceMessage: "好的，下面是你要的题目～",
				Type:           "question",
			}, nil
		},
	}
}

// MakeExamTool 生成 Exam Tool Call 工具
func MakeExamTool() chat_utils.CompletionTool {
	return chat_utils.CompletionTool{
		Param: openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        "generate_an_exam",
				Description: openai.String("When the user required to generate a exam, this tool can generate a exam."),
				Parameters: openai.FunctionParameters{
					"type":        "object",
					"description": "",
					"properties": map[string]interface{}{
						"topic": map[string]string{
							"type":        "string",
							"description": "the topic of the exam",
						},
						"description": map[string]string{
							"type":        "string",
							"description": "the description of the exam",
						},
						"count": map[string]string{
							"type":        "string",
							"description": "the count of problems in the exam paper, between 5 ~ 10",
						},
					},
					"required": []string{"topic", "description", "count"},
				},
			},
		},
		UserTip: "生成测验中...",
		Handler: func(args ...interface{}) (*chat_utils.CompletionToolHandlerReturn, error) {
			if len(args) == 0 {
				return nil, nil
			}
			params := struct {
				Topic       string `json:"topic"`
				Description string `json:"description"`
				Count       string `json:"count"`
			}{}
			err := json.Unmarshal([]byte(args[0].(string)), &params)
			if err != nil {
				return nil, err
			}
			// 定义常量集合
			//problemTypes := []schema.ProblemType{
			//	schema.SingleChoice,
			//	schema.MultipleChoice,
			//	schema.FillBlank,
			//	schema.ShortAnswer,
			//	schema.TrueFalse,
			//}

			countInt, err := convertor.ToInt(params.Count)
			if err != nil {
				return nil, err
			}

			// 并发生成题目
			var questions []schema.Problem
			var wg sync.WaitGroup
			var mutex sync.Mutex

			for i := int64(0); i < countInt; i++ {
				wg.Add(1) // 增加 WaitGroup 的计数器
				go func() {
					defer wg.Done() // 协程完成后减少计数器

					// 随机选择一个类型
					//randomIndex := rand.Intn(len(problemTypes))
					//randomProblemType := problemTypes[randomIndex]

					// 生成题目
					question, err := GetMakeQuestionService().MakeQuestion(
						schema.AnyProblemType,
						params.Topic+","+params.Description,
					)
					if err != nil {
						return
					}

					mutex.Lock() // 加锁保护共享资源
					questions = append(questions, *question)
					mutex.Unlock() // 解锁
				}()
			}
			wg.Wait() // 等待所有协程完成

			examProblems := make([]schema.ExamProblem, len(questions))
			for i, question := range questions {
				examProblems[i] = schema.ExamProblem{
					ProblemID: question.ID,
					Score:     1000,
					SortOrder: i,
				}
			}

			// Create the exam
			exam := &schema.Exam{
				Name:        params.Topic,
				Description: params.Description,
				Subjects:    params.Topic, // ignore
				Problems:    examProblems,
				TotalScore:  uint64(len(examProblems) * 1000),
				LimitTime:   120 * len(examProblems),
			}
			if err := GetMakeQuestionService().Gorm.Create(&exam).Error; err != nil {
				return nil, err
			}
			return &chat_utils.CompletionToolHandlerReturn{
				Data:           exam,
				ReplaceMessage: "好的，下面是你要的测验～",
				Type:           "exam",
			}, nil
		},
	}
}

func (s *MakeQuestionService) StartGenerate(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定义常量集合
			problemTypes := []schema.ProblemType{
				schema.SingleChoice,
				schema.MultipleChoice,
				schema.FillBlank,
				schema.ShortAnswer,
				schema.TrueFalse,
			}
			// 随机选择一个类型
			randomIndex := rand.Intn(len(problemTypes))
			randomProblemType := problemTypes[randomIndex]
			_, err := s.MakeQuestion(randomProblemType, "从科学、历史、社会、计算机、哲学、课本知识等类别中任意出题")
			if err != nil {
				// do nothing
			}
		case <-ctx.Done():
			s.Logger.Info("Auto make question stopped")
			return
		}
	}
}

const (
	MakeQuestionSingleChoicePresetName   = "tue_make_question_" + string(schema.SingleChoice)
	MakeQuestionMultipleChoicePresetName = "tue_make_question_" + string(schema.MultipleChoice)
	MakeQuestionTrueFalsePresetName      = "tue_make_question_" + string(schema.TrueFalse)
	MakeQuestionFillBlankPresetName      = "tue_make_question_" + string(schema.FillBlank)
	MakeQuestionShortAnswerPresetName    = "tue_make_question_" + string(schema.ShortAnswer)
	MakeQuestionAnyPresetName            = "tue_make_question_" + string(schema.AnyProblemType)
)

// MakeQuestionService 考试评分服务
type MakeQuestionService struct {
	BaseService
}

var (
	makeQuestionServiceInstance *MakeQuestionService
	makeQuestionServiceOnce     sync.Once
)

// InitMakeQuestionService 初始化新的考试评分服务
func InitMakeQuestionService(base *BaseService) {
	makeQuestionServiceOnce.Do(
		func() {
			makeQuestionServiceInstance = &MakeQuestionService{BaseService: *base}
			// 注册题目评分预设
			presetService := GetPresetService()
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionSingleChoicePresetName,
				"TUE 单选题生成",
				6,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。
题目类型：单选
题目数量：1
题目主题：{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 选项需使用结构化格式，格式：[{"id":1,"content":"北京","correct":true},{"id":2,"content":"上海","correct":false},{"id":3,"content":"广州","correct":false}]。
4. 解析应清晰地说明每个选项正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionMultipleChoicePresetName,
				"TUE 多选题生成",
				6,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。
题目类型：多选
题目数量：1
题目主题：{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 选项需使用结构化格式，格式：[{"id":1,"content":"北京","correct":true},{"id":2,"content":"上海","correct":true},{"id":3,"content":"广州","correct":false}]。
4. 解析应清晰地说明每个选项正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionTrueFalsePresetName,
				"TUE 判断题生成",
				6,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。
题目类型：判断题
题目数量：1
题目主题：{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 答案为 true 或 false。
4. 解析应清晰地说明正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionFillBlankPresetName,
				"TUE 填空题生成",
				6,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。
题目类型：填空题
题目数量：1
题目主题：{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 无论有多少个空，答案需使用数组格式，格式：["北京","上海","广州"]
4. 解析应清晰地进行分析。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionShortAnswerPresetName,
				"TUE 简答题生成",
				6,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。
题目类型：简答题
题目数量：1
题目主题：{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 答案为参考答案，列出答案的要点即可。
4. 解析应清晰地进行分析。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionAnyPresetName,
				"TUE 任意题目生成",
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`
你的任务是根据给定的题目描述进行出题，同时提供答案和解析。
题目类型：根据描述决定，仅可为（单选题、多选题、判断题、填空题、简答题）
题目数量：1
题目描述：{TOPIC}
在出题时，按照题目类型的不同，请遵循以下指南：
1. 题目应与给定的描述相关。
2.1. 单选题/多选题的答案选项需使用结构化 JSON 格式，格式：[{"id":1,"content":"选项 1","correct":true},{"id":2,"content":"选项 2","correct":false},{"id":3,"content":"选项 3","correct":false}]
2.2. 判断题的答案为文本，true 或 false
2.3. 填空题的答案为结构化数组格式：["关键词 1","关键词 2","关键词 3"]
2.4. 简答题的答案为参考答案，列出要点。
3. 解析应清晰地进行分析。
请在<question></question>标签内写下问题，在<answers></answers>标签内写下答案或选项，在<explanation></explanation>标签内写下解析，标签中禁止换行。
`,
					),
				},
			)
		},
	)
}

func GetMakeQuestionService() *MakeQuestionService {
	if makeQuestionServiceInstance == nil {
		panic("PresetService not initialized")
	}
	return makeQuestionServiceInstance
}
