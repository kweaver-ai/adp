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
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	_ "go.uber.org/automaxprocs"

	"data-model/common"
	"data-model/drivenadapters/data_connection"
	"data-model/drivenadapters/data_dict"
	"data-model/drivenadapters/data_model_job"
	"data-model/drivenadapters/data_view"
	"data-model/drivenadapters/event_model"
	"data-model/drivenadapters/index_base"
	"data-model/drivenadapters/metric_model"
	"data-model/drivenadapters/objective_model"
	"data-model/drivenadapters/permission"
	"data-model/drivenadapters/scan_record"
	"data-model/drivenadapters/trace_model"
	"data-model/drivenadapters/uniquery"
	"data-model/drivenadapters/vega"
	"data-model/driveradapters"
	"data-model/logics"
)

type mgrService struct {
	appSetting  *common.AppSetting
	restHandler driveradapters.RestHandler
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

	db := libdb.NewDB(&appSetting.DBSetting)
	logics.SetDB(db)

	audit.Init(&appSetting.MQSetting)

	// Set顺序按字母升序排序
	logics.SetDataConnectionAccess(data_connection.NewDataConnectionAccess(appSetting))
	logics.SetDataDictAccess(data_dict.NewDataDictAccess(appSetting))
	logics.SetDataDictItemAccess(data_dict.NewDictItemAccess(appSetting))
	logics.SetDataModelJobAccess(data_model_job.NewDataModelJobAccess(appSetting))
	logics.SetDataSourceAccess(vega.NewDataSourceAccess(appSetting))
	logics.SetDataViewAccess(data_view.NewDataViewAccess(appSetting))
	logics.SetDataViewGroupAccess(data_view.NewDataViewGroupAccess(appSetting))
	logics.SetDataViewRowColumnRuleAccess(data_view.NewDataViewRowColumnRuleAccess(appSetting))
	logics.SetEventModelAccess(event_model.NewEventModelAccess(appSetting))
	logics.SetIndexBaseAccess(index_base.NewIndexBaseAccess(appSetting))
	logics.SetMetricModelAccess(metric_model.NewMetricModelAccess(appSetting))
	logics.SetMetricModelGroupAccess(metric_model.NewMetricModelGroupAccess(appSetting))
	logics.SetMetricModelTaskAccess(metric_model.NewMetricModelTaskAccess(appSetting))
	logics.SetObjectiveModelAccess(objective_model.NewObjectiveModelAccess(appSetting))
	logics.SetPermissionAccess(permission.NewPermissionAccess(appSetting))
	logics.SetScanRecordAccess(scan_record.NewScanRecordAccess(appSetting))
	logics.SetTraceModelAccess(trace_model.NewTraceModelAccess(appSetting))
	logics.SetUniqueryAccess(uniquery.NewUniqueryAccess(appSetting))
	logics.SetVegaGatewayAccess(vega.NewVegaGatewayAccess(appSetting))
	logics.SetVegaMetadataAccess(vega.NewVegaMetadataAccess(appSetting))

	server := &mgrService{
		appSetting:  appSetting,
		restHandler: driveradapters.NewRestHandler(appSetting),
	}
	server.start()
}
