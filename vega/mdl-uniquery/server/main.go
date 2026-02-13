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

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	_ "go.uber.org/automaxprocs"

	"uniquery/common"
	"uniquery/common/middleware"
	access "uniquery/drivenadapters"
	"uniquery/drivenadapters/permission"
	"uniquery/driveradapters"
	"uniquery/logics"
	"uniquery/logics/data_dict"
)

type mgrService struct {
	appSetting  *common.AppSetting
	restHandler driveradapters.RestHandler
	subHandler  driveradapters.SubHandler
}

func (server *mgrService) start() {
	logger.Info("Server Starting")

	// 创建gin.engine 并注册Public API
	engine := gin.New()
	engine.Use(middleware.AccessLog())
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

	// NOTE: 启动事件监听进程
	func() {
		if server.appSetting.ServerSetting.EventSubscribeEnabled {
			ess := driveradapters.NewSubscribeHandler(server.appSetting)
			ess.Listen()
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
	// 开启 pprof
	// go func() {
	// 	http.ListenAndServe("0.0.0.0:6060", nil)
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

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Set顺序按字母升序排序
	logics.SetDataConnectionAccess(access.NewDataConnectionAccess(appSetting))
	logics.SetDataDictAccess(access.NewDataDictAccess(appSetting))
	logics.SetDataDictService(data_dict.NewDataDictService(appSetting))
	logics.SetDataManagerAccess(access.NewDataManagerAccess(appSetting))
	logics.SetDataViewAccess(access.NewDataViewAccess(appSetting))
	logics.SetDataViewRowColumnRuleAccess(access.NewDataViewRowColumnRuleAccess(appSetting))
	logics.SetEventModelAccess(access.NewEventModelAccess(appSetting))
	logics.SetIndexBaseAccess(access.NewIndexBaseAccess(appSetting))
	logics.SetKafkaAccess(access.NewKafkaAccess(appSetting))
	logics.SetLogGroupAccess(access.NewLogGroupAccess(appSetting))
	logics.SetMetricModelAccess(access.NewMetricModelAccess(appSetting))
	logics.SetObjectiveModelAccess(access.NewObjectiveModelAccess(appSetting))
	logics.SetOpenSearchAccess(access.NewOpenSearchAccess(appSetting))
	logics.SetPermissionAccess(permission.NewPermissionAccess(appSetting))
	logics.SetSearchAccess(access.NewSearchAccess(appSetting))
	logics.SetStaticAccess(access.NewStaticAccess(appSetting))
	logics.SetTraceModelAccess(access.NewTraceModelAccess(appSetting))
	logics.SetVegaDataSourceAccess(access.NewVegaDataSourceAccess(appSetting))
	logics.SetVegaGatewayAccess(access.NewVegaGatewayAccess(appSetting))
	logics.SetVegaViewAccess(access.NewVegaAccess(appSetting))

	server := &mgrService{
		appSetting:  appSetting,
		restHandler: driveradapters.NewRestHandler(appSetting),
		subHandler:  driveradapters.NewSubscribeHandler(appSetting),
	}
	server.start()
}
