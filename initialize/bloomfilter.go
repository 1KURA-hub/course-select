package initialize

import (
	"fmt"
	"go-course/global"
	"go-course/model"

	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
)

// 初始化布隆过滤器
func InitBloomFilter() {
	// 库会自动在底层计算出需要多大的 bit 数组和几个 Hash 函数
	global.CourseBloomFilter = bloom.NewWithEstimates(10000, 0.01)

	// 从数据库全量查出所有课程 ID (Pluck只查一个字段)
	var courseIDs []uint
	err := global.DB.Model(&model.Course{}).Pluck("id", &courseIDs).Error
	if err != nil {
		global.Logger.Fatal("预热布隆过滤器失败", zap.Error(err))
	}

	// 把所有的真实 ID 写入布隆过滤器
	for _, id := range courseIDs {
		global.CourseBloomFilter.AddString(fmt.Sprintf("%d", id))
	}

	global.Logger.Info("布隆过滤器预热完成", zap.Int("ID数量", len(courseIDs)))
}
