package logics

import (
	"context"
	"errors"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"ontology-manager/common"
	"ontology-manager/interfaces"
)

func Init(ctx context.Context, appSetting *common.AppSetting) error {
	// 初始化 OpenSearch 索引
	logger.Info("InitKNConceptIndex Start")

	if appSetting.ServerSetting.DefaultSmallModelEnabled {
		// 初始化模型工厂
		smallModel, err := MFA.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel err:%v", err)
			return err
		}
		if smallModel == nil {
			logger.Errorf("GetDefaultModel return nil")
			return errors.New("GetDefaultModel return nil")
		}
		interfaces.KN_CONCEPT_INDEX_VECTOR_PROP["dimension"] = smallModel.EmbeddingDim
	}

	err := OSA.PutIndexTemplate(ctx, interfaces.KN_CONCEPT_INDEX_TEMP_NAME, interfaces.KN_CONCEPT_INDEX_TEMP)
	if err != nil {
		logger.Errorf("PutKNConceptIndexTemplate err:%v", err)
		return err
	}

	exists, err := OSA.IndexExists(ctx, interfaces.KN_CONCEPT_INDEX_NAME)
	if err != nil {
		logger.Errorf("CheckKNConceptIndexExists err:%v", err)
		return err
	}

	if !exists {
		err = OSA.CreateIndex(ctx, interfaces.KN_CONCEPT_INDEX_NAME, map[string]string{})
		if err != nil {
			logger.Errorf("CreateKNConceptIndex err:%v", err)
			return err
		}
	}

	logger.Info("InitKNConceptIndex Success")
	return nil
}
