package rds

import (
	"sync"
)

var (
	confDao     ConfDao
	confDaoOnce sync.Once

	aiModelDao     AiModelDao
	aiModelDaoOnce sync.Once

	alarmRuleDao     AlarmRuleDao
	alarmRuleDaoOnce sync.Once

	contentAdminDao     ContentAmdinDao
	contentAdminDaoOnce sync.Once

	agentDao     AgentDao
	agentDaoOnce sync.Once

	dagInstanceEventRepository     DagInstanceEventRepository
	dagInstanceEventRepositoryOnce sync.Once

	dagInstanceExtDataDao     DagInstanceExtDataDao
	dagInstanceExtDataDaoOnce sync.Once

	executorDao     ExecutorDao
	executorDaoOnce sync.Once

	taskCache     TaskCache
	taskCacheOnce sync.Once
)

func SetConfDao(dao ConfDao) {
	confDao = dao
}

func SetAiModelDao(dao AiModelDao) {
	aiModelDao = dao
}

func SetAlarmRuleDao(dao AlarmRuleDao) {
	alarmRuleDao = dao
}

func SetContentAdminDao(dao ContentAmdinDao) {
	contentAdminDao = dao
}

func SetAgentDao(dao AgentDao) {
	agentDao = dao
}

func SetDagInstanceEventRepository(repo DagInstanceEventRepository) {
	dagInstanceEventRepository = repo
}

func SetDagInstanceExtDataDao(dao DagInstanceExtDataDao) {
	dagInstanceExtDataDao = dao
}

func SetExecutorDao(dao ExecutorDao) {
	executorDao = dao
}

func SetTaskCache(cache TaskCache) {
	taskCache = cache
}

func GetConfDao() ConfDao {
	return confDao
}

func GetAiModelDao() AiModelDao {
	return aiModelDao
}

func GetAlarmRuleDao() AlarmRuleDao {
	return alarmRuleDao
}

func GetContentAdminDao() ContentAmdinDao {
	return contentAdminDao
}

func GetAgentDao() AgentDao {
	return agentDao
}

func GetDagInstanceEventRepository() DagInstanceEventRepository {
	return dagInstanceEventRepository
}

func GetDagInstanceExtDataDao() DagInstanceExtDataDao {
	return dagInstanceExtDataDao
}

func GetExecutorDao() ExecutorDao {
	return executorDao
}

func GetTaskCache() TaskCache {
	return taskCache
}
