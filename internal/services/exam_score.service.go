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

		// 这里返回的 score 是 0-100 分
		score, comments, err := s.ScoreProblemSync(ctx, problem.ID, record.UserID, answer.Answer)

		if err != nil {
			record.Answers[i].Status = schema.StatusFailed
			record.Answers[i].Comments = fmt.Sprintf("评分失败: %s", err.Error())
			errorOccurred = true
		} else {
			// 获取在 exam 中的分数
			q, ok := slice.FindBy(
				record.Exam.Problems, func(index int, item schema.ExamProblem) bool {
					return item.ProblemID == problem.ID
				},
			)
			if !ok {
				err = fmt.Errorf("题目信息获取失败")
			}
			// 按比例计算最终分数
			finalScore := uint64(float64(q.Score) * float64(score) / 100)

			record.Answers[i].Score = finalScore
			record.Answers[i].Comments = comments
			record.Answers[i].Status = schema.StatusCompleted
			totalScore += finalScore
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

// ScoreProblemSync 评分单个问题
//
//	Returns:
//		int		分数（0-100，答案完全正确则为 100）
//		string	评价
//		error	错误
func (s *ExamScoreService) ScoreProblemSync(ctx context.Context, problemID uint64, userID uint64, answer any) (uint64, string, error) {
	// 查找题目信息
	var problem schema.Problem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return 0, "", errors.New("题目信息获取失败")
	}

	// 预插入数据
	record := schema.ProblemUserRecord{
		UserID:    userID,
		ProblemID: problemID,
		Answer:    schema.ProblemAnswer{Answer: answer},
		Score:     0,
		Comment:   "",
	}
	if err := s.db.Create(&record).Error; err != nil {
		return 0, "", errors.New("failed to create problem user record")
	}

	// 根据题目类型评分
	var score uint64
	var comments string
	var err error

	switch problem.Type {
	case schema.SingleChoice, schema.MultipleChoice, schema.TrueFalse, schema.FillBlank:
		// 精确匹配评分
		score, comments, err = s.scoreExactMatch(problem, answer)
	case schema.ShortAnswer:
		// 使用大模型评分
		score, comments, err = s.scoreWithAI(ctx, problem, answer)
	default:
		return 0, "", errors.New("不支持的题目类型")
	}

	if err := s.db.Updates(
		&schema.ProblemUserRecord{
			ID:      record.ID,
			Score:   score,
			Comment: comments,
		},
	).Error; err != nil {
		return 0, "", errors.New("failed to update problem user record")
	}

	return score, comments, err
}

// ScoreProblemAsync 评分单个问题
//
//	Returns:
//		recordId	分数（0-100，答案完全正确则为 100）
//		error		错误
func (s *ExamScoreService) ScoreProblemAsync(ctx context.Context, problemID uint64, userID uint64, answer any) (uint64, error) {
	// 查找题目信息
	var problem schema.Problem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return 0, errors.New("题目信息获取失败")
	}

	// 预插入数据
	record := schema.ProblemUserRecord{
		UserID:    userID,
		ProblemID: problemID,
		Answer:    schema.ProblemAnswer{Answer: answer},
		Score:     0,
		Comment:   "",
	}
	if err := s.db.Create(&record).Error; err != nil {
		return 0, errors.New("failed to create problem user record")
	}

	go func() {
		// 根据题目类型评分
		var score uint64
		var comments string

		switch problem.Type {
		case schema.SingleChoice, schema.MultipleChoice, schema.TrueFalse, schema.FillBlank:
			// 精确匹配评分
			score, comments, _ = s.scoreExactMatch(problem, answer)
		case schema.ShortAnswer:
			// 使用大模型评分
			score, comments, _ = s.scoreWithAI(ctx, problem, answer)
		default:
			return
		}

		if err := s.db.Updates(
			&schema.ProblemUserRecord{
				ID:      record.ID,
				Score:   score,
				Comment: comments,
			},
		).Error; err != nil {
			return
		}
	}()
	return record.ID, nil
}

func InvalidCorrectAnswerFormat() (uint64, string, error) {
	return 0, "标准答案格式错误", errors.New("invalid correct answer format")
}

func InvalidUserAnswerFormat(msg string) (uint64, string, error) {
	return 0, fmt.Sprintf("用户答案格式错误：%s", msg), errors.New("invalid correct answer format")
}

// scoreExactMatch 精确匹配评分（适用于选择题、判断题、填空题）
func (s *ExamScoreService) scoreExactMatch(problem schema.Problem, answer any) (uint64, string, error) {
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

		userOptionIDs, ok := answer.([]any)
		if !ok || len(userOptionIDs) != 1 {
			return InvalidUserAnswerFormat("非数组或选择了多个选项")
		}
		userOptionIntID, err := convertor.ToInt(userOptionIDs[0])
		if err != nil {
			return InvalidUserAnswerFormat("选项 ID 非整数")
		}

		if userOptionIntID == int64(correctOptionID) {
			return 100, "正确", nil
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

		userOptionIDs, ok := answer.([]any)
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

		return 100, "正确", nil

	case schema.TrueFalse:
		// 判断题，比较布尔值
		correctAnswer, ok := problem.Answer.Answer.(bool)
		if !ok {
			return InvalidCorrectAnswerFormat()
		}

		userBool, ok := answer.(bool)
		if !ok {
			return InvalidUserAnswerFormat("非布尔值")
		}

		if userBool == correctAnswer {
			return 100, "正确", nil
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

		userText, ok := answer.(string)
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
			return 100, "正确", nil
		}
		return 0, "未包含所有关键词", nil
	}

	return 0, "不支持的题目类型", errors.New("unsupported problem type")
}

// scoreWithAI 使用AI进行评分（适用于简答题）
func (s *ExamScoreService) scoreWithAI(ctx context.Context, problem schema.Problem, answer any) (uint64, string, error) {
	userAnswer, ok := answer.(string)
	if !ok {
		return 0, "答案格式错误，请输入文本", errors.New("invalid user answer format")
	}

	standardAnswer, ok := problem.Answer.Answer.(string)
	if !ok {
		standardAnswer = "无标准答案" // 如果没有标准答案，设置为空
	}

	// 调用AI接口进行评分
	aiResponse, _, err := BuiltinPresetCompletion(
		ExamScoreShortAnswerPresetName, map[string]string{
			"QUESTION":        problem.Description,
			"STANDARD_ANSWER": standardAnswer,
			"USER_ANSWER":     userAnswer,
		},
	)

	if err != nil {
		return 0, "AI评分失败", fmt.Errorf("AI scoring failed: %w", err)
	}

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
	finalScore := uint64(aiScore)

	return finalScore, comments, nil
}

const (
	ExamScoreShortAnswerPresetName = "tue_exam_score_short_answer"
)

// ExamScoreService 考试评分服务
type ExamScoreService struct {
	db *gorm.DB
}

// NewExamScoreService 创建新的考试评分服务
func NewExamScoreService(db *gorm.DB) *ExamScoreService {
	// 注册题目评分预设
	presetService := GetPresetService()
	presetService.RegisterBuiltinPresetsSimple(
		ExamScoreShortAnswerPresetName,
		"TUE 简答题评分",
		2,
		"你是一个专业的教育评分助手，请客观公正地评价学生答案，并提供有建设性的反馈。",
		[]chat_utils.Message{
			chat_utils.UserMessage(
				`你是一个专业的教育评分助手。请评估以下答案的质量和准确性。

问题: {QUESTION}

标准答案: {STANDARD_ANSWER}

学生答案: {USER_ANSWER}

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
<suggestion>改进建议</suggestion>`,
			),
		},
	)
	return &ExamScoreService{db: db}
}
