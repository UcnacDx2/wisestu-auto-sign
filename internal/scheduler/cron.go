package scheduler

import (
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"zhxg-signin/internal/config"
	"zhxg-signin/internal/logger"
	"zhxg-signin/internal/signin"
)

// StartScheduler 启动定时签到任务
func StartScheduler(cfg config.Config) {
	log := logger.GetLogger()
	if !cfg.Scheduler.Enabled {
		log.Info("定时任务未启用")
		return
	}

	loc, err := time.LoadLocation(cfg.Scheduler.Timezone)
	if err != nil {
		log.Error("加载时区失败", zap.Error(err))
		return
	}

	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc(cfg.Scheduler.Cron, func() {
		log.Info("开始执行定时签到任务")
		service := signin.NewService(cfg)
		if err := service.Run(); err != nil {
			log.Error("定时签到任务失败", zap.Error(err))
		} else {
			log.Info("定时签到任务成功")
		}
	})

	if err != nil {
		log.Error("添加 cron 任务失败", zap.Error(err))
		return
	}

	log.Info("定时任务已启动", zap.String("cron", cfg.Scheduler.Cron))
	c.Start()

	// 阻塞主 goroutine
	select {}
}