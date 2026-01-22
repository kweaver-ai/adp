package driveradapters

import (
	"github.com/gin-gonic/gin"
)

// setGinMode 设置 Gin 为测试模式并返回恢复函数
func setGinMode() func() {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(oldMode)
	}
}
