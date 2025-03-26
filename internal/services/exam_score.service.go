package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"strings"

	"gorm.io/gorm"

	"github.com/fcraft/open-chat/internal/schema"
	"github.com/fcraft/open-chat/internal/utils/chat_utils"
)

// ExamScoreService 考试评分服务
type ExamScoreService struct {
	db *gorm.DB
}

// NewExamScoreService 创建新的考试评分服务
func NewExamScoreService(db *gorm.DB) *ExamScoreService {
	return &ExamScoreService{db: db}
}

// ScoreExam 评分整个考试
func (s *ExamScoreService) ScoreExam(ctx context.Context, recordID uint64) error {
	// 查询考试记录
	var record schema.ExamUserRecord
	if err := s.db.Preload("Exam.Problems.Problem").Preload("Answers").First(&record, recordID).Error; err != nil {
		return fmt.Errorf("failed to find exam record: %w", err)
	}

	// 检查评分状态
	if record.Status == schema.StatusScoring {
		return errors.New("exam is already being scored")
	}

	// 更新状态为评分中
	if err := s.db.Model(&record).Update("status", schema.StatusScoring).Error; err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// 执行评分
	var totalScore uint64 = 0
	var errorOccurred bool

	// 处理每个题目的答案
	for i, answer := range record.Answers {
		// 查找题目信息
		var problem schema.Problem
		if err := s.db.First(&problem, answer.ProblemID).Error; err != nil {
			record.Answers[i].Status = schema.StatusFailed
			record.Answers[i].Comments = "题目信息获取失败"
			errorOccurred = true
			continue
		}

		// 根据题目类型评分
		var score uint64
		var comments string
		var status schema.ScoreStatus
		var err error

		switch problem.Type {
		case schema.SingleChoice, schema.MultipleChoice, schema.TrueFalse, schema.FillBlank:
			// 精确匹配评分
			score, comments, err = s.scoreExactMatch(problem, answer)
			status = schema.StatusCompleted
		case schema.ShortAnswer:
			// 使用大模型评分
			score, comments, err = s.scoreWithAI(ctx, problem, answer)
			status = schema.StatusCompleted
		default:
			comments = "不支持的题目类型"
			err = errors.New(comments)
			status = schema.StatusFailed
		}

		if err != nil {
			record.Answers[i].Status = schema.StatusFailed
			record.Answers[i].Comments = fmt.Sprintf("评分失败: %s", err.Error())
			errorOccurred = true
		} else {
			record.Answers[i].Score = score
			record.Answers[i].Comments = comments
			record.Answers[i].Status = status
			totalScore += score
		}
	}

	// 更新评分状态和总分
	finalStatus := schema.StatusCompleted
	if errorOccurred {
		finalStatus = schema.StatusFailed
	}

	if err := s.db.Model(&record).Session(&gorm.Session{FullSaveAssociations: true}).Updates(
		schema.ExamUserRecord{
			Status:     finalStatus,
			TotalScore: totalScore,
			Answers:    record.Answers,
		},
	).Error; err != nil {
		return fmt.Errorf("failed to update exam record: %w", err)
	}

	return nil
}

func InvalidCorrectAnswerFormat() (uint64, string, error) {
	return 0, "标准答案格式错误", errors.New("invalid correct answer format")
}

func InvalidUserAnswerFormat(msg string) (uint64, string, error) {
	return 0, fmt.Sprintf("用户答案格式错误：%s", msg), errors.New("invalid correct answer format")
}

// scoreExactMatch 精确匹配评分（适用于选择题、判断题、填空题）
func (s *ExamScoreService) scoreExactMatch(problem schema.Problem, answer schema.ExamUserRecordAnswer) (uint64, string, error) {
	// 获取问题的满分（从关联表获取）
	var examProblem schema.ExamProblem
	if err := s.db.Where("problem_id = ?", problem.ID).First(&examProblem).Error; err != nil {
		return 0, "", fmt.Errorf("failed to get problem score: %w", err)
	}

	fullScore := examProblem.Score
	userAnswer := answer.Answer

	switch problem.Type {
	case schema.SingleChoice:
		// 单选题，比较选项ID是否匹配
		correctOptionIDs, ok := problem.Answer.Answer.([]any)
		if !ok || len(correctOptionIDs) != 1 {
			return InvalidCorrectAnswerFormat()
		}
		correctOptionID, ok := correctOptionIDs[0].(float64)
		if !ok {
			return InvalidCorrectAnswerFormat()
		}

		userOptionIDs, ok := userAnswer.([]any)
		if !ok || len(userOptionIDs) != 1 {
			return InvalidUserAnswerFormat("非数组或选择了多个选项")
		}
		userOptionIntID, err := convertor.ToInt(userOptionIDs[0])
		if err != nil {
			return InvalidUserAnswerFormat("选项 ID 非整数")
		}

		if userOptionIntID == int64(correctOptionID) {
			return fullScore, "正确", nil
		}
		return 0, "错误", nil

	case schema.MultipleChoice:
		// 多选题，比较选项ID数组
		correctOptionIDs, ok := problem.Answer.Answer.([]any)
		if !ok || len(correctOptionIDs) == 0 {
			return InvalidCorrectAnswerFormat()
		}
		// 正确答案集合
		var correctOptionsSet = make(map[int64]bool)
		for _, id := range correctOptionIDs {
			// 先粗略转换是否为 float64 类型
			if _, ok := id.(float64); !ok {
				return InvalidCorrectAnswerFormat()
			} else {
				idInt, err := convertor.ToInt(id)
				if err != nil {
					return InvalidCorrectAnswerFormat()
				}
				correctOptionsSet[idInt] = true
			}
		}

		userOptionIDs, ok := userAnswer.([]any)
		if !ok || len(userOptionIDs) == 0 {
			return InvalidUserAnswerFormat("非数组或未选择")
		}
		userOptionsSet := make(map[int64]bool)
		for _, id := range userOptionIDs {
			// 先粗略转换是否为 float64 类型
			if _, ok := id.(float64); !ok {
				return InvalidUserAnswerFormat("选项 ID 非数字")
			} else {
				idInt, err := convertor.ToInt(id)
				if err != nil {
					return InvalidUserAnswerFormat("选项 ID 非整数")
				}
				userOptionsSet[idInt] = true
			}
		}

		// 检查数量是否相同
		if len(correctOptionsSet) != len(userOptionsSet) {
			return 0, "选择的答案数量不正确", nil
		}

		// 检查是否完全匹配
		for id := range correctOptionsSet {
			if !userOptionsSet[id] {
				return 0, "选择的答案不完全正确", nil
			}
		}

		return fullScore, "正确", nil

	case schema.TrueFalse:
		// 判断题，比较布尔值
		correctAnswer, ok := problem.Answer.Answer.(bool)
		if !ok {
			return InvalidCorrectAnswerFormat()
		}

		userBool, ok := userAnswer.(bool)
		if !ok {
			return InvalidUserAnswerFormat("非布尔值")
		}

		if userBool == correctAnswer {
			return fullScore, "正确", nil
		}
		return 0, "错误", nil

	case schema.FillBlank:
		// 填空题，比较关键词 TODO: 支持判断模式（正则、关键词、完全匹配）等
		var correctKeywords []string
		switch problem.Answer.Answer.(type) {
		case []any:
			keywords := make([]string, len(problem.Answer.Answer.([]any)))
			slice.ForEachWithBreak(
				problem.Answer.Answer.([]any), func(_ int, item any) bool {
					keyword := convertor.ToString(item)
					if keyword == "" {
						return false
					}
					keywords = append(keywords, keyword)
					return true
				},
			)
			if len(keywords) == 0 {
				return InvalidCorrectAnswerFormat()
			}
			correctKeywords = keywords
		case string:
			correctKeywords = []string{problem.Answer.Answer.(string)}
		}
		if len(correctKeywords) == 0 {
			return InvalidCorrectAnswerFormat()
		}

		userText, ok := userAnswer.(string)
		if !ok {
			return InvalidUserAnswerFormat("非文本")
		}

		// 检查每个关键词是否存在
		allMatched := true
		for _, keyword := range correctKeywords {
			if !strings.Contains(userText, keyword) {
				allMatched = false
				break
			}
		}

		if allMatched {
			return fullScore, "正确", nil
		}
		return 0, "未包含所有关键词", nil
	}

	return 0, "不支持的题目类型", errors.New("unsupported problem type")
}

// scoreWithAI 使用AI进行评分（适用于简答题）
func (s *ExamScoreService) scoreWithAI(ctx context.Context, problem schema.Problem, answer schema.ExamUserRecordAnswer) (uint64, string, error) {
	// 获取问题的满分
	var examProblem schema.ExamProblem
	if err := s.db.Where("problem_id = ?", problem.ID).First(&examProblem).Error; err != nil {
		return 0, "", fmt.Errorf("failed to get problem score: %w", err)
	}

	fullScore := examProblem.Score
	userAnswer, ok := answer.Answer.(string)
	if !ok {
		return 0, "答案格式错误，请输入文本", errors.New("invalid user answer format")
	}

	standardAnswer, ok := problem.Answer.Answer.(string)
	if !ok {
		standardAnswer = "无标准答案" // 如果没有标准答案，设置为空
	}

	// 构建评分提示
	prompt := fmt.Sprintf(
		`你是一个专业的教育评分助手。请评估以下答案的质量和准确性。

问题: %s

标准答案: %s

学生答案: %s

请根据以下标准进行评分:
1. 内容准确性: 答案是否准确，与标准答案相符
2. 完整性: 是否涵盖了问题的所有方面
3. 清晰度: 表达是否清晰

请提供:
1. 分数评估(满分100分)
2. 简短评语(不超过100字)
3. 改进建议(如果有)

最后，按照以下 tag 格式返回:
<score>分数(0-100之间的整数)</score>
<comment>评语</comment>
<suggestion>改进建议</suggestion>
`, problem.Description, standardAnswer, userAnswer,
	)

	// 查询配置，获取默认的AI模型提供商 TODO：目前临时使用 deepseek-v3，后续更新可配置
	var model schema.Model
	if err := s.db.Model(&model).Preload("Provider").Preload("Provider.APIKeys").Where(
		"id = ?",
		4,
	).First(&model).Error; err != nil {
		return 0, "", fmt.Errorf("failed to get default AI provider: %w", err)
	}

	// 调用AI接口进行评分
	resp, err := chat_utils.Completion(
		ctx, chat_utils.GetCommonCompletionOptions(
			&model, &chat_utils.CompletionOptions{
				CompletionModelConfig: chat_utils.CompletionModelConfig{
					MaxTokens:   1000, // 输出长度限制
					Temperature: 0.3,  // 较低的温度，提高一致性
				},
				SystemPrompt: "你是一个专业的教育评分助手，请客观公正地评价学生答案，并提供有建设性的反馈。",
				Messages: []chat_utils.Message{
					{
						Role:    "user",
						Content: prompt,
					},
				},
			},
		),
	)

	if err != nil {
		return 0, "AI评分失败", fmt.Errorf("AI scoring failed: %w", err)
	}

	// 解析AI响应（这里简化处理，实际应解析JSON）
	aiResponse := resp.Content

	// 简单解析，尽量规避 JSON 格式异常
	scoreStr := strutil.SubInBetween(aiResponse, "<score>", "</score>")
	comments := strutil.SubInBetween(aiResponse, "<comment>", "</comment>")
	suggestions := strutil.SubInBetween(aiResponse, "<suggestion>", "</suggestion>")

	// 移除引号并添加标识
	comments = fmt.Sprintf("AI:%s:%s", comments, suggestions)

	// 转换分数
	var aiScore float64
	_, err = fmt.Sscanf(scoreStr, "%f", &aiScore)
	if err != nil {
		aiScore = 60 // 默认分数
	}

	// 按比例计算最终分数
	finalScore := uint64(float64(fullScore) * aiScore / 100)

	return finalScore, comments, nil
}
