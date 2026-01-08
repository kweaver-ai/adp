package logics

import (
	"context"
	"ontology-manager/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

func Init(ctx context.Context) error {
	// 初始化 OpenSearch 索引
	logger.Info("InitKNConceptIndex Start")

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
