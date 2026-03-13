// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/knowledge_network"
)

var (
	bServiceOnce sync.Once
	bService     interfaces.BKNService
)

type bknService struct {
	appSetting *common.AppSetting
	kns        interfaces.KNService
}

// NewBKNService 创建 BKN 服务
func NewBKNService(appSetting *common.AppSetting) interfaces.BKNService {
	bServiceOnce.Do(func() {
		bService = &bknService{
			appSetting: appSetting,
			kns:        knowledge_network.NewKNService(appSetting),
		}
	})
	return bService
}

// ImportFromTar 从 tar 包导入 BKN 定义（纯内存处理）
func (bs *bknService) Import(ctx context.Context, bknNetwork *bknsdk.BknNetwork) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "BKN从Tar导入", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logger.Debugf("BKN Import Start: kn_id=%s, branch=%s", bknNetwork.ID, bknNetwork.Branch)

	// 校验 BKN 文件
	if err := bs.validateNetwork(bknNetwork); err != nil {
		logger.Errorf("Validation failed: %s", err.Error())
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// 执行导入
	kn := logics.ToADPNetWork(bknNetwork)
	otMap := make(map[string]*interfaces.ObjectType)
	for _, bknObj := range bknNetwork.ObjectTypes {
		ot := logics.ToADPObjectType(kn.KNID, kn.Branch, bknObj)
		kn.ObjectTypes = append(kn.ObjectTypes, ot)
		otMap[ot.OTID] = ot
	}
	for _, bknRel := range bknNetwork.RelationTypes {
		rt := logics.ToADPRelationType(kn.KNID, kn.Branch, bknRel)
		kn.RelationTypes = append(kn.RelationTypes, rt)
	}
	for _, bknAct := range bknNetwork.ActionTypes {
		act := logics.ToADPActionType(kn.KNID, kn.Branch, bknAct)
		kn.ActionTypes = append(kn.ActionTypes, act)
	}
	for _, bknCG := range bknNetwork.ConceptGroups {
		cg := logics.ToADPConceptGroup(kn.KNID, kn.Branch, bknCG)
		kn.ConceptGroups = append(kn.ConceptGroups, cg)

		for _, otID := range bknCG.ObjectTypes {
			if ot, ok := otMap[otID]; ok {
				ot.ConceptGroups = append(ot.ConceptGroups, cg)
			}
		}
	}

	// 调用创建单个知识网络
	knID, err := bs.kns.CreateKN(ctx, kn, interfaces.ImportMode_Overwrite, false)
	if err != nil {
		logger.Errorf("Failed to create KN for %s (%s %s): %v", kn.KNName, kn.KNID, kn.Branch, err)
		return "", err
	}

	logger.Debugf("BKN ImportFromTar Completed: kn_id=%s", knID)
	return knID, nil
}

// ExportToTar 将知识网络导出为 tar 包
func (bs *bknService) ExportToTar(ctx context.Context, knID string, branch string) ([]byte, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "BKN导出为Tar", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	logger.Debugf("BKN ExportToTar Start: kn_id=%s", knID)

	kn, err := bs.kns.GetKNByID(ctx, knID, branch, interfaces.Mode_Export)
	if err != nil {
		logger.Errorf("BKN GetKNByID failed: %s", err.Error())
		return nil, err
	}

	bknNetwork := logics.ToBKNNetWork(kn)
	for _, ot := range kn.ObjectTypes {
		bknNetwork.ObjectTypes = append(bknNetwork.ObjectTypes, logics.ToBKNObjectType(ot))
	}
	for _, rt := range kn.RelationTypes {
		bknNetwork.RelationTypes = append(bknNetwork.RelationTypes, logics.ToBKNRelationType(rt))
	}
	for _, act := range kn.ActionTypes {
		bknNetwork.ActionTypes = append(bknNetwork.ActionTypes, logics.ToBKNActionType(act))
	}
	for _, cg := range kn.ConceptGroups {
		bknNetwork.ConceptGroups = append(bknNetwork.ConceptGroups, logics.ToBKNConceptGroup(cg))
	}

	var buf bytes.Buffer
	err = bknsdk.WriteNetworkToTar(bknNetwork, &buf)
	if err != nil {
		logger.Errorf("BKN ExportToTar failed: %s", err.Error())
		return nil, err
	}
	tarData := buf.Bytes()

	logger.Debugf("BKN ExportToTar Completed: size=%d", len(tarData))
	return tarData, nil
}

// validateNetwork 校验网络结构
func (bs *bknService) validateNetwork(network *bknsdk.BknNetwork) error {
	if network.Type != "network" {
		return fmt.Errorf("root file must be type: network, got: %s", network.Type)
	}

	return nil
}
