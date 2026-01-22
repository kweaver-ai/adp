package com.eisoo.dc.metadata.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.eisoo.dc.common.metadata.entity.TaskScanEntity;
import com.eisoo.dc.metadata.domain.vo.*;
import org.springframework.http.ResponseEntity;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * @author Tian.lan
 */
public interface ITaskScanService extends IService<TaskScanEntity> {
    ResponseEntity<?> createScanTaskAndStart(HttpServletRequest request, TaskScanVO taskScanVO);

    ResponseEntity<?> getScanTaskInfo(HttpServletRequest request, String taskId);

    ResponseEntity<?> getScanTaskStatus(HttpServletRequest request, TableStatusVO req);

    ResponseEntity<?> retryScanTable(HttpServletRequest request, TableRetryVO req);

    ResponseEntity<?> getScanTaskTableStatus(String userId, String taskId, String status, String keyword, int limit, int offset, String sort, String direction);

    ResponseEntity<?> getScanTaskList(String userId, List<Integer> type, String dsId, String status, String keyword, int limit, int offset, String sort, String direction);

    ResponseEntity<?> queryDslStatement(HttpServletRequest request, QueryStatementVO req);

    ResponseEntity<?> createScanTaskAndStartBatch(HttpServletRequest request, List<TaskScanVO> req);

    void submitDsScanTask(String taskId, String userId) throws Exception;

    void submitDsScheduleScanTask(String jobId, String userId);

    void submitTablesScanTask(String taskId, String userId);

    void updateByIdNewRequires(TaskScanEntity taskScanEntity);

    ResponseEntity<?> changeScanStatus(HttpServletRequest request, ScheduleJobStatusVo req);

    ResponseEntity<?> getScheduleScanJob(HttpServletRequest request, String jobId,Integer type);

    ResponseEntity<?> getScheduleScanExecList(HttpServletRequest request, String scheduleId, int limit, int offset);

    ResponseEntity<?> updateScheduleScanJob(HttpServletRequest request, ScheduleTaskScanVO scheduleTaskScanVO);

}
