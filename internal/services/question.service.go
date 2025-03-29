// Package services Question 出题服务
package services

import (
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
	"sync"
)

func (s *MakeQuestionService) MakeQuestion(problemType schema.ProblemType, problemDescription string) (uint64, error) {
	presetName := "tue_make_question_" + string(problemType)
	paramMap := map[string]string{
		"TOPIC": problemDescription,
	}
	// 开始执行生成题目
	preset := GetPresetService().GetBuiltinPreset(presetName)
	if preset == nil {
		return 0, fmt.Errorf("preset %s not found", presetName)
	}
	go func() {
		completion, recordId, err := BuiltinPresetCompletion(presetName, paramMap)
		if err != nil {
			// 生成失败
			return
		}
		// 解析题目
		s.Logger.Info("make question completion: %s", completion)
		s.Gorm.Create(
			&schema.ProblemMakeRecord{
				PresetCompletionRecordID: recordId,
				ProblemID:                0,
			},
		)
	}()
	return 0, nil
}

func ParseProblemFromCompletion(completion string) (*schema.Problem, error) {
	// 解析题目
	return nil, nil
}

const (
	MakeQuestionSingleChoicePresetName   = "tue_make_question_" + string(schema.SingleChoice)
	MakeQuestionMultipleChoicePresetName = "tue_make_question_" + string(schema.MultipleChoice)
	MakeQuestionTrueFalsePresetName      = "tue_make_question_" + string(schema.TrueFalse)
	MakeQuestionFillBlankPresetName      = "tue_make_question_" + string(schema.FillBlank)
	MakeQuestionShortAnswerPresetName    = "tue_make_question_" + string(schema.ShortAnswer)
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
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`你是一个出题专家，请根据要求和客观事实输出高质量题目。

请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
题目类型：单选
题目主题：
{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 选项需使用结构化格式，格式：[{"id":1,"content":"北京","correct":true},{"id":2,"content":"上海","correct":false},{"id":3,"content":"广州","correct":false}]。
4. 解析应清晰地说明每个选项正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionMultipleChoicePresetName,
				"TUE 多选题生成",
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`你是一个出题专家，请根据要求和客观事实输出高质量题目。

你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。请仔细阅读以下信息，并按照指示完成任务。
题目类型：多选
题目主题：
{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 选项需使用结构化格式，格式：[{"id":1,"content":"北京","correct":true},{"id":2,"content":"上海","correct":true},{"id":3,"content":"广州","correct":false}]。
4. 解析应清晰地说明每个选项正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionTrueFalsePresetName,
				"TUE 判断题生成",
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`你是一个出题专家，请根据要求和客观事实输出高质量题目。

请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
题目类型：判断题
题目主题：
{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 答案为 true 或 false。
4. 解析应清晰地说明正确或错误的原因。
请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionFillBlankPresetName,
				"TUE 填空题生成",
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`你是一个出题专家，请根据要求和客观事实输出高质量题目。

你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。请仔细阅读以下信息，并按照指示完成任务。
题目类型：填空题
题目主题：
{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 无论有多少个空，答案需使用数组格式，格式：["北京","上海","广州"]
4. 解析应清晰地进行分析。
请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
`,
					),
				},
			)
			presetService.RegisterBuiltinPresetsSimple(
				MakeQuestionShortAnswerPresetName,
				"TUE 简答题生成",
				1,
				"你是一个出题专家，请根据要求和客观事实输出高质量题目。",
				[]chat_utils.Message{
					chat_utils.UserMessage(
						`你是一个出题专家，请根据要求和客观事实输出高质量题目。

你的任务是根据给定的题目类型和主题进行出题，同时提供答案和解析。请仔细阅读以下信息，并按照指示完成任务。
题目类型：简答题
题目主题：
{TOPIC}
在出题时，请遵循以下指南：
1. 根据题目类型出题。
2. 题目应与给定的主题相关。
3. 答案为参考答案，列出答案的要点即可。
4. 解析应清晰地进行分析。
请在<question></question>标签内写下问题，在<answers></answer>标签内写下答案，在<explanation></explanation>标签内写下解析。
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
