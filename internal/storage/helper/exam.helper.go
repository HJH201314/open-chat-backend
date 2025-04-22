package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fcraft/open-chat/internal/schema"
	"time"
)

func (q *QueryHelper) GetExam(examId uint64) (*schema.Exam, error) {
	var examCacheKey = fmt.Sprintf("exam:%d", examId)
	// 从 redis 中查询
	cachedExam, err := q.Redis.Get(context.Background(), examCacheKey).Result()
	var exam *schema.Exam
	if err = json.Unmarshal([]byte(cachedExam), exam); err == nil {
		return exam, nil
	}
	// 从数据库中查询
	exam, err = q.GormStore.GetExamWithDetails(examId)
	if err != nil {
		return nil, err
	}
	// 缓存到 redis
	if err = q.Redis.Set(context.Background(), examCacheKey, exam, 1*time.Hour).Err(); err != nil {
	}
	return exam, nil
}
