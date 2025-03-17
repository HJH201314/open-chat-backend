package entity

import (
	"fmt"
	"slices"
	"strings"
)

type SortParam struct {
	SortExpr       string `json:"sort_expr" form:"sort_expr"`
	StabilityField string `json:"-"` // 稳定字段，由业务层注入
	DefaultOrder   string `json:"-"` // 默认排序，由业务层注入
}

func (p *SortParam) WithDefault(defaultOrder string, stabilityFiled string) *SortParam {
	p.DefaultOrder = defaultOrder
	p.StabilityField = stabilityFiled
	return p
}

func (p *SortParam) SafeExpr(whitelist []string) string {
	// 初始化稳定性字段
	if p.StabilityField == "" {
		p.StabilityField = "id" // 默认值
	}

	// 处理空值
	if p.SortExpr == "" {
		if p.DefaultOrder != "" {
			return fmt.Sprintf("%s, %s", p.DefaultOrder, p.StabilityField)
		} else {
			return "" // 无默认排序，直接返回空字符串
		}
	}

	// 分割多字段排序
	var validOrders []string
	for _, expr := range strings.Split(p.SortExpr, ",") {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			continue
		}

		// 解析单个表达式
		parts := strings.Fields(expr)
		if len(parts) > 2 {
			continue
		}

		// 提取字段和方向
		field := parts[0]
		direction := "ASC"
		if len(parts) > 1 {
			direction = strings.ToUpper(parts[1])
			if direction != "ASC" && direction != "DESC" {
				direction = "ASC"
			}
		}

		// 白名单验证
		if len(whitelist) != 0 && !slices.Contains(whitelist, field) {
			continue
		}

		validOrders = append(validOrders, fmt.Sprintf("%s %s", field, direction))
	}

	// 构造最终 SQL
	if len(validOrders) == 0 {
		return fmt.Sprintf("%s, %s", p.DefaultOrder, p.StabilityField)
	}

	// 自动追加稳定性字段
	return fmt.Sprintf(
		"%s, %s %s",
		strings.Join(validOrders, ", "),
		p.StabilityField,
		inferStabilityDirection(validOrders),
	)
}

// 推断稳定性字段排序方向
func inferStabilityDirection(orders []string) string {
	lastExpr := orders[len(orders)-1]
	if strings.HasSuffix(lastExpr, " DESC") {
		return "DESC"
	}
	return "ASC"
}
