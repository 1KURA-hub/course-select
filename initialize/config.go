package initialize

import (
	"fmt"
	"go-course/global"

	"github.com/spf13/viper"
)

// 初始化配置函数
func InitConfig() {

	viper.SetConfigName("config")

	viper.SetConfigType("yaml")

	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {

		panic(fmt.Errorf("配置读取失败: %s", err))
	}

	if err := viper.Unmarshal(&global.Settings); err != nil {
		panic(fmt.Errorf("配置解析失败: %s", err))
	}

}
