package constants

// CommonStatus 处理状态
type CommonStatus string

const (
	StatusPending   CommonStatus = "pending"   // 待评分
	StatusHandling  CommonStatus = "handling"  // 评分中
	StatusCompleted CommonStatus = "completed" // 评分完成
	StatusFailed    CommonStatus = "failed"    // 评分失败
)
