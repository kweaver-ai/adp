package com.eisoo.dc.metadata.service.impl;

import cn.hutool.json.JSONArray;
import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.JSONObject;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.eisoo.dc.common.config.MetaDataConfig;
import com.eisoo.dc.common.config.ScanTaskPoolConfig;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.connector.ConnectorConfigCache;
import com.eisoo.dc.common.constant.Description;
import com.eisoo.dc.common.constant.Detail;
import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.constant.ResourceAuthConstant;
import com.eisoo.dc.common.driven.Authorization;
import com.eisoo.dc.common.driven.UserManagement;
import com.eisoo.dc.common.driven.service.ServiceEndpoints;
import com.eisoo.dc.common.enums.ConnectorEnums;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.metadata.mapper.TaskScanMapper;
import com.eisoo.dc.common.metadata.mapper.TaskScanTableMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.LockUtil;
import com.eisoo.dc.common.util.StringUtils;
import com.eisoo.dc.common.util.http.MetadataHttpUtils;
import com.eisoo.dc.common.util.http.OpensearchHttpUtils;
import com.eisoo.dc.common.util.jdbc.db.DbConnectionStrategyFactory;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.eisoo.dc.common.vo.ResourceAuthVo;
import com.eisoo.dc.metadata.domain.dto.*;
import com.eisoo.dc.metadata.domain.vo.QueryStatementVO;
import com.eisoo.dc.metadata.domain.vo.TableRetryVO;
import com.eisoo.dc.metadata.domain.vo.TableStatusVO;
import com.eisoo.dc.metadata.domain.vo.TaskScanVO;
import com.eisoo.dc.metadata.service.IFieldScanService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanService;
import com.eisoo.dc.metadata.service.ITaskScanTableService;
import com.google.common.util.concurrent.Futures;
import com.google.common.util.concurrent.ListenableFuture;
import com.google.common.util.concurrent.ListeningExecutorService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpServletRequest;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class TaskScanServiceImpl extends ServiceImpl<TaskScanMapper, TaskScanEntity> implements ITaskScanService {
    @Autowired(required = false)
    private TaskScanMapper taskScanMapper;
    @Autowired(required = false)
    private TaskScanTableMapper taskScanTableMapper;
    @Autowired
    private ITableScanService tableScanService;
    @Autowired
    private IFieldScanService fieldScanService;
    @Autowired
    private ITaskScanTableService taskScanTableService;
    @Autowired
    @Qualifier("tableScanExecutor") // 按名称注入
    private ListeningExecutorService tableScanExecutor;
    @Autowired
    @Qualifier("dsScanExecutor") // 按名称注入
    private ListeningExecutorService dsScanExecutor;
    @Autowired
    @Qualifier("scanTaskExecutor") // 按名称注入
    private ListeningExecutorService scanTaskExecutor;
    @Autowired
    private MetaDataFetchServiceBase metaDataFetchServiceBase;
    @Autowired(required = false)
    private DataSourceMapper dataSourceMapper;
    @Autowired(required = false)
    private MetaDataConfig metaDataConfig;
    @Autowired(required = false)
    private ScanTaskPoolConfig scanTaskPoolConfig;
    @Autowired
    private ServiceEndpoints serviceEndpoints;
    @Autowired
    ConnectorConfigCache connectorConfigCache;


    @Override
    public ResponseEntity<?> createScanTaskAndStart(HttpServletRequest request, TaskScanVO taskScanVO) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        // 校验参数
        validateParam(taskScanVO, introspectInfo);
        TaskCreateDto taskCreateDto = buildScanTask(
                userId,
                taskScanVO,
                token);
        return ResponseEntity.ok(taskCreateDto);
    }

    @Override
    public ResponseEntity<?> getScanTaskInfo(HttpServletRequest request, String taskId) {
        TaskScanEntity taskScanEntity = taskScanMapper.selectById(taskId);
        if (null == taskScanEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.SCAN_NOT_FOUND_ERROR,
                    String.format(Detail.META_DATA_SCAN_NOT_FOUND, taskId),
                    Message.MESSAGE_SERVICE_ERROR);
        }
        TaskStatusInfoDto.TaskData taskData = new TaskStatusInfoDto.TaskData();
        taskData.setTaskResultInfo(JSONObject.parseObject(taskScanEntity.getTaskResultInfo(),
                TaskStatusInfoDto.TaskResultInfo.class)
        );
        taskData.setTaskProcessInfo(JSONObject.parseObject(taskScanEntity.getTaskProcessInfo(),
                TaskStatusInfoDto.TaskProcessInfo.class)
        );

        TaskStatusInfoDto taskStatusInfoDto = new TaskStatusInfoDto(
                taskId,
                ScanStatusEnum.fromCode(taskScanEntity.getScanStatus()),
                taskData
        );
        return ResponseEntity.ok(taskStatusInfoDto);
    }

    @Override
    public ResponseEntity<?> getScanTaskStatus(HttpServletRequest request, TableStatusVO req) {
        String taskId = req.getId();
        List<String> tables = req.getTables();
        TableStatusDto tableStatusDto = new TableStatusDto();
        tableStatusDto.setId(taskId);
        List<TableStatusDto.TableStatus> list = new ArrayList<>(tables.size());
        List<TaskScanTableEntity> taskScanTableEntities = taskScanTableMapper.selectByTaskId(taskId);
        for (TaskScanTableEntity table : taskScanTableEntities) {
            String tableId = table.getTableId();
            if (tables.contains(tableId)) {
                TableStatusDto.TableStatus tableStatus = new TableStatusDto.TableStatus(
                        tableId,
                        table.getTableName(),
                        ScanStatusEnum.fromCode(table.getScanStatus()),
                        table.getStartTime()
                );
                list.add(tableStatus);
            }
        }
        tableStatusDto.setTables(list);
        return ResponseEntity.ok(tableStatusDto);
    }

    @Override
    public ResponseEntity<?> retryScanTable(HttpServletRequest request, TableRetryVO req) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        String taskId = req.getId();
        List<String> tables = req.getTables();

        TableStatusDto tableStatusDto = new TableStatusDto();
        tableStatusDto.setId(taskId);

        List<TaskScanTableEntity> listPre = taskScanTableMapper.selectByTaskIdAndIds(taskId, tables);
        submitTablesScanTaskInner(listPre, 1, userId, true);

        List<TaskScanTableEntity> listPost = taskScanTableMapper.selectByTaskIdAndIds(taskId, tables);
        // 更新结果信息
        updateScanTaskResultInfo(taskId, listPost);
        // 构造返回
        List<TableStatusDto.TableStatus> list = new ArrayList<>(tables.size());

        for (TaskScanTableEntity table : listPost) {
            String tableId = table.getTableId();
            TableStatusDto.TableStatus tableStatus = new TableStatusDto.TableStatus();
            tableStatus.setTableId(tableId);
            tableStatus.setStatus(ScanStatusEnum.fromCode(table.getScanStatus()));
            list.add(tableStatus);
        }
        tableStatusDto.setTables(list);
        return ResponseEntity.ok(tableStatusDto);
    }

    @Override
    public ResponseEntity<?> getScanTaskTableStatus(String userId, String taskId, String status, String keyword, int limit, int offset, String sort, String direction) {
        Set<Integer> statusList = new HashSet<>();
        if (CommonUtil.isEmpty(status)) {
            statusList.add(ScanStatusEnum.WAIT.getCode());
            statusList.add(ScanStatusEnum.RUNNING.getCode());
            statusList.add(ScanStatusEnum.SUCCESS.getCode());
            statusList.add(ScanStatusEnum.FAIL.getCode());
        } else {
            statusList.add(ScanStatusEnum.fromDesc(status));
        }
        TaskScanEntity taskScanEntity = taskScanMapper.selectById(taskId);
        if (null == taskScanEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.TASK_NOT_FOUND_ERROR,
                    String.format(Detail.TASK_NOT_EXIST, taskId),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        JSONObject response = new JSONObject();
        List<TaskScanTableEntity> dsList = taskScanTableMapper.selectTaskScanTables(taskId, statusList, keyword);
        long count = taskScanTableMapper.selectCountTaskScanTables(null, taskId, statusList, keyword);
        if (dsList.size() == 0) {
            response.put("entries", new JSONArray());
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        Set<String> ids = dsList.stream().map(TaskScanTableEntity::getId).collect(Collectors.toSet());
        List<TaskScanTableEntity> entities = taskScanTableMapper.selectPageTaskScanTables(ids,
                keyword,
                offset,
                limit,
                sort,
                direction
        );
        List<TaskScanTableDto> results = entities.stream().map(t -> new TaskScanTableDto(t.getTaskId(),
                t.getTableId(),
                t.getTableName(),
                ScanStatusEnum.fromCode(t.getScanStatus()),
                t.getStartTime()
        )).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> getScanTaskList(String userId,
                                             String dsId,
                                             String status,
                                             String keyword,
                                             int limit,
                                             int offset,
                                             String sort,
                                             String direction) {
        //不是空的时候验证数据源ID
        if (CommonUtil.isNotEmpty(dsId)) {
            DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
            if (null == dataSourceEntity) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        Description.DS_NOT_FOUND_ERROR,
                        String.format(Detail.DS_NOT_EXIST, dsId),
                        Message.MESSAGE_INTERNAL_ERROR);
            }
        }
        List<Integer> statusList = new ArrayList<>(4);
        if (CommonUtil.isEmpty(status)) {
            statusList.add(ScanStatusEnum.WAIT.getCode());
            statusList.add(ScanStatusEnum.RUNNING.getCode());
            statusList.add(ScanStatusEnum.SUCCESS.getCode());
            statusList.add(ScanStatusEnum.FAIL.getCode());
        } else {
            statusList.add(ScanStatusEnum.fromDesc(status));
        }
        JSONArray entries = new JSONArray();
        JSONObject response = new JSONObject();
        List<TaskScanEntity> dsList = taskScanMapper.selectTaskScans(null, dsId, statusList, keyword);
        long count = taskScanMapper.selectCount(null, dsId, statusList, keyword);

        if (dsList.size() == 0) {
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        //TODO:这里可以根据userId过滤资源
        Set<String> ids = dsList.stream().map(TaskScanEntity::getId).collect(Collectors.toSet());
        List<TaskScanEntity> taskScanEntities = taskScanMapper.selectPage(ids, keyword, offset, limit, sort, direction);
        // 获取用户名称
        Set<String> userIds = new HashSet<>();
        taskScanEntities.forEach(t -> userIds.add(t.getCreateUser()));
        Map<String, String[]> userInfosMap = UserManagement.batchGetUserInfosByUserIds(serviceEndpoints.getUserManagementPrivate(), userIds);

        List<TaskScanDto> results = taskScanEntities.stream().map(t -> {
            String userName = userInfosMap.getOrDefault(t.getCreateUser(), new String[]{"", ""})[1];
            DataSourceEntity ds = dataSourceMapper.selectById(t.getDsId());
            return new TaskScanDto(
                    t.getId(),
                    t.getType(),
                    t.getName(),
                    ds.getFType(),
                    ds.getFType().equals(ConnectorEnums.OPENSEARCH.getConnector()),
                    userName,
                    ScanStatusEnum.fromCode(t.getScanStatus()),
                    t.getStartTime(),
                    t.getTaskProcessInfo(),
                    t.getTaskResultInfo()
            );
        }).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);

    }

    @Override
    public ResponseEntity<?> queryDslStatement(HttpServletRequest request, QueryStatementVO req) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        String dsId = req.getDsId();
        String index = req.getIndex();
        String statement = req.getStatement();
        String response = null;
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        if (null == dataSourceEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_NOT_FOUND_ERROR,
                    String.format(Detail.DS_NOT_EXIST, dsId),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        String fType = dataSourceEntity.getFType();
        if (ConnectorEnums.OPENSEARCH.getConnector().equals(fType)) {
            try {
                response = OpensearchHttpUtils.queryStatement(dataSourceEntity, index, statement);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        } else {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_QUERY_SUPPORTED_ERROR,
                    String.format(Detail.META_DATA_QUERY_UNSUPPORTED, fType),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        System.out.println(response);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> createScanTaskAndStartBatch(HttpServletRequest request, List<TaskScanVO> reqs) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        // 校验参数
        for (TaskScanVO taskScanVO : reqs) {
            validateParam(taskScanVO, introspectInfo);
        }
        List<TaskCreateDto> taskCreateDtos = new ArrayList<>(reqs.size());
        for (TaskScanVO taskScanVO : reqs) {
            TaskCreateDto taskCreateDto = buildScanTask(
                    userId,
                    taskScanVO,
                    token);
            taskCreateDtos.add(taskCreateDto);
        }
        return ResponseEntity.ok(taskCreateDtos);
    }

    @Override
    public void submitDsScanTask(String taskId, String userId) {
        // 1,两阶段提交
        // 1.1,更新t_task_scan:running
        TaskScanEntity taskScanEntityStart = taskScanMapper.selectById(taskId);
        Integer type = taskScanEntityStart.getType();
        taskScanMapper.updateScanStatusStart(taskId, ScanStatusEnum.RUNNING.getCode());
        String dsId = taskScanEntityStart.getDsId();
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        String dsType = dataSourceEntity.getFType();
        log.info("【元数据扫描】dsId:{};dsType:{};taskId:{};taskType:{}:更新task扫描状态:{}-成功;开始进行【stage-1】:获取table元数据......",
                dsId,
                dsType,
                taskId,
                type,
                ScanStatusEnum.RUNNING);
        // 1.2,ds扫描，首先获取所有的table
        if (0 == type) {
            try {
                metaDataFetchServiceBase.getTables(taskScanEntityStart, userId);
            } catch (Exception e) {
                log.error("【元数据扫描:stage-1】dsId:{};dsType:{};taskId:{}  获取table失败,退出!",
                        dsId,
                        dsType,
                        taskId,
                        e);
                throw new RuntimeException(e);
            }
        }
        log.info("【元数据扫描:stage-1】dsId:{};dsType:{};taskId:{};taskType:{}: 获取tables结束，开始进行【stage-2】-获取field元数据",
                dsId,
                dsType,
                taskId,
                type);
        //2,提交field
        submitTablesScanTask(taskId, userId);
    }

    @Override
    public void submitTablesScanTask(String taskId, String userId) {
        // 0,更新t_task_scan
        TaskScanEntity taskScanEntityStart = taskScanMapper.selectById(taskId);
        // 1,查出来table信息
        List<TaskScanTableEntity> taskScanTableEntities = taskScanTableMapper.selectByTaskId(taskId);
        submitTablesScanTaskInner(taskScanTableEntities, taskScanEntityStart.getType(), userId, false);
        // 更新结果状态
        updateScanTaskResultInfo(taskId, taskScanTableEntities);
    }

    @Override
    @Transactional(rollbackFor = Exception.class, propagation = Propagation.REQUIRES_NEW)
    public void updateByIdNewRequires(TaskScanEntity taskScanEntity) {
        updateById(taskScanEntity);
    }

    private void submitTablesScanTaskInner(List<TaskScanTableEntity> taskScanTableEntities,
                                           Integer type,
                                           String userId,
                                           boolean isRetry) {
        List<TaskScanTableEntity> waitingTableEntities = new ArrayList<>(taskScanTableEntities.size());
        List<ListenableFuture<String>> futures = new ArrayList<>();
        List<String> lockIds = new ArrayList<>();
        int totalCount = taskScanTableEntities.size();
        String taskId = null;
        for (TaskScanTableEntity tableEntity : taskScanTableEntities) {
            String tableId = tableEntity.getTableId();
            taskId = tableEntity.getTaskId();
            String tableName = tableEntity.getTableName();
            String dsId = tableEntity.getDsId();
            DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
            ConnectorConfig connectorConfig = connectorConfigCache.getConnectorConfig(dataSourceEntity.getFType());
            boolean getLock = LockUtil.GLOBAL_MULTI_TASK_LOCK.tryLock(tableId, 2, TimeUnit.SECONDS, true);
            if (getLock) {
                // 这个后面要释放
                log.info("【元数据扫描:stage-2】采集任务获取锁成功，准备提交到pool执行:taskId:{};tableId:{};tableName:{}",
                        taskId,
                        tableId,
                        tableName);
                lockIds.add(tableId);
                FieldFetchCallable fieldFetchCallable = new FieldFetchCallable(
                        tableEntity,
                        tableScanService,
                        fieldScanService,
                        taskScanTableService,
                        dataSourceEntity,
                        userId,
                        connectorConfig
                );
                if (type == 1) {
                    futures.add(tableScanExecutor.submit(fieldFetchCallable));
                } else if (type == 0) {
                    futures.add(dsScanExecutor.submit(fieldFetchCallable));
                }
                log.info("【元数据扫描:stage-2】采集任务获取锁成功，已经提交到pool执行:taskId:{};tableId:{};tableName:{}",
                        taskId,
                        tableId,
                        tableName);
            } else {
                log.warn("【元数据扫描:stage-2】采集任务获取锁失败，存放到waiting列表下次执行:taskId:{};tableId:{};tableName:{}",
                        taskId,
                        tableId,
                        tableName);
                waitingTableEntities.add(tableEntity);
            }
            if (lockIds.size() >= scanTaskPoolConfig.getNumThread()) {
                // 到达数量，提交
                ListenableFuture<List<String>> listListenableFuture = Futures.allAsList(futures);
                List<String> waitMainThread = null;
                try {
                    waitMainThread = listListenableFuture.get();
                } catch (Exception e) {
                    log.error("【元数据扫描:stage-2】采集任务失败", e);
                } finally {
                    for (String id : lockIds) {
                        // 确保释放锁（只释放当前线程持有的锁）
                        if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                            LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
                            log.info("【元数据扫描:stage-2】采集任务结束:当前线程:{} 成功释放table:{} 的锁",
                                    Thread.currentThread().getName(),
                                    id);
                        }
                    }
                    // 更新中间任务状态信息
                    if (!isRetry) {
                        updateScanTaskProcessInfo(taskId, totalCount, waitMainThread);
                    }
                    lockIds.clear();
                    futures.clear();
                }
            }
        }
        if (lockIds.size() != 0) {
            ListenableFuture<List<String>> listListenableFuture = Futures.allAsList(futures);
            List<String> waitMainThread = null;
            try {
                waitMainThread = listListenableFuture.get();
            } catch (Exception e) {
                log.error("【元数据扫描:stage-2】采集任务失败", e);
            } finally {
                for (String id : lockIds) {
                    if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                        LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
                        log.info("【元数据扫描:stage-2】采集任务结束:当前线程:{} 成功释放table:{} 的锁",
                                Thread.currentThread().getName(),
                                id);
                    }
                }
            }
            // 更新状态信息
            if (!isRetry) {
                updateScanTaskProcessInfo(taskId, totalCount, waitMainThread);
            }
        }
        lockIds.clear();
        futures.clear();
        if (waitingTableEntities.size() != 0) {
            submitTablesScanTaskInner(waitingTableEntities, type, userId, isRetry);
        }
    }

    private void updateScanTaskProcessInfo(String taskId, int totalCount, List<String> waitMainThread) {
        int fail = 0;
        int success = 0;
        if (waitMainThread != null) {
            for (String flag : waitMainThread) {
                if (ScanStatusEnum.fromCode(ScanStatusEnum.FAIL.getCode()).equals(flag)) {
                    ++fail;
                } else if (ScanStatusEnum.fromCode(ScanStatusEnum.SUCCESS.getCode()).equals(flag)) {
                    ++success;
                }
            }
            TaskScanEntity taskScanEntity = taskScanMapper.selectById(taskId);
            String taskProcessInfo = taskScanEntity.getTaskProcessInfo();
            TaskStatusInfoDto.TaskProcessInfo taskProcess;
            if (CommonUtil.isEmpty(taskProcessInfo)) {
                taskProcess = new TaskStatusInfoDto.TaskProcessInfo();
                taskProcess.setTableCount(totalCount);
                taskProcess.setSuccessCount(success);
                taskProcess.setFailCount(fail);
            } else {
                taskProcess = JSONObject.parseObject(taskProcessInfo, TaskStatusInfoDto.TaskProcessInfo.class);
                Integer successOld = taskProcess.getSuccessCount();
                Integer failOld = taskProcess.getFailCount();
                taskProcess.setFailCount(fail + failOld);
                taskProcess.setSuccessCount(success + successOld);
            }
            String jsonString = JSONObject.toJSONString(taskProcess, WriteMapNullValue);
            taskScanEntity.setTaskProcessInfo(jsonString);
            taskScanMapper.updateById(taskScanEntity);
            log.info("【元数据扫描】taskId:[{}];taskName:[{}];更新了中间状态:{}",
                    taskScanEntity.getId(),
                    taskScanEntity.getName(),
                    jsonString);
        }
    }

    private void updateScanTaskResultInfo(String taskId, List<TaskScanTableEntity> taskScanTableEntities) {
        log.info("【元数据扫描】taskId:{}-采集任务成功，开始更新结果信息......", taskId);
        // 2,更新结果信息
        int tablesCount = taskScanTableEntities.size();
        int successCount = 0;
        int failCount = 0;
        // 3,查出来所有的扫描table
        List<String> ids = taskScanTableEntities.stream().map(TaskScanTableEntity::getId).collect(Collectors.toList());
        List<TaskScanTableEntity> tables = taskScanTableMapper.selectBatchIds(ids);
        for (TaskScanTableEntity table : tables) {
            Integer scanStatus = table.getScanStatus();
            if (scanStatus == ScanStatusEnum.SUCCESS.getCode()) {
                successCount++;
            } else if (scanStatus == ScanStatusEnum.FAIL.getCode()) {
                failCount++;
            }
        }
        TaskStatusInfoDto.TaskResultInfo taskResultInfo = new TaskStatusInfoDto.TaskResultInfo();
        taskResultInfo.setTableCount(tablesCount);
        taskResultInfo.setSuccessCount(successCount);
        taskResultInfo.setFailCount(failCount);
        TaskScanEntity taskScanEntity = taskScanMapper.selectById(taskId);
        int scanStatus = ScanStatusEnum.SUCCESS.getCode();
        if (failCount != 0) {
            scanStatus = ScanStatusEnum.FAIL.getCode();
        }
        taskScanEntity.setScanStatus(scanStatus);
        taskScanEntity.setEndTime(new Date());
        taskScanEntity.setTaskResultInfo(JSONObject.toJSONString(taskResultInfo, WriteMapNullValue));
        taskScanMapper.updateById(taskScanEntity);
        log.info("【元数据扫描】taskId:{}-采集任务成功!Task信息如下：\n{}", taskId, taskScanEntity);
    }

    private void validateParam(TaskScanVO taskScanVO, IntrospectInfo introspectInfo) {
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        TaskScanVO.DsInfo dsInfo = taskScanVO.getDsInfo();
        // ds_info参数校验
        if (CommonUtil.isEmpty(dsInfo)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "ds_info是空",
                    "ds_info不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        Integer type = taskScanVO.getType();
        Set<String> allConnectors = ConnectorEnums.getAllConnectors();
        String dsId = dsInfo.getDsId();
        String dsType = dsInfo.getDsType();
        // ds_info参数校验
        if (CommonUtil.isEmpty(dsId)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "ds_id是空",
                    "ds_id不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        if (null == dataSourceEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_NOT_FOUND_ERROR,
                    String.format(Detail.DS_NOT_EXIST, dsId),
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (CommonUtil.isEmpty(dsType)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "ds_type是空",
                    "ds_type不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (!dsType.equals(dataSourceEntity.getFType())) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "ds_info数据错误",
                    "ds_id与ds_type不匹配",
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (!allConnectors.contains(dsType)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_QUERY_SUPPORTED_ERROR,
                    String.format(Detail.DS_NOT_UNSUPPORTED, type),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        // 任务类型参数校验
        if (CommonUtil.isEmpty(type)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "type是空",
                    "type不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (0 != type && 1 != type) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_SCAN_SUPPORTED_ERROR,
                    String.format(Detail.META_DATA_SCAN__TASK_UNSUPPORTED, type),
                    Message.MESSAGE_INTERNAL_ERROR);
        }


        // 判断是否有扫描数据源的权限
        if (StringUtils.isBlank(userId)) {
            throw new AiShuException(ErrorCodeEnum.UnauthorizedError);
        }
        boolean isOk = Authorization.checkResourceOperation(
                serviceEndpoints.getAuthorizationPrivate(),
                userId,
                introspectInfo.getAccountType(),
                new ResourceAuthVo(dsId, ResourceAuthConstant.RESOURCE_TYPE_DATA_SOURCE),
                ResourceAuthConstant.RESOURCE_OPERATION_TYPE_SCAN);
        if (!isOk) {
            throw new AiShuException(ErrorCodeEnum.ForbiddenError,
                    String.format(Detail.RESOURCE_PERMISSION_ERROR, ResourceAuthConstant.RESOURCE_OPERATION_TYPE_SCAN));
        }
        // 检查数据源是否正在扫描:当type=0时
        int runningDs = taskScanMapper.getRunningDs(dsId);
        if (runningDs != 0) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_SCAN_RUNNING_ERROR,
                    String.format(Detail.DATASOURCE_SCAN_IS_RUNNING, dsId),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        // 非 OPEN_SEARCH只支持数据源类型扫描任务
        if (!DbConnectionStrategyFactory.supportNewScan(dsType) && type != 0) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_SCAN_SUPPORTED_ERROR,
                    String.format(Detail.META_DATA_SCAN_UNSUPPORTED, dsType),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        // tables不能没有数据，当type=1的时候
        List<String> tables = taskScanVO.getTables();
        if (type == 1 && (tables == null || tables.isEmpty())) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.TABLES_SCAN_EMPTY_ERROR,
                    Detail.TABLES_NOT_EMPLOY,
                    Message.MESSAGE_INTERNAL_ERROR);
        }
    }

    private TaskCreateDto buildScanTask(String userId, TaskScanVO taskScanVO, String token) {
        TaskScanVO.DsInfo dsInfo = taskScanVO.getDsInfo();
        String dsId = dsInfo.getDsId();
        String dsType = dsInfo.getDsType();
        Integer type = taskScanVO.getType();
        List<String> tables = taskScanVO.getTables();
        Date now = new Date();
        String taskId = UUID.randomUUID().toString();
        TaskScanEntity taskScanEntity = new TaskScanEntity(
                taskId,
                type,
                taskScanVO.getScanName(),
                dsId,
                ScanStatusEnum.WAIT.getCode(),
                now,
                null,
                userId,
                null,
                null,
                null
        );
        switch (type) {
            case 0:
                TaskScanEntity.TaskParamsInfo taskParamsInfo = new TaskScanEntity.TaskParamsInfo(
                        taskScanVO.getCronExpression(),
                        null,
                        null);
                taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParamsInfo, WriteMapNullValue));
                break;
            case 1:
                TaskScanEntity.TaskParamsInfo taskParams = new TaskScanEntity.TaskParamsInfo(
                        taskScanVO.getCronExpression(),
                        tables.size(),
                        tables);
                taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParams, WriteMapNullValue));
                // 将扫描的table插入到表里
                taskScanTableService.insertBatch(tables, dsId, taskId, userId);
                break;
            case 2:
                //TODO:暂不实现
                break;
            case 3:
                //TODO:暂不实现
                break;
        }
        // 持久化扫描任务
        taskScanMapper.insert(taskScanEntity);
        // 创建扫描任务
        TaskCreateDto taskCreateDto = new TaskCreateDto(
                taskId,
                dsId,
                ScanStatusEnum.WAIT.getDesc()
        );
        log.info("创建扫描任务成功:{}", taskScanEntity);
        // 目前支持的：maria mysql opensearch
        if (!DbConnectionStrategyFactory.supportNewScan(dsType)) {
            try {
                MetadataHttpUtils.sendMetaDataScan(metaDataConfig, taskId, token);
            } catch (Exception e) {
                taskScanMapper.updateScanStatusEnd(taskScanEntity.getId(), ScanStatusEnum.FAIL.getCode());
                throw new AiShuException(ErrorCodeEnum.CalculateError, e.getMessage());
            }
        } else {
            scanTaskExecutor.submit(() -> submitDsScanTask(taskId, userId));
        }
        return taskCreateDto;
    }
}
