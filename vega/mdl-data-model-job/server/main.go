// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package main

import (
	"context"
	"net/http"

	// _ "net/http/pprof"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/gin-gonic/gin"

	"data-model-job/common"
	access "data-model-job/drivenadapters"
	"data-model-job/driveradapters"
	"data-model-job/interfaces"
	"data-model-job/logics"
)

type mgrService struct {
	appSetting  *common.AppSetting
	restHandler driveradapters.RestHandler
	jobService  interfaces.JobService
}

func (server *mgrService) start() {
	logger.Info("Server Starting")

	// 创建gin.engine 并注册Public API
	engine := gin.New()
	server.restHandler.RegisterPublic(engine)
	logger.Info("Server Register API Success")

	// 监听中断信号（SIGINT、SIGTERM）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// 在收到信号的时候，会自动触发 ctx 的 Done ，这个 stop 是不再捕获注册的信号的意思，算是一种释放资源。
	defer stop()

	// 初始化http服务
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(server.appSetting.ServerSetting.HttpPort),
		Handler:        engine,
		ReadTimeout:    server.appSetting.ServerSetting.ReadTimeOut * time.Second,
		WriteTimeout:   server.appSetting.ServerSetting.WriteTimeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动http服务
	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("s.ListenAndServe err:%v", err)
		}
	}()

	<-ctx.Done()
	// 重置系统中断信号处理
	// stop()

	// 设置系统最后处理时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 停止http服务
	logger.Info("Server Start Shutdown")
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown:%v", err)
	}
	logger.Info("Server Exiting")
}

func main() {
	// go func() {
	// 	http.ListenAndServe("0.0.0.0:9009", nil) //开启一个http服务，nil表示绑定默认路由器DefaultServeMux
	// }()

	logger.Info("Server Starting")

	// 初始化服务配置
	appSetting := common.NewSetting()

	logger.Info("Server Init Setting Success")

	// 设置错误码语言
	rest.SetLang(appSetting.ServerSetting.Language)
	logger.Info("Server Set Language Success")

	// 设置gin运行模式
	gin.SetMode(appSetting.ServerSetting.RunMode)
	logger.Info("Server Set RunMode Success")

	logger.Infof("Server Start By Port:%d", appSetting.ServerSetting.HttpPort)

	logics.SetJobAccess(access.NewJobAccess(appSetting))
	logics.SetKafkaAccess(access.NewKafkaAccess(appSetting))
	logics.SetIndexBaseAccess(access.NewIndexBaseAccess(appSetting))
	logics.SetMetricModelAccess(access.NewMetricModelAccess(appSetting))
	logics.SetUniqueryAccess(access.NewUniqueryAccess(appSetting))
	logics.SetEventModelAccess(access.NewEventModelAccess(appSetting))

	server := &mgrService{
		appSetting:  appSetting,
		restHandler: driveradapters.NewRestHandler(appSetting),
		jobService:  logics.NewJobService(appSetting),
	}
	server.start()
}
