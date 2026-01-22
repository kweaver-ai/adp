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
import com.eisoo.dc.common.enums.*;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.*;
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
import com.eisoo.dc.metadata.domain.vo.*;
import com.eisoo.dc.metadata.service.*;
import com.google.common.util.concurrent.Futures;
import com.google.common.util.concurrent.ListenableFuture;
import com.google.common.util.concurrent.ListeningExecutorService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.concurrent.ThreadPoolTaskScheduler;
import org.springframework.scheduling.config.CronTask;
import org.springframework.scheduling.support.CronExpression;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import javax.annotation.PostConstruct;
import javax.servlet.http.HttpServletRequest;
import java.text.SimpleDateFormat;
import java.util.*;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantLock;
import java.util.stream.Collectors;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class TaskScanServiceImpl extends ServiceImpl<TaskScanMapper, TaskScanEntity> implements ITaskScanService {
    private final ReentrantLock lock = new ReentrantLock(true);

    @Autowired
    private ThreadPoolTaskScheduler threadPoolTaskScheduler;
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
    @Qualifier("dsScheduleScanExecutor")
    private ListeningExecutorService dsScheduleScanExecutor;
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
    private ConnectorConfigCache connectorConfigCache;
    @Autowired
    private ITaskScanScheduleService taskScanScheduleService;

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
        // 构造返回
        List<TableStatusDto.TableStatus> list = new ArrayList<>(tables.size());
        try {
            submitTablesScanTaskInner(listPre, 1, userId, true);
        } catch (Exception e) {
            throw new RuntimeException(e);
        } finally {
            List<TaskScanTableEntity> listPost = taskScanTableMapper.selectByTaskIdAndIds(taskId, tables);
            // 更新结果信息
            updateScanTaskResultInfo(taskId, listPost);
            for (TaskScanTableEntity table : listPost) {
                String tableId = table.getTableId();
                TableStatusDto.TableStatus tableStatus = new TableStatusDto.TableStatus();
                tableStatus.setTableId(tableId);
                tableStatus.setStatus(ScanStatusEnum.fromCode(table.getScanStatus()));
                list.add(tableStatus);
            }
            tableStatusDto.setTables(list);
        }
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
        List<TaskScanTableDto> results = entities.stream().map(t -> {
            Integer scanStatus = t.getScanStatus();
            TaskScanTableDto taskScanTableDto = new TaskScanTableDto(t.getTaskId(),
                    t.getTableId(),
                    t.getTableName(),
                    ScanStatusEnum.fromCode(scanStatus),
                    t.getStartTime(),
                    null
            );
            if (ScanStatusEnum.FAIL.getCode() == scanStatus || ScanStatusEnum.SUCCESS.getCode() == scanStatus) {
                taskScanTableDto.setEndTime(t.getEndTime());
            }
            return taskScanTableDto;
        }).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> getScanTaskList(String userId,
                                             List<Integer> type,
                                             String dsId,
                                             String status,
                                             String keyword,
                                             int limit,
                                             int offset,
                                             String sort,
                                             String direction) {
        if (CommonUtil.isNotEmpty(type) && type.size() != 0) {
            for (Integer taskType : type) {
                if (taskType != 0 && taskType != 1 && taskType != 2) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "参数type错误",
                            "参数type错误:type只能是[0,1,2]其中一种",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            }
        }
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
        // 查询t_task_scan的type!=3
        List<TaskScanEntity> dsList = taskScanMapper.selectTaskScans(null, dsId, type, statusList, keyword);
        long count = taskScanMapper.selectCount(null, dsId, type, statusList, keyword);

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
            Integer typeTask = t.getType();
            String taskStatus = ScheduleJobStatusEnum.OPEN.getDesc();
            String scheduleId = null;
            if (typeTask == ScanTaskEnm.SCHEDULE_DS.getCode()) {
                TaskScanScheduleEntity taskScanScheduleEntity = taskScanScheduleService.getById(t.getScheduleId());
                if (taskScanScheduleEntity != null) {
                    taskStatus = ScheduleJobStatusEnum.fromCode(taskScanScheduleEntity.getTaskStatus());
                    scheduleId = t.getScheduleId();
                }
            }
            return new TaskScanDto(
                    t.getId(),
                    scheduleId,
                    t.getType(),
                    t.getName(),
                    ds.getFType(),
                    DbConnectionStrategyFactory.supportNewScan(ds.getFType()),
                    userName,
                    ScanStatusEnum.fromCode(t.getScanStatus()),
                    taskStatus,
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
        String response;
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
        Integer type = null;
        String dsType = null;
        TaskScanEntity taskScanEntityStart = null;
        try {
            taskScanEntityStart = taskScanMapper.selectById(taskId);
            type = taskScanEntityStart.getType();
            taskScanMapper.updateScanStatusStart(taskId, ScanStatusEnum.RUNNING.getCode());
            String dsId = taskScanEntityStart.getDsId();
            DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
            dsType = dataSourceEntity.getFType();
            log.info("【元数据扫描】dsType:{};taskId:{};taskType:{}:更新task扫描状态:{}-成功;开始进行【stage-1】:获取table元数据......",
                    dsType,
                    taskId,
                    type,
                    ScanStatusEnum.RUNNING);
        } catch (Exception e) {
            assert taskScanEntityStart != null;
            metaDataFetchServiceBase.saveFail(taskScanEntityStart,
                    1,
                    e.getMessage());
            log.error("【元数据扫描:stage-1】dsType:{};taskId:{}获取table失败,退出!",
                    dsType,
                    taskId,
                    e);
            throw new RuntimeException(e);
        }
        // 1.2,ds扫描，首先获取所有的table
        if (ScanTaskEnm.IMMEDIATE_TABLES.getCode() != type) {
            try {
                metaDataFetchServiceBase.getTables(taskScanEntityStart, userId);
            } catch (Exception e) {
                log.error("【元数据扫描:stage-1】dsType:{};taskId:{}获取table失败,退出!",
                        dsType,
                        taskId,
                        e);
                throw new RuntimeException(e);
            }
        }
        log.info("【元数据扫描:stage-1】dsType:{};taskId:{};taskType:{}: 获取tables结束，开始进行【stage-2】-获取field元数据",
                dsType,
                taskId,
                type);
        //2,提交field
        submitTablesScanTask(taskId, userId);
    }

    @Override
    public void submitDsScheduleScanTask(String jobId, String userId) {
        dsScheduleScanExecutor.submit(() -> submitDsScheduleScanTaskInternal(jobId, userId));
    }

    @Override
    public void submitTablesScanTask(String taskId, String userId) {
        // 0,更新t_task_scan
        TaskScanEntity taskScanEntityStart = taskScanMapper.selectById(taskId);
        // 1,查出来table信息
        List<TaskScanTableEntity> taskScanTableEntities = taskScanTableMapper.selectByTaskId(taskId);
        try {
            submitTablesScanTaskInner(taskScanTableEntities,
                    taskScanEntityStart.getType(),
                    userId,
                    false);
        } catch (Exception ignored) {
        } finally {
            // 更新结果状态
            updateScanTaskResultInfo(taskId, taskScanTableEntities);
        }
    }

    @Override
    @Transactional(rollbackFor = Exception.class, propagation = Propagation.REQUIRES_NEW)
    public void updateByIdNewRequires(TaskScanEntity taskScanEntity) {
        updateById(taskScanEntity);
    }

    /***
     * 修改定时扫描任务状态
     * @param request:请求request信息
     * @param req:定时任务信息
     * @return:操作结果信息
     */
    @Override
    public ResponseEntity<?> changeScanStatus(HttpServletRequest request, ScheduleJobStatusVo req) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        String targetStatus = req.getStatus();
        String jobId = req.getScheduleId();
        TaskScanScheduleEntity taskScanScheduleEntity = taskScanScheduleService.getById(jobId);
        // 查询
        if (null == taskScanScheduleEntity) {
            // 查询
            TaskScanEntity taskScanEntity = taskScanMapper.selectById(jobId);
            if (null != taskScanEntity && taskScanEntity.getType() != ScanTaskEnm.SCHEDULE_DS.getCode()) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "即时扫描任务不支持修改状态。",
                        "即时扫描任务不支持修改状态:id=" + jobId,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "定时扫描任务不存在。",
                    "定时扫描任务不存在:id=" + jobId,
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        Integer taskStatusCurrent = taskScanScheduleEntity.getTaskStatus();
        if (ScheduleJobStatusEnum.fromCode(taskStatusCurrent).equals(targetStatus)) {
            JSONObject response = new JSONObject();
            response.put("schedule_id", jobId);
            response.put("status", "success");
            return ResponseEntity.ok(response);
        }
        String cronExpression = taskScanScheduleEntity.getCronExpression();
        TaskScanVO.CronExpressionObj cronExpressionObj = JSONObject.parseObject(cronExpression, TaskScanVO.CronExpressionObj.class);
        String expression = cronExpressionObj.getExpression();
        String type = cronExpressionObj.getType();
        log.info("处理之前：type:{};expression:{}", type, expression);
        if ("FIX_RATE".equalsIgnoreCase(type)) {
            expression = getCronExpression(expression);
        }
        log.info("处理之后：type:{};expression:{}", type, expression);
        if (ScheduleJobStatusEnum.CLOSE.getDesc().equals(targetStatus)) {
            // ENABLE->DISABLE
            try {
                removeScheduleTask(jobId);
            } catch (Exception e) {
                log.error("【元数据扫描】定时任务status切换失败:scheduleJobId:{}", jobId, e);
                throw new AiShuException(ErrorCodeEnum.InternalError,
                        e.getMessage(),
                        Message.MESSAGE_INTERNAL_ERROR
                );
            }
        } else if (ScheduleJobStatusEnum.OPEN.getDesc().equals(targetStatus)) {
            // DISABLE->ENABLE
            boolean getLock = LockUtil.SCHEDULE_SCAN_TASK_LOCK.tryLock(jobId,
                    0,
                    TimeUnit.SECONDS,
                    true);
            if (getLock) {
                Runnable runnable = () -> submitDsScheduleScanTask(jobId, userId);
                CronTask cronTask = new CronTask(runnable, expression);
                // 4. 注册任务并保存到映射表
                ScheduledFuture<?> scheduledTask = threadPoolTaskScheduler.schedule(
                        cronTask.getRunnable(),
                        cronTask.getTrigger()
                );
                CommonUtil.SCHEDULE_JOB_MAP.put(jobId, scheduledTask);
                log.info("【元数据扫描】定时任务status切换:任务ID:{};Cron表达式:{}:任务注册成功",
                        jobId,
                        cronExpression
                );
                if (LockUtil.SCHEDULE_SCAN_TASK_LOCK.isHoldingLock(jobId)) {
                    LockUtil.SCHEDULE_SCAN_TASK_LOCK.unlock(jobId);
                }
            }
        }
        log.info("【元数据扫描】定时任务status切换:jobId:{};tartget:{}",
                jobId,
                targetStatus);
        try {
            //2,更新t_task_scan_schedule
            taskScanScheduleEntity.setTaskStatus(ScheduleJobStatusEnum.fromDesc(targetStatus));
            taskScanScheduleEntity.setOperationUser(userId);
            taskScanScheduleEntity.setOperationTime(new Date());
            taskScanScheduleEntity.setOperationType(OperationTyeEnum.UPDATE.getCode());
            taskScanScheduleService.updateById(taskScanScheduleEntity);
        } catch (Exception e) {
            log.error("【元数据扫描】定时任务status切换:持久化db失败！scheduleJobId:{}", jobId, e);
            throw new AiShuException(ErrorCodeEnum.InternalError,
                    e.getMessage(),
                    Message.MESSAGE_INTERNAL_ERROR
            );
        }
        JSONObject response = new JSONObject();
        response.put("schedule_id", jobId);
        response.put("status", "success");
        return ResponseEntity.ok(response);
    }

    /***
     * 获取定时任务信息
     * @param request:请求request信息
     * @param jobId:定时任务id
     * @return: 定时任务信息
     */
    @Override
    public ResponseEntity<?> getScheduleScanJob(HttpServletRequest request, String jobId, Integer type) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        // 对Integer type进行参数校验
        if (null == type) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "参数type不能为空",
                    "参数type不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (ScanTaskEnm.IMMEDIATE_DS.getCode() != type && ScanTaskEnm.SCHEDULE_DS.getCode() != type) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "参数type错误",
                    "参数type错误:type只能时[0,2]其中一种",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        ScheduleTaskStatusInfoDto scheduleTaskStatusInfoDto = new ScheduleTaskStatusInfoDto();
        TaskScanEntity taskScanEntity;
        if (ScanTaskEnm.SCHEDULE_DS.getCode() == type) {
            TaskScanScheduleEntity taskScanScheduleEntity = taskScanScheduleService.getById(jobId);
            // 查询
            if (null == taskScanScheduleEntity) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "定时扫描任务不存在。",
                        "定时扫描任务不存在:id=" + jobId,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            taskScanEntity = taskScanMapper.selectLastScheduleScanTask(jobId);
            if (null == taskScanEntity) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "定时扫描-最近一次任务不存在。",
                        "定时扫描-最近一次任务不存在:id=" + jobId,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            // insert,update,delete
            String scanStrategy = taskScanScheduleEntity.getScanStrategy();
            String cronExpression = taskScanScheduleEntity.getCronExpression();

            if (CommonUtil.isNotEmpty(scanStrategy)) {
                scheduleTaskStatusInfoDto.setScanStrategy(Arrays.asList(scanStrategy.split(",")));
            }
            if (CommonUtil.isNotEmpty(cronExpression)) {
                scheduleTaskStatusInfoDto.setCronExpression(JSONObject.parseObject(cronExpression, TaskScanVO.CronExpressionObj.class));
            }
            scheduleTaskStatusInfoDto.setTaskStatus(ScheduleJobStatusEnum.fromCode(taskScanScheduleEntity.getTaskStatus()));

        } else {
            taskScanEntity = taskScanMapper.selectById(jobId);
            if (null == taskScanEntity) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "即时扫描任务不存在。",
                        "即时扫描任务不存在:id=" + jobId,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            String taskParamsInfo = taskScanEntity.getTaskParamsInfo();
            TaskScanEntity.TaskParamsInfo info = JSONObject.parseObject(taskParamsInfo, TaskScanEntity.TaskParamsInfo.class);
            List<String> scanStrategy = info.getScanStrategy();
            scheduleTaskStatusInfoDto.setScanStrategy(scanStrategy);
            scheduleTaskStatusInfoDto.setTaskStatus(ScheduleJobStatusEnum.OPEN.getDesc());
            scheduleTaskStatusInfoDto.setCronExpression(null);
        }
        // 获取最近一次执行
        scheduleTaskStatusInfoDto.setLastScanTaskId(taskScanEntity.getId());
        Integer scanStatus = taskScanEntity.getScanStatus();
        Date startTime = taskScanEntity.getStartTime();
        Date endTime = taskScanEntity.getEndTime();

        long startMillis = startTime.getTime();
        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
        if (null == endTime) {
            endTime = new Date();
        }
        long endMillis = endTime.getTime();
        scheduleTaskStatusInfoDto.setDuration(String.valueOf((endMillis - startMillis) / 1000));

        scheduleTaskStatusInfoDto.setStartTime(sdf.format(startTime));
        if (ScanStatusEnum.FAIL.getCode() == scanStatus || ScanStatusEnum.SUCCESS.getCode() == scanStatus) {
            scheduleTaskStatusInfoDto.setEndTime(sdf.format(endTime));
        }
        scheduleTaskStatusInfoDto.setScanStatus(ScanStatusEnum.fromCode(scanStatus));
        return ResponseEntity.ok(scheduleTaskStatusInfoDto);
    }

    /***
     * 获取定时任务执行列表
     * @param request:请求request信息
     * @param scheduleId:定时任务id
     * @param limit:分页条数
     * @param offset:分页偏移量
     * @return: 定时任务信息
     */
    @Override
    public ResponseEntity<?> getScheduleScanExecList(HttpServletRequest request, String scheduleId, int limit, int offset) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        int totalCount = taskScanMapper.selectScheduleScanExecCount(scheduleId);
        List<TaskScanEntity> taskScanEntities = taskScanMapper.selectScheduleScanExecList(scheduleId, limit, offset);
        JSONObject response = new JSONObject();
        JSONArray entries = new JSONArray();

        if (totalCount == 0) {
            response.put("entries", entries);
            response.put("total_count", totalCount);
            return ResponseEntity.ok(response);
        }
        List<ScheduleTaskInfoDto> results = new ArrayList<>(taskScanEntities.size());
        for (TaskScanEntity taskScanEntity : taskScanEntities) {
            ScheduleTaskInfoDto scheduleTaskInfoDto = new ScheduleTaskInfoDto();
            Integer scanStatus = taskScanEntity.getScanStatus();
            Date startTime = taskScanEntity.getStartTime();
            Date endTime = taskScanEntity.getEndTime();
            long startMillis = startTime.getTime();
            SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

            scheduleTaskInfoDto.setTaskId(taskScanEntity.getId());
            scheduleTaskInfoDto.setStartTime(sdf.format(startTime));
            scheduleTaskInfoDto.setScanStatus(ScanStatusEnum.fromCode(scanStatus));
            // 结束有endTime
            if (ScanStatusEnum.FAIL.getCode() == scanStatus || ScanStatusEnum.SUCCESS.getCode() == scanStatus) {
                if (null == endTime) {
                    endTime = new Date();
                } else {
                    scheduleTaskInfoDto.setEndTime(sdf.format(endTime));
                }
            }
            long endMillis = endTime.getTime();
            scheduleTaskInfoDto.setDuration(String.valueOf((endMillis - startMillis) / 1000));
            scheduleTaskInfoDto.setTaskProcessInfo(taskScanEntity.getTaskProcessInfo());
            scheduleTaskInfoDto.setTaskResultInfo(taskScanEntity.getTaskResultInfo());
            results.add(scheduleTaskInfoDto);
        }
        response.put("entries", results);
        response.put("total_count", totalCount);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> updateScheduleScanJob(HttpServletRequest request, ScheduleTaskScanVO scheduleTaskScanVO) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String token = CommonUtil.getToken(request);
        log.info("userId:{}", userId);
        log.info("token:{}", token);
        String scheduleId = scheduleTaskScanVO.getScheduleId();
        List<String> scanStrategy = scheduleTaskScanVO.getScanStrategy();
        // 校验scanStrategy
        if (CommonUtil.isNotEmpty(scanStrategy)) {
            validateScanStrategy(scanStrategy);
            scanStrategy = scanStrategy.stream().distinct().collect(Collectors.toList());
        }
        // 校验CronExpression
        TaskScanVO.CronExpressionObj cronExpression = scheduleTaskScanVO.getCronExpression();
        validateCronExpressionObj(cronExpression);
        String typeCron = cronExpression.getType();
        String expression = cronExpression.getExpression();
        String status = scheduleTaskScanVO.getStatus();
        validateScheduleStatus(status);
        TaskScanScheduleEntity taskScanScheduleEntity = taskScanScheduleService.getById(scheduleId);
        // 查询
        if (null == taskScanScheduleEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "定时扫描任务不存在。",
                    "定时扫描任务不存在:id=" + scheduleId,
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        String cronPre = taskScanScheduleEntity.getCronExpression();
        Integer taskStatusPre = taskScanScheduleEntity.getTaskStatus();
        TaskScanVO.CronExpressionObj cronObjPre = JSONObject.parseObject(cronPre, TaskScanVO.CronExpressionObj.class);
        boolean isCronChanged = !cronObjPre.getType().equals(typeCron)
                || !expression.equals(cronObjPre.getExpression())
                || !status.equals(ScheduleJobStatusEnum.fromCode(taskStatusPre));
        //2,更新t_task_scan_schedule
        if (scanStrategy != null && scanStrategy.size() > 0) {
            String scanStrategyStr = String.join(",", scanStrategy);
            taskScanScheduleEntity.setScanStrategy(scanStrategyStr);
        } else {
            taskScanScheduleEntity.setScanStrategy(null);
        }
        taskScanScheduleEntity.setCronExpression(JSONObject.toJSONString(cronExpression));
        taskScanScheduleEntity.setTaskStatus(ScheduleJobStatusEnum.fromDesc(status));
        taskScanScheduleEntity.setOperationUser(userId);
        taskScanScheduleEntity.setOperationTime(new Date());
        taskScanScheduleEntity.setOperationType(OperationTyeEnum.UPDATE.getCode());
        taskScanScheduleService.updateEntityById(taskScanScheduleEntity);
        // 发生了变更更新定时任务
        if (isCronChanged) {
            try {
                removeScheduleTask(scheduleId);
                // 如果新任务的状态是open
                if (status.equals(ScheduleJobStatusEnum.OPEN.getDesc())) {
                    boolean getLock = LockUtil.SCHEDULE_SCAN_TASK_LOCK.tryLock(scheduleId,
                            0,
                            TimeUnit.SECONDS,
                            true);
                    if (getLock) {
                        Runnable runnable = () -> submitDsScheduleScanTask(scheduleId, userId);
                        String cron = cronExpression.getExpression();
                        if ("FIX_RATE".equalsIgnoreCase(cronExpression.getType())) {
                            cron = getCronExpression(cron);
                        }
                        CronTask cronTask = new CronTask(runnable, cron);
                        // 4. 注册任务并保存到映射表
                        ScheduledFuture<?> scheduledTask = threadPoolTaskScheduler.schedule(
                                cronTask.getRunnable(),
                                cronTask.getTrigger()
                        );
                        CommonUtil.SCHEDULE_JOB_MAP.put(scheduleId, scheduledTask);
                        if (LockUtil.SCHEDULE_SCAN_TASK_LOCK.isHoldingLock(scheduleId)) {
                            LockUtil.SCHEDULE_SCAN_TASK_LOCK.unlock(scheduleId);
                        }
                    }
                }
            } catch (Exception e) {
                log.error("【元数据扫描】update定时任务失败:scheduleJobId:{}", scheduleId, e);
                throw new AiShuException(ErrorCodeEnum.InternalError,
                        e.getMessage(),
                        Message.MESSAGE_INTERNAL_ERROR
                );
            }
        }
        JSONObject response = new JSONObject();
        response.put("schedule_id", scheduleId);
        response.put("status", "success");
        return ResponseEntity.ok(response);
    }

    private void submitTablesScanTaskInner(List<TaskScanTableEntity> taskScanTableEntities,
                                           Integer type,
                                           String userId,
                                           boolean isRetry) throws Exception {
        int totalCount = taskScanTableEntities.size();
        String taskId = null;
        // 把delete的持久化到中间结果
        if (!isRetry) {
            List<TaskScanTableEntity> deletes = taskScanTableEntities.stream()
                    .filter(t -> t.getOperationType().equals(OperationTyeEnum.DELETE.getCode()))
                    .collect(Collectors.toList());
            List<String> flags = new ArrayList<>(deletes.size());
            for (TaskScanTableEntity delete : deletes) {
                if (taskId == null) {
                    taskId = delete.getTaskId();
                }
                flags.add(ScanStatusEnum.SUCCESS.getDesc());
            }
            if (deletes.size() > 0) {
                updateScanTaskProcessInfo(taskId, totalCount, flags);
            }
        }
        List<TaskScanTableEntity> waitingTableEntities = new ArrayList<>(taskScanTableEntities.size());
        List<ListenableFuture<String>> futures = new ArrayList<>();
        List<String> lockIds = new ArrayList<>();
        for (TaskScanTableEntity tableEntity : taskScanTableEntities) {
            String tableId = tableEntity.getTableId();
            taskId = tableEntity.getTaskId();
            String tableName = tableEntity.getTableName();
            String dsId = tableEntity.getDsId();
            if (tableEntity.getOperationType().equals(OperationTyeEnum.DELETE.getCode())) {
                continue;
            }
            try {
                DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
                ConnectorConfig connectorConfig = connectorConfigCache.getConnectorConfig(dataSourceEntity.getFType());
                boolean getLock = LockUtil.GLOBAL_MULTI_TASK_LOCK.tryLock(tableId,
                        2,
                        TimeUnit.SECONDS,
                        true);
                if (getLock) {
                    // 这个后面要释放
                    log.info("【元数据扫描:stage-2】taskId:{};tableName:{}:采集任务获取锁成功，准备提交到pool执行......",
                            taskId,
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
                    if (type == ScanTaskEnm.IMMEDIATE_TABLES.getCode()) {
                        // 指定table的扫描任务
                        futures.add(tableScanExecutor.submit(fieldFetchCallable));
                    } else {
                        // 数据源类型扫描任务
                        futures.add(dsScanExecutor.submit(fieldFetchCallable));
                    }
                } else {
                    log.warn("【元数据扫描:stage-2】taskId:{};tableName:{}：采集任务获取锁失败，存放到waiting列表下次执行......",
                            taskId,
                            tableName);
                    waitingTableEntities.add(tableEntity);
                }
            } catch (Exception e) {
                log.error("【元数据扫描】taskId:{}:在stage2阶段失败，退出!",
                        taskId,
                        e);
                throw new Exception(e);
            }
            if (lockIds.size() >= scanTaskPoolConfig.getNumThread()) {
                // 到达数量，提交
                ListenableFuture<List<String>> listListenableFuture = Futures.allAsList(futures);
                List<String> waitMainThread = null;
                try {
                    waitMainThread = listListenableFuture.get();
                } catch (Exception e) {
                    log.warn("【元数据扫描:stage-2】有table采集field任务失败了", e);
                } finally {
                    for (String id : lockIds) {
                        // 确保释放锁（只释放当前线程持有的锁）
                        if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                            LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
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
                log.warn("【元数据扫描:stage-2】有table采集field任务失败了", e);
            } finally {
                for (String id : lockIds) {
                    if (LockUtil.GLOBAL_MULTI_TASK_LOCK.isHoldingLock(id)) {
                        LockUtil.GLOBAL_MULTI_TASK_LOCK.unlock(id);
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
                if (flag.equals(ScanStatusEnum.fromCode(ScanStatusEnum.FAIL.getCode()))) {
                    ++fail;
                } else if (flag.equals(ScanStatusEnum.fromCode(ScanStatusEnum.SUCCESS.getCode()))) {
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
        if (!ids.isEmpty()) {
            List<TaskScanTableEntity> tables = taskScanTableMapper.selectBatchIds(ids);
            for (TaskScanTableEntity table : tables) {
                Integer scanStatus = table.getScanStatus();
                if (scanStatus == ScanStatusEnum.SUCCESS.getCode()) {
                    successCount++;
                } else if (scanStatus == ScanStatusEnum.FAIL.getCode()) {
                    failCount++;
                }
            }
        }
        TaskStatusInfoDto.TaskResultInfo taskResultInfo = new TaskStatusInfoDto.TaskResultInfo();
        taskResultInfo.setTableCount(tablesCount);
        taskResultInfo.setSuccessCount(successCount);
        taskResultInfo.setFailCount(failCount);
        TaskScanEntity taskScanEntity = taskScanMapper.selectById(taskId);
        //如果没有中间进度信息说明没有table，这个时候要补全
        String taskProcessInfo = taskScanEntity.getTaskProcessInfo();
        if (CommonUtil.isEmpty(taskProcessInfo)) {
            TaskStatusInfoDto.TaskProcessInfo taskProcess = new TaskStatusInfoDto.TaskProcessInfo();
            taskProcess.setTableCount(0);
            taskProcess.setSuccessCount(0);
            taskProcess.setFailCount(0);
            String jsonString = JSONObject.toJSONString(taskProcess, WriteMapNullValue);
            taskScanEntity.setTaskProcessInfo(jsonString);
        }
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
        List<String> tables = taskScanVO.getTables();
        Set<String> allConnectors = ConnectorEnums.getAllConnectors();
        String dsId = dsInfo.getDsId();
        String dsType = dsInfo.getDsType();
        List<String> scanStrategy = dsInfo.getScanStrategy();
        // 非 OPEN_SEARCH只支持数据源类型扫描任务
        if (!DbConnectionStrategyFactory.supportNewScan(dsType)) {
            if (type == ScanTaskEnm.IMMEDIATE_TABLES.getCode()) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        Description.DS_SCAN_SUPPORTED_ERROR,
                        dsType + ":不支持多表扫描",
                        Message.MESSAGE_INTERNAL_ERROR);
            } else if (type == ScanTaskEnm.SCHEDULE_DS.getCode()) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        Description.DS_SCAN_SUPPORTED_ERROR,
                        dsType + ":不支持定时扫描",
                        Message.MESSAGE_INTERNAL_ERROR);
            } else if (scanStrategy != null && scanStrategy.size() > 0) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "不支持扫描策略",
                        dsType + ":不支持使用扫描策略",
                        Message.MESSAGE_INTERNAL_ERROR);
            }
        }
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
        } else if (type == ScanTaskEnm.IMMEDIATE_TABLES.getCode()) {
            // tables不能没有数据，当type=1的时候
            if (tables == null || tables.isEmpty()) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        Description.TABLES_SCAN_EMPTY_ERROR,
                        Detail.TABLES_NOT_EMPLOY,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            // 查询这些table是否属于这个dsId
            List<TableScanEntity> tableScanEntities = tableScanService.selectBatchIds(tables);
            HashSet<String> ids = new HashSet<>();
            for (TableScanEntity table : tableScanEntities) {
                String dsId2 = table.getFDataSourceId();
                String fId = table.getFId();
                if (tables.contains(fId) && !dsId2.equals(dsId)) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "table_id与ds_id不匹配",
                            "table:" + fId + " 不属于ds_id：" + dsId + " 这个数据源",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
                ids.add(fId);
            }
            for (String tableId : tables) {
                if (!ids.contains(tableId)) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "table_id不存在",
                            "table_id:" + tableId + "不存在",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            }
        }
        // 任务类型参数校验
        if (CommonUtil.isEmpty(type)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "type是空",
                    "type不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (ScanTaskEnm.getByCode(type) == null) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_SCAN_SUPPORTED_ERROR,
                    String.format(Detail.META_DATA_SCAN__TASK_UNSUPPORTED, type),
                    Message.MESSAGE_INTERNAL_ERROR);
        } else if (ScanTaskEnm.SCHEDULE_DS.getCode() == type || ScanTaskEnm.IMMEDIATE_DS.getCode() == type) {
            if (CommonUtil.isNotEmpty(scanStrategy)) {
                validateScanStrategy(scanStrategy);
            }
        }
        // 定时扫描的cron校验
        if (ScanTaskEnm.SCHEDULE_DS.getCode() == type) {
            List<TaskScanScheduleEntity> activeTaskScanSchedule = taskScanScheduleService.getTaskScanSchedule(dsId);
            if (activeTaskScanSchedule.size() > 0) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "定时扫描任务已存在",
                        "定时扫描任务已存在，不能重复创建：dsId：" + dsId,
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            TaskScanVO.CronExpressionObj cronExpressionObj = taskScanVO.getCronExpression();
            validateCronExpressionObj(cronExpressionObj);
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
                    String.format(Detail.RESOURCE_PERMISSION_ERROR,
                            ResourceAuthConstant.RESOURCE_OPERATION_TYPE_SCAN)
            );
        }
        // 检查数据源是否正在扫描:当type=0时
        if (type == ScanTaskEnm.IMMEDIATE_DS.getCode()) {
            int runningDs = taskScanMapper.getRunningDs(dsId);
            if (runningDs != 0) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        Description.DS_SCAN_RUNNING_ERROR,
                        String.format(Detail.DATASOURCE_SCAN_IS_RUNNING, dsId),
                        Message.MESSAGE_INTERNAL_ERROR);
            }
        }
    }

    private TaskCreateDto buildScanTask(String userId, TaskScanVO taskScanVO, String token) {
        TaskScanVO.DsInfo dsInfo = taskScanVO.getDsInfo();
        String dsId = dsInfo.getDsId();
        String dsType = dsInfo.getDsType();
        List<String> scanStrategy = dsInfo.getScanStrategy();
        // 对scanStrategy去重
        if (scanStrategy != null && scanStrategy.size() > 0) {
            scanStrategy = scanStrategy.stream().distinct().collect(Collectors.toList());
        }
        Integer type = taskScanVO.getType();
        List<String> tables = taskScanVO.getTables();
        Date now = new Date();
        String taskId = UUID.randomUUID().toString();
        // 创建扫描任务
        TaskCreateDto taskCreateDto = new TaskCreateDto(
                taskId,
                dsId,
                ScanStatusEnum.WAIT.getDesc()
        );
        boolean isScheduleTask = type == ScanTaskEnm.SCHEDULE_DS.getCode();
        // 定时数据源:2
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
                null,
                null
        );
        if (!isScheduleTask) {
            TaskScanEntity.TaskParamsInfo taskParamsInfo = new TaskScanEntity.TaskParamsInfo(
                    null,
                    null,
                    null,
                    scanStrategy);
            if (ScanTaskEnm.IMMEDIATE_TABLES.getCode() == type) {
                // 对tables去重
                List<String> tablesDistinct = tables.stream().distinct().collect(Collectors.toList());
                taskParamsInfo.setTables(tablesDistinct);
                taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParamsInfo, WriteMapNullValue));
                // 持久化扫描任务
                taskScanMapper.insert(taskScanEntity);
                taskScanTableService.insertBatch(tablesDistinct, dsId, taskId, userId);
            } else {
                taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParamsInfo, WriteMapNullValue));
                taskScanMapper.insert(taskScanEntity);
            }
            // 目前支持的：maria mysql open search
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
            log.info("创建扫描任务成功:{}", taskScanEntity);
        } else {
            try {
                String jobId = createScheduleJob(userId, taskScanVO);
                log.info("创建定时扫描任务成功:jobId:{}", jobId);
            } catch (Exception e) {
                throw new AiShuException(ErrorCodeEnum.CalculateError, e.getMessage());
            }
        }
        return taskCreateDto;
    }

    /**
     * 创建定时扫描任务
     *
     * @param userId     用户id
     * @param taskScanVO 请求参数
     * @return 定时jobId
     */
    private String createScheduleJob(String userId, TaskScanVO taskScanVO) throws Exception {
        String jobId = UUID.randomUUID().toString();
        boolean getLock = LockUtil.SCHEDULE_SCAN_TASK_LOCK.tryLock(jobId,
                0,
                TimeUnit.SECONDS,
                true);
        if (getLock) {
            TaskScanVO.DsInfo dsInfo = taskScanVO.getDsInfo();
            String dsId = dsInfo.getDsId();
            String dsType = dsInfo.getDsType();
            List<String> scanStrategy = dsInfo.getScanStrategy();
            // 对scanStrategy去重
            if (scanStrategy != null && scanStrategy.size() > 0) {
                scanStrategy = scanStrategy.stream().distinct().collect(Collectors.toList());
            }
            String scanStrategyStr = null;
            if (scanStrategy != null && scanStrategy.size() > 0) {
                scanStrategyStr = String.join(",", scanStrategy);
            }
            TaskScanVO.CronExpressionObj cronExpression = taskScanVO.getCronExpression();
            String cron = JSON.toJSONString(cronExpression, WriteMapNullValue);
            Date now = new Date();
            String status = taskScanVO.getStatus();
            TaskScanScheduleEntity taskScanScheduleEntity = new TaskScanScheduleEntity(
                    jobId,
                    taskScanVO.getType(),
                    taskScanVO.getScanName(),
                    dsId,
                    cron,
                    scanStrategyStr,
                    ScheduleJobStatusEnum.fromDesc(status),
                    now,
                    userId,
                    now,
                    userId,
                    OperationTyeEnum.INSERT.getCode()
            );
            String expression = cronExpression.getExpression();
            String type = cronExpression.getType();
            log.info("处理之前：type:{};expression:{}", type, expression);
            if ("FIX_RATE".equalsIgnoreCase(type)) {
                expression = getCronExpression(expression);
            }
            log.info("处理之后：type:{};expression:{}", type, expression);
            // 这里创建一个，否则前端看不到定时任务
            TaskScanEntity taskScanEntity = new TaskScanEntity(
                    UUID.randomUUID().toString(),
                    ScanTaskEnm.SCHEDULE_DS.getCode(),
                    taskScanVO.getScanName(),
                    dsId,
                    ScanStatusEnum.WAIT.getCode(),
                    now,
                    null,
                    userId,
                    null,
                    null,
                    null,
                    jobId
            );
            TaskScanEntity.TaskParamsInfo taskParamsInfo = new TaskScanEntity.TaskParamsInfo(
                    expression,
                    null,
                    null,
                    scanStrategy);
            taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParamsInfo, WriteMapNullValue));
            try {
                // 持久化扫描任务
                taskScanMapper.insert(taskScanEntity);
                taskScanScheduleService.insert(taskScanScheduleEntity);
                // 如果时open
                if (status.equals(ScheduleJobStatusEnum.OPEN.getDesc())) {
                    Runnable runnable = () -> submitDsScheduleScanTask(jobId, userId);
                    CronTask cronTask = new CronTask(runnable, expression);
                    // 4. 注册任务并保存到映射表
                    ScheduledFuture<?> scheduledTask = threadPoolTaskScheduler.schedule(cronTask.getRunnable(),
                            cronTask.getTrigger());
                    CommonUtil.SCHEDULE_JOB_MAP.put(jobId, scheduledTask);
                }
                log.info("任务ID:{};dsType:{};dsId:{};Cron表达式:{}:任务注册成功",
                        jobId,
                        dsType,
                        dsId,
                        cronExpression
                );
            } catch (Exception e) {
                log.error("任务ID:{};dsType:{};dsId:{};Cron表达式:{}:任务注册成功",
                        jobId,
                        dsType,
                        dsId,
                        cronExpression,
                        e
                );
                throw new Exception(e);
            } finally {
                if (LockUtil.SCHEDULE_SCAN_TASK_LOCK.isHoldingLock(jobId)) {
                    LockUtil.SCHEDULE_SCAN_TASK_LOCK.unlock(jobId);
                }
            }
        }

        return jobId;
    }

    /***
     * 删除定时任务
     * @param jobId:定时任务id
     * @throws Exception:异常
     */
    public void removeScheduleTask(String jobId) throws Exception {
        boolean getLock = LockUtil.SCHEDULE_SCAN_TASK_LOCK.tryLock(jobId,
                0,
                TimeUnit.SECONDS,
                true);
        if (getLock) {
            try {
                ScheduledFuture<?> scheduledTask = CommonUtil.SCHEDULE_JOB_MAP.get(jobId);
                if (scheduledTask != null) {
                    // 取消任务调度
                    scheduledTask.cancel(true);
                    // 从映射表移除
                    CommonUtil.SCHEDULE_JOB_MAP.remove(jobId);
                    log.info("定时任务ID：{}，删除成功", jobId);
                } else {
                    log.warn("定时任务ID：{}，不存在，删除失败", jobId);
                }
            } catch (Exception e) {
                log.error("定时任务ID：{}，删除失败", jobId, e);
                throw new Exception(e);
            } finally {
                if (LockUtil.SCHEDULE_SCAN_TASK_LOCK.isHoldingLock(jobId)) {
                    LockUtil.SCHEDULE_SCAN_TASK_LOCK.unlock(jobId);
                }
            }
        }
    }

    /**
     * 提交定时扫描任务
     *
     * @param jobId  定时任务ID
     * @param userId 用户ID
     */
    private void submitDsScheduleScanTaskInternal(String jobId, String userId) {
        TaskScanScheduleEntity taskScanScheduleEntity = taskScanScheduleService.getById(jobId);
        String dsId = taskScanScheduleEntity.getDsId();
        String taskId = UUID.randomUUID().toString();
        String nowTime = CommonUtil.getNowTime();
        Integer type = taskScanScheduleEntity.getType();
        String name = taskScanScheduleEntity.getName();
        String cronExpression = taskScanScheduleEntity.getCronExpression();
        String scanStrategy = taskScanScheduleEntity.getScanStrategy();
        TaskScanVO.CronExpressionObj cronExpressionObj = JSONObject.parseObject(cronExpression, TaskScanVO.CronExpressionObj.class);
        String expression = cronExpressionObj.getExpression();
        String typeCron = cronExpressionObj.getType();
        log.info("处理之前：type:{};expression:{}", typeCron, expression);
        if ("FIX_RATE".equalsIgnoreCase(typeCron)) {
            expression = getCronExpression(expression);
        }
        log.info("处理之后：type:{};expression:{}", typeCron, expression);
        log.info("【元数据定时扫描】开始:jobId:{},name:{},现在时间:{},cron:{}",
                jobId,
                name,
                nowTime,
                cronExpression);
        // 1,上一次任务是否还没有结束 + 这个数据源是否正在扫描
        // 检查数据源是否正在扫描:当type=0时
        int runningDs = taskScanMapper.getRunningDs(dsId);
        int status = ScanStatusEnum.WAIT.getCode();
        if (runningDs != 0) {
            status = ScanStatusEnum.SKIP.getCode();
        }
        // 查询是否有相同名字并且是wait的task
        List<TaskScanEntity> taskScanEntityList = taskScanMapper.getWaitTaskByName(jobId);
        TaskScanEntity taskScanEntity;
        boolean exist = taskScanEntityList.size() == 1 && taskScanEntityList.get(0).getScanStatus() == ScanStatusEnum.WAIT.getCode();
        if (!exist) {
            taskScanEntity = new TaskScanEntity(
                    taskId,
                    type,
                    name,
                    dsId,
                    status,
                    new Date(),
                    null,
                    userId,
                    null,
                    null,
                    null,
                    jobId
            );
        } else {
            taskScanEntity = taskScanEntityList.get(0);
            taskScanEntity.setScanStatus(status);
            taskScanEntity.setStartTime(new Date());
        }
        List<String> list = null;
        if (CommonUtil.isNotEmpty(scanStrategy)) {
            list = Arrays.asList(scanStrategy.split(","));
        }
        TaskScanEntity.TaskParamsInfo taskParamsInfo = new TaskScanEntity.TaskParamsInfo(
                expression,
                null,
                null,
                list);
        taskScanEntity.setTaskParamsInfo(JSON.toJSONString(taskParamsInfo, WriteMapNullValue));
        // 持久化扫描任务
        if (!exist) {
            taskScanMapper.insert(taskScanEntity);
        } else {
            taskScanMapper.updateById(taskScanEntity);
        }
        // 提交任务
        if (status != ScanStatusEnum.SKIP.getCode()) {
            submitDsScanTask(taskScanEntity.getId(), userId);
        }
    }

    /**
     * 每次启动完成需要把t_task_scan_schedule有效的定时任务加载到threadPoolTaskScheduler里执行，
     * 否则宕机之后再启动定时任务不会执行
     */
    @PostConstruct
    @SuppressWarnings("PMD.UnusedPrivateMethod")
    private void init() {
        try {
            lock.lock();
            List<TaskScanScheduleEntity> activeTaskScanSchedule = taskScanScheduleService.getActiveTaskScanSchedule();
            for (TaskScanScheduleEntity taskScanScheduleEntity : activeTaskScanSchedule) {
                String jobId = taskScanScheduleEntity.getId();
                String cronExpression = taskScanScheduleEntity.getCronExpression();
                TaskScanVO.CronExpressionObj cronExpressionObj = JSONObject.parseObject(cronExpression, TaskScanVO.CronExpressionObj.class);
                String expression = cronExpressionObj.getExpression();
                String typeCron = cronExpressionObj.getType();
                log.info("处理之前：type:{};expression:{}", typeCron, expression);
                if ("FIX_RATE".equalsIgnoreCase(typeCron)) {
                    expression = getCronExpression(expression);
                }
                log.info("处理之后：type:{};expression:{}", typeCron, expression);

                String userId = taskScanScheduleEntity.getCreateUser();
                Integer type = taskScanScheduleEntity.getType();
                String dsId = taskScanScheduleEntity.getDsId();
                Runnable runnable = () -> submitDsScheduleScanTask(jobId, userId);
                CronTask cronTask = new CronTask(runnable, expression);
                // 4. 注册任务并保存到映射表
                ScheduledFuture<?> scheduledTask = threadPoolTaskScheduler.schedule(cronTask.getRunnable(),
                        cronTask.getTrigger());
                CommonUtil.SCHEDULE_JOB_MAP.put(jobId, scheduledTask);
                log.info("定时任务ID:{};dsType:{};dsId:{};Cron表达式:{}:load成功",
                        jobId,
                        type,
                        dsId,
                        cronExpression
                );
            }
            log.info("-----------------------------------loadDsScheduleScanTask  success---------------------------------------");
        } catch (Exception e) {
            log.error("定时任务load失败", e);
            throw new AiShuException(ErrorCodeEnum.InternalError, e.getMessage(), Message.MESSAGE_INTERNAL_ERROR);
        } finally {
            lock.unlock();
        }
    }

    private String getCronExpression(String cronExpression) {
        // 3m 3h 3d 3M 3w:0   */5  *   *   * ?
        String expression = null;
        int length = cronExpression.length();
        String flag = cronExpression.substring(length - 1);
        String value = cronExpression.substring(0, length - 1);
        switch (flag) {
            case "m":
                //0   */5  *   *   * ? 分钟级任务：以任务创建时间为基准，下一分钟启动任务
                expression = "0 */" + value + " * * * ?";
                break;
            case "h":
                //0 0 */2 * * ? 以任务创建时间为基准，下一个整点启动任务
                expression = "0 0 */" + value + " * * ?";
                break;
            case "d":
                //0 */2 * * * ? 以任务创建时间为基准，第二天零点启动任务
                expression = "0 0 0 */" + value + " * ?";
                break;
            case "M":
                // 0 0 0 1 */2 *
                expression = "0 0 0 1 */" + value + " ?";
                break;
            case "w":
                // 周级任务：以任务创建时间为基准，下周一零点启动任务
                expression = "0 0 0 ? * 1/" + value;
                break;
        }
        return expression;
    }

    private void validateCronExpressionObj(TaskScanVO.CronExpressionObj cronExpression) {
        // 1,cron_expression不能缺省
        if (CommonUtil.isEmpty(cronExpression)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "cron_expression是空",
                    "定时扫描任务cron_expression不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        String typeCron = cronExpression.getType();
        String expression = cronExpression.getExpression();
        // FIX_RATE+CRON
        if (!"FIX_RATE".equalsIgnoreCase(typeCron) && !"CRON".equalsIgnoreCase(typeCron)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "cron_expression错误",
                    "type必须是[FIX_RATE,CRON]其中之一",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        if (CommonUtil.isEmpty(expression)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "cron_expression是空",
                    "定时扫描任务cron_expression的expression不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        if ("CRON".equalsIgnoreCase(typeCron)) {
            try {
                CronExpression.parse(expression);
                CronTask cronTask = new CronTask(() -> {
                }, expression);
                ScheduledFuture<?> scheduledTask = threadPoolTaskScheduler.schedule(
                        cronTask.getRunnable(),
                        cronTask.getTrigger()
                );
                if (null == scheduledTask) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            "定时扫描任务cron_expression格式错误",
                            Message.MESSAGE_INTERNAL_ERROR);
                } else {
                    scheduledTask.cancel(true);
                }
            } catch (Exception e) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "cron_expression格式错误",
                        "定时扫描任务cron_expression格式错误",
                        Message.MESSAGE_INTERNAL_ERROR);
            }
        } else if ("FIX_RATE".equalsIgnoreCase(typeCron)) {
            if (expression.length() < 2) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "cron_expression格式错误",
                        "定时扫描任务cron_expression必须大于等于两位,例如2m 2d 2h",
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            int length = expression.length();
            String flag = expression.substring(length - 1);
            String value = expression.substring(0, length - 1);
            if (!StringUtils.isNumeric(value)) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "cron_expression格式错误",
                        value + ":不是数字，这个必须是数字",
                        Message.MESSAGE_INTERNAL_ERROR);
            }
            if (!"m".equals(flag) && !"d".equals(flag) && !"h".equals(flag) && !"M".equals(flag) && !"w".equals(flag)) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "cron_expression格式错误",
                        flag + ":不符合规范,必须是[m、d、h、M、w]其中一个",
                        Message.MESSAGE_INTERNAL_ERROR);
            } else if ("m".equals(flag)) {
                if (Integer.parseInt(value) < 1 || Integer.parseInt(value) > 59) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            value + ":不符合规范,必须是[1-59]其中一个",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            } else if ("h".equals(flag)) {
                if (Integer.parseInt(value) < 1 || Integer.parseInt(value) > 23) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            value + ":不符合规范,必须是[1-23]其中一个",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            } else if ("M".equals(flag)) {
                if (Integer.parseInt(value) < 1 || Integer.parseInt(value) > 12) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            value + ":不符合规范,必须是[1-12]其中一个",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            } else if ("d".equals(flag)) {
                if (Integer.parseInt(value) < 1 || Integer.parseInt(value) > 31) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            value + ":不符合规范,必须是[1-31]其中一个",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            } else {
                if (Integer.parseInt(value) < 1 || Integer.parseInt(value) > 7) {
                    throw new AiShuException(ErrorCodeEnum.BadRequest,
                            "cron_expression格式错误",
                            value + ":不符合规范,必须是[1-7]其中一个",
                            Message.MESSAGE_INTERNAL_ERROR);
                }
            }
        }
    }

    private void validateScheduleStatus(String status) {
        if (CommonUtil.isEmpty(status)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "status是空",
                    "定时扫描任务status不能为空",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        if (!"open".equalsIgnoreCase(status) && !"close".equalsIgnoreCase(status)) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    "status错误",
                    "status必须是[open,close]其中之一",
                    Message.MESSAGE_INTERNAL_ERROR);
        }
    }

    private void validateScanStrategy(List<String> scanStrategy) {
        for (String scanStrategyItem : scanStrategy) {
            if (!CommonUtil.OPERATON_TYPES.contains(scanStrategyItem)) {
                throw new AiShuException(ErrorCodeEnum.BadRequest,
                        "scan_strategy参数错误",
                        scanStrategyItem + ":不符合规范",
                        Message.MESSAGE_INTERNAL_ERROR);
            }
        }
    }
}
