package com.eisoo.dc.metadata.controller;

import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.StringUtils;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.eisoo.dc.metadata.domain.vo.*;
import com.eisoo.dc.metadata.service.IFieldScanService;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.ITaskScanService;
import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.validation.constraints.Min;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Pattern;
import java.util.List;

@Api(tags = "元数据管理")
@RestController
@Validated
@RequestMapping("/api/data-connection/v1/metadata")
public class MetaDataController {

    @Autowired(required = false)
    private ITaskScanService taskScanService;
    @Autowired(required = false)
    private ITableScanService tableScanService;
    @Autowired(required = false)
    private IFieldScanService fieldScanService;

    @ApiOperation(value = "新增元数据扫描任务", notes = "新增元数据扫描任务接口")
    @PostMapping("/scan")
    public ResponseEntity<?> createScanTaskAndStart(HttpServletRequest request, @Validated @RequestBody TaskScanVO req) {
        return taskScanService.createScanTaskAndStart(request, req);
    }

    @ApiOperation(value = "新增元数据扫描批量任务", notes = "新增元数据扫描批量任务接口")
    @PostMapping("/scan/batch")
    public ResponseEntity<?> createScanTaskAndStartBatch(HttpServletRequest request, @Validated @RequestBody List<TaskScanVO> req) {
        return taskScanService.createScanTaskAndStartBatch(request, req);
    }

    @ApiOperation(value = "查询扫描任务状态", notes = "查询扫描任务状态接口")
    @GetMapping("/scan/{taskId}")
    public ResponseEntity<?> getScanTask(HttpServletRequest request, @PathVariable("taskId") String taskId) {
        return taskScanService.getScanTaskInfo(request, taskId);
    }

    @ApiOperation(value = "查询定时扫描任务状态", notes = "查询定时扫描任务状态接口")
    @GetMapping("/scan/schedule/{scheduleId}")
    public ResponseEntity<?> getScheduleScanJob(HttpServletRequest request, @PathVariable("scheduleId") String scheduleId,
                                                @RequestParam(value = "type", required = true) Integer type) {
        return taskScanService.getScheduleScanJob(request, scheduleId, type);
    }

    @ApiOperation(value = "更新定时扫描任务", notes = "更新定时扫描任务接口")
    @PutMapping("/scan/schedule")
    public ResponseEntity<?> updateScheduleScanJob(HttpServletRequest request, @Validated @RequestBody ScheduleTaskScanVO scheduleTaskScanVO) {
        return taskScanService.updateScheduleScanJob(request, scheduleTaskScanVO);
    }

    @ApiOperation(value = "查询定时扫描任务执行列表", notes = "查询定时扫描任务状态接口")
    @GetMapping("/scan/schedule/task/{scheduleId}")
    public ResponseEntity<?> getScheduleScanExecList(HttpServletRequest request, @PathVariable("scheduleId") String scheduleId,
                                                     @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                                     @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset) {
        return taskScanService.getScheduleScanExecList(request, scheduleId, limit, offset);
    }
//    @ApiOperation(value = "查询table扫描状态", notes = "查询table扫描状态接口")
//    @PostMapping("/scan/status")
//    public ResponseEntity<?> getScanTaskStatus(HttpServletRequest request, @Validated @RequestBody TableStatusVO req) {
//        return taskScanService.getScanTaskStatus(request, req);
//    }

    @ApiOperation(value = "查询扫描任务的table信息", notes = "查询扫描任务的table信息接口")
    @GetMapping("/scan/info/{taskId}")
    public ResponseEntity<?> getScanTaskTable(HttpServletRequest request,
                                              @PathVariable(value = "taskId") @NotBlank String taskId,
                                              @RequestParam(value = "status", required = false, defaultValue = "")
                                              @Pattern(regexp = "^$|wait|running|success|fail", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：''、wait、running、success、fail") String status,
                                              @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                              @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                              @RequestParam(value = "keyword", required = false) String keyword,
                                              @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                              @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc") String direction,
                                              @RequestParam(value = "sort", required = false, defaultValue = "table_name")
                                              @Pattern(regexp = "start_time|table_name", message = "可选参数值：start_time、table_name") String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        return taskScanService.getScanTaskTableStatus(userId, taskId, status, keyword, limit, offset, sort, direction);
    }

    @ApiOperation(value = "table扫描重试", notes = "table扫描重试接口")
    @PostMapping("/scan/retry")
    public ResponseEntity<?> retryScanTable(HttpServletRequest request, @Validated @RequestBody TableRetryVO req) {
        return taskScanService.retryScanTable(request, req);
    }

    @ApiOperation(value = "查询扫描任务列表", notes = "查询扫描任务列表接口")
    @GetMapping("/scan")
    public ResponseEntity<?> getScanTaskList(HttpServletRequest request,
                                             @RequestParam(value = "ds_id", required = false, defaultValue = "") String dsId,
                                             @RequestParam(value = "type", required = false, defaultValue = "") List<Integer> type,
                                             @RequestParam(value = "status", required = false, defaultValue = "")
                                             @Pattern(regexp = "^$|wait|running|success|fail", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：''、wait、running、success、fail") String status,
                                             @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                             @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                             @RequestParam(value = "keyword", required = false) String keyword,
                                             @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                             @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc") String direction,
                                             @RequestParam(value = "sort", required = false, defaultValue = "start_time")
                                             @Pattern(regexp = "start_time|name", message = "可选参数值：start_time、name") String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        return taskScanService.getScanTaskList(userId,type, dsId, status, keyword, limit, offset, sort, direction);
    }

    @ApiOperation(value = "查询指定数据源下的所有表", notes = "查询指定数据源下的所有表接口")
    @GetMapping("/data-source/{dsId}")
    public ResponseEntity<?> getTableListByDsId(HttpServletRequest request,
                                                @PathVariable("dsId") String dsId,
                                                @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                                @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                                @RequestParam(value = "keyword", required = false) String keyword,
                                                @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                                @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc")
                                                String direction,
                                                @RequestParam(value = "sort", required = false, defaultValue = "f_name")
                                                @Pattern(regexp = "f_create_time|f_operation_time|f_name", message = "可选参数值：f_create_time、f_operation_time、f_name")
                                                String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        return tableScanService.getTableListByDsId(userId, dsId, keyword, limit, offset, sort, direction);
    }

    @ApiOperation(value = "查询指定表下的所有列", notes = "查询指定表下的所有列接口")
    @GetMapping("/table/{tableId}")
    public ResponseEntity<?> getFieldListByTableId(HttpServletRequest request,
                                                   @PathVariable("tableId") String tableId,
                                                   @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                                   @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                                   @RequestParam(value = "keyword", required = false) String keyword,
                                                   @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                                   @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc")
                                                   String direction,
                                                   @RequestParam(value = "sort", required = false, defaultValue = "f_field_name")
                                                   @Pattern(regexp = "f_field_name|f_create_time", message = "可选参数值：f_field_name、f_create_time")
                                                   String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        return fieldScanService.getFieldListByTableId(userId, tableId, keyword, limit, offset, sort, direction);
    }

    @PostMapping("/scan/view")
    public ResponseEntity<?> view(HttpServletRequest request, @Validated @RequestBody QueryStatementVO req) {
        return taskScanService.queryDslStatement(request, req);
    }

    @PutMapping("/scan/schedule/status")
    public ResponseEntity<?> changeScheduleJobStatus(HttpServletRequest request, @Validated @RequestBody ScheduleJobStatusVo req) {
        return taskScanService.changeScanStatus(request, req);
    }

}
