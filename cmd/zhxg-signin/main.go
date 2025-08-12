package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"zhxg-signin/internal/config"
	"zhxg-signin/internal/logger"
	"zhxg-signin/internal/scheduler"
	"zhxg-signin/internal/signin"
)

var (
	cfgFile string
	cfg     config.Config
)

var rootCmd = &cobra.Command{
	Use:   "zhxg-signin",
	Short: "一个用于智慧学工的自动签到工具",
	Long:  `一个功能强大的智慧学工（wisestu）自动签到工具，支持验证码自动识别和定时任务。`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 加载配置
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			fmt.Printf("加载配置失败: %v\n", err)
			os.Exit(1)
		}

		// 绑定命令行标志
		viper.BindPFlag("user.username", cmd.Flags().Lookup("username"))
		viper.BindPFlag("user.password", cmd.Flags().Lookup("password"))
		viper.BindPFlag("location.longitude", cmd.Flags().Lookup("lng"))
		viper.BindPFlag("location.latitude", cmd.Flags().Lookup("lat"))
		viper.BindPFlag("llm.api_key", cmd.Flags().Lookup("api-key"))

		// 重新加载配置以应用命令行标志
		viper.Unmarshal(&cfg)

		// 初始化日志
		logger.InitLogger(cfg.Logging)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "执行一次签到任务",
	Run: func(cmd *cobra.Command, args []string) {
		logger.GetLogger().Info("开始执行一次性签到任务")
		service := signin.NewService(cfg)
		if err := service.Run(); err != nil {
			logger.GetLogger().Error("签到任务失败", zap.Error(err))
			os.Exit(1)
		}
		logger.GetLogger().Info("签到任务执行完毕")
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "以守护进程模式运行，执行定时任务",
	Run: func(cmd *cobra.Command, args []string) {
		scheduler.StartScheduler(cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./configs", "配置文件路径")
	
	runCmd.Flags().StringP("username", "u", "", "登录用户名")
	runCmd.Flags().StringP("password", "p", "", "登录密码")
	runCmd.Flags().Float64("lng", 0, "经度")
	runCmd.Flags().Float64("lat", 0, "纬度")
	runCmd.Flags().StringP("api-key", "k", "", "LLM API Key")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(daemonCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}