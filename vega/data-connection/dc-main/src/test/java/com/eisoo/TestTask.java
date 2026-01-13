package com.eisoo;

import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.metadata.entity.TaskScanTableEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.datasource.service.impl.CatalogServiceImpl;
import com.eisoo.dc.main.Application;
import com.eisoo.dc.metadata.domain.vo.*;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.metadata.mapper.TaskScanTableMapper;
import com.eisoo.dc.metadata.service.ITableScanService;
import com.eisoo.dc.metadata.service.impl.TaskScanServiceImpl;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.ResponseEntity;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;

import static com.alibaba.fastjson2.JSONWriter.Feature.WriteMapNullValue;

@SpringBootTest(classes = {Application.class})
public class TestTask {


    @Autowired
    private OpenSearchMetaDataFetchServiceImpl openSearchMetaDataFetchServiceImpl;

    @Autowired
    private TaskScanServiceImpl taskScanServiceImpl;

    @Autowired
    private TableScanMapper tableScanMapper;
    @Autowired
    private TaskScanTableMapper taskScanTableMapper;
    @Autowired
    private CatalogServiceImpl catalogServiceImpl;
    @Autowired
    private ITableScanService tableScanService;
    @Autowired
    private DataSourceMapper dataSourceMapper;

    @Test
    public void Test() throws Exception {
//        openSearchMetaDataFetchServiceImpl.getTables(null);
        TaskScanTableEntity taskScanTableEntity = taskScanTableMapper.selectById("27f1d00c0ac7d57673bf77449f2dca65");
        taskScanTableEntity.setEndTime(new Date());
        taskScanTableMapper.updateById(taskScanTableEntity);
    }

    @Test
    public void Test2() throws Exception {
        openSearchMetaDataFetchServiceImpl.getFieldsByTable("as-operation-log-document_domain_sync");
    }

    @Test
    public void Test3() throws Exception {
//        ArrayList<String> list = new ArrayList<>();
//        list.add("tab-1");
//        list.add("tab-2");
        TaskScanVO taskScanVO = new TaskScanVO();
        taskScanVO.setType(0);
        taskScanVO.setScanName("扫描一个数据源");
        TaskScanVO.DsInfo dsInfo = new TaskScanVO.DsInfo();
        dsInfo.setDsId("1-2-3-ds");
        dsInfo.setDsType("opensearch");
        taskScanVO.setDsInfo(dsInfo);
        taskScanVO.setTables(null);
//        taskScanServiceImpl.createScanTaskAndStart(null, taskScanVO);
        System.out.println(JSONObject.toJSONString(taskScanVO));
    }

    @Test
    public void Test4() {
        TaskScanVO taskScanVO = new TaskScanVO();
        taskScanVO.setType(0);
        taskScanVO.setScanName("扫描一个数据源");
        TaskScanVO.DsInfo dsInfo = new TaskScanVO.DsInfo();
        dsInfo.setDsId("7303152d-3c16-4b16-9c43-bfb4b6867f0e");
        dsInfo.setDsType("opensearch");
        taskScanVO.setDsInfo(dsInfo);
        taskScanVO.setTables(null);
        taskScanServiceImpl.createScanTaskAndStart(null, taskScanVO);
    }

    @Test
    public void Test5() {
        TaskScanVO taskScanVO = new TaskScanVO();
        taskScanVO.setType(1);
        taskScanVO.setScanName("扫描几个table");
        TaskScanVO.DsInfo dsInfo = new TaskScanVO.DsInfo();
        dsInfo.setDsId("7303152d-3c16-4b16-9c43-bfb4b6867f0e");
        dsInfo.setDsType("opensearch");
        taskScanVO.setDsInfo(dsInfo);
        List<String> list = new ArrayList<>();
        list.add("d1772bad00a214dd112f8f50e143d702");
        list.add("2c9cbe56e00f82c5a85b16f754e6095a");
        list.add("32d3b3f88e60b63b939b3cf1103f24b1");
        taskScanVO.setTables(list);
        taskScanServiceImpl.createScanTaskAndStart(null, taskScanVO);
    }

    // getScanTaskInfo
    @Test
    public void Test6() {
        ResponseEntity<?> scanTaskInfo = taskScanServiceImpl.getScanTaskInfo(null, "a1fbf677-9aa5-44fb-99e8-2e8c89e5c3f2");
        Object body = scanTaskInfo.getBody();
        String jsonString = JSONObject.toJSONString(body, WriteMapNullValue);
        System.out.println("------" + jsonString);
    }

    @Test
    public void Test7() {
        TableStatusVO tableStatusVO = new TableStatusVO();
        tableStatusVO.setId("1f075c5e-9acc-4ed6-bfd9-42e6aafb1a4a");
        ArrayList<String> list = new ArrayList<>();
        list.add("d1772bad00a214dd112f8f50e143d702");
        list.add("2c9cbe56e00f82c5a85b16f754e6095a");
        list.add("32d3b3f88e60b63b939b3cf1103f24b1");
        tableStatusVO.setTables(list);
        ResponseEntity<?> scanTaskInfo = taskScanServiceImpl.getScanTaskStatus(null, tableStatusVO);
        Object body = scanTaskInfo.getBody();
        String jsonString = JSONObject.toJSONString(body, WriteMapNullValue);
        System.out.println("------" + jsonString);
    }

    @Test
    public void Test8() {
        TableRetryVO tableStatusVO = new TableRetryVO();
        tableStatusVO.setId("1f075c5e-9acc-4ed6-bfd9-42e6aafb1a4a");
        ArrayList<String> list = new ArrayList<>();
        list.add("d1772bad00a214dd112f8f50e143d702");
        list.add("2c9cbe56e00f82c5a85b16f754e6095a");
        list.add("32d3b3f88e60b63b939b3cf1103f24b1");
        tableStatusVO.setTables(list);
        ResponseEntity<?> scanTaskInfo = taskScanServiceImpl.retryScanTable(null, tableStatusVO);
        Object body = scanTaskInfo.getBody();
        String jsonString = JSONObject.toJSONString(body, WriteMapNullValue);
        System.out.println("------" + jsonString);
    }

    @Test
    public void Test9() {
        String query = "{\n" +
                "  \"query\": {\n" +
                "     \"match_all\": { \n" +
                "     }\n" +
                "  }\n" +
                "}";
        System.out.println(query);
        QueryStatementVO queryStatementVO = new QueryStatementVO("638561e0-04e1-4d8b-b550-b1d5bc3d463f", "as-operation-log-user_login", query);
        ResponseEntity<?> responseEntity = taskScanServiceImpl.queryDslStatement(null, queryStatementVO);
    }

    @Test
    public void Test10() {
        catalogServiceImpl.deleteResource("bd702fc2-16cf-4d1d-a343-8e74644974a7");
    }

    @Test
    public void Test11() {
        TableIdsVo tableIdsVo = new TableIdsVo();
        List<String> list = new ArrayList<>();
        list.add("1");
        list.add("t2");
        tableIdsVo.setTableIds(list);
        ResponseEntity<?> tableAndFieldDetailBatch = tableScanService.getTableAndFieldDetailBatch("", "", tableIdsVo);
        System.out.println(tableAndFieldDetailBatch.toString());
    }
    @Test
    public void Test12() {
        List<String> strings = dataSourceMapper.selectAllId(null);
        System.out.println(strings.toString());
    }

}
