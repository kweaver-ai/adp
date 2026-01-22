package com.eisoo.dc.metadata.service.impl;

import cn.hutool.json.JSONArray;
import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.JSONObject;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.eisoo.dc.common.constant.Description;
import com.eisoo.dc.common.constant.Detail;
import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableOldEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.mapper.DataSourceMapper;
import com.eisoo.dc.common.metadata.mapper.FieldScanMapper;
import com.eisoo.dc.common.metadata.mapper.TableOldMapper;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.metadata.domain.dto.TableAndFieldDto;
import com.eisoo.dc.metadata.domain.dto.TableScanDto;
import com.eisoo.dc.metadata.domain.vo.DataSourceIdsVo;
import com.eisoo.dc.metadata.domain.vo.TableIdsVo;
import com.eisoo.dc.metadata.service.ITableScanService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.stream.Collectors;

@Service
@Slf4j
public class TableScanServiceImpl extends ServiceImpl<TableScanMapper, TableScanEntity> implements ITableScanService {
    @Autowired(required = false)
    private TableScanMapper tableScanMapper;
    @Autowired(required = false)
    private TableOldMapper tableOldMapper;
    @Autowired(required = false)
    private DataSourceMapper dataSourceMapper;
    @Autowired(required = false)
    private FieldScanMapper fieldScanMapper;

    @Override
    public List<TableScanEntity> selectByDsId(String dsId) {
        return tableScanMapper.selectByDsId(dsId);
    }

    @Override
    public List<TableScanEntity> selectBatchIds(List<String> tables) {
        return tableScanMapper.selectBatchIds(tables);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public void tableScanBatch(List<TableScanEntity> deletes, List<TableScanEntity> updateList, List<TableScanEntity> saveList) {
        if (deletes != null && !deletes.isEmpty()) {
            boolean delete = updateTableScanBatchById(deletes);
        }
        if (!updateList.isEmpty()) {
            updateBatchTableScan(updateList, 100);
        }
        if (!saveList.isEmpty()) {
            saveBatchTableScan(saveList, 100);
        }
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public boolean updateTableScanBatchById(Collection<TableScanEntity> entityList) {
        return updateBatchById(entityList, 1000);
    }


    @Transactional(rollbackFor = Exception.class)
    @Override
    public boolean updateBatchTableScan(Collection<TableScanEntity> entityList, int batch) {
        return updateBatchById(entityList, batch);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public boolean saveBatchTableScan(Collection<TableScanEntity> entityList, int batch) {
        return saveBatch(entityList, batch);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public int updateScanStatusById(String id, int status) {
        return tableScanMapper.updateScanStatusById(id, status);
    }

    @Override
    @Transactional(rollbackFor = Exception.class, propagation = Propagation.REQUIRES_NEW)
    public int updateScanStatusByIdNewRequires(String id, String taskId, int status) {
        return tableScanMapper.updateScanStatusById(id, status);
    }

    @Transactional(rollbackFor = Exception.class)
    @Override
    public int updateScanStatusAndOperationTimeById(String tableId, String taskId, int status) {
        return tableScanMapper.updateScanStatusAndOperationTimeById(tableId, taskId, status);
    }

    @Override
    public ResponseEntity<?> getTableListByDsId(String userId, String dsId, String keyword, int limit, int offset, String sort, String direction) {
        JSONObject response = new JSONObject();
        DataSourceEntity dataSourceEntity = dataSourceMapper.selectById(dsId);
        if (null == dataSourceEntity) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.DS_NOT_FOUND_ERROR,
                    String.format(Detail.DS_NOT_EXIST, dsId),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        String dataSourceType = dataSourceEntity.getFType();
        // 首先从新的表查询，查不出来再去旧的表查询
        List<TableScanEntity> dsList = tableScanMapper.getTableListByDsId(dsId, keyword);
        long count = tableScanMapper.selectCount(dsId, keyword);
        if (dsList.size() == 0) {
            // 从old查询
            return getTableOldListByDsId(userId, dsId, keyword, limit, offset, sort, direction);
        }
        //TODO:这里可以根据userId过滤资源
        Set<String> ids = dsList.stream().map(TableScanEntity::getFId).collect(Collectors.toSet());
        List<TableScanEntity> tableScanEntities = tableScanMapper.selectPage(ids, keyword, offset, limit, sort, direction);
        if (tableScanEntities == null || tableScanEntities.size() == 0) {
            response.put("entries", new JSONArray());
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        List<TableScanDto> results = tableScanEntities.stream().map(t -> {
            TableScanDto tableScanDto = new TableScanDto(
                    t.getFId(),
                    t.getFName(),
                    t.getFCreateTime(),
                    t.getFOperationTime(),
                    null
            );
            String fAdvancedParams = t.getFAdvancedParams();
            String fDataSourceName = t.getFDataSourceName();
            // excel mysql open search 类型的需要高级参数
            if (CommonUtil.isNotEmpty(fAdvancedParams)) {
                setAdvancedParams(dataSourceType,
                        fDataSourceName,
                        fAdvancedParams,
                        t.getFScanSource(),
                        t.getFDescription(),
                        tableScanDto);
            }
            return tableScanDto;
        }).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> getTableListByDsIdsBatch(String userId, String accountType, DataSourceIdsVo
            req, String updateTime, String keyword, int limit, int offset, String sort, String direction) {
        JSONArray entries = new JSONArray();
        JSONObject response = new JSONObject();
        List<String> dsIds = req.getDsIds();
        // 查询存在的数据源
        List<String> dsIdsExist = dataSourceMapper.selectAllId(dsIds);
        if (CommonUtil.isNotEmpty(updateTime)) {
            DateTimeFormatter targetFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
            OffsetDateTime offsetDateTime = OffsetDateTime.parse(updateTime);
            updateTime = offsetDateTime.format(targetFormatter);
        }
        // 查询所有
        List<String> dsList = tableScanMapper.getTableListByDsIdsBatch(dsIds, updateTime, keyword);
        long count = tableScanMapper.selectCountByDsIdsBatch(dsIds, updateTime, keyword);
        if (dsList.size() == 0) {
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        //TODO:这里可以根据userId过滤资源
        Set<String> ids = new HashSet<>(dsList);
        List<TableScanEntity> tableScanEntities = tableScanMapper.selectPageBatch(ids, keyword, offset, limit, sort, direction)
                .stream()
                .filter(t -> dsIdsExist.contains(t.getFDataSourceId()))
                .collect(Collectors.toList());
        List<TableScanDto> results = tableScanEntities.stream().map(t -> {
            TableScanDto tableScanDto = new TableScanDto();
            tableScanDto.setId(t.getFId());
            tableScanDto.setName(t.getFName());
            tableScanDto.setUpdateTime(t.getFOperationTime());
            return tableScanDto;
        }).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    @Override
    public ResponseEntity<?> getTableAndFieldDetailBatch(String accountId, String accountType, TableIdsVo req) {
        JSONObject response = new JSONObject();
        List<String> tableIds = req.getTableIds();
        if (tableIds == null || tableIds.size() == 0) {
            response.put("entries", null);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }

        // 查询所有
        // TODO:这里可以根据userId过滤资源
        Set<String> ids = new HashSet<>(tableIds);
        HashMap<String, Boolean> tableExistMap = new HashMap<>(ids.size());
        List<TableScanEntity> tableScanEntities = tableScanMapper.selectPageBatchIds(ids);
        if (tableScanEntities == null || tableScanEntities.size() == 0) {
            response.put("entries", null);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        HashMap<String, TableScanEntity> mapTab = new HashMap<>(tableScanEntities.size());
        HashMap<String, DataSourceEntity> mapDs = new HashMap<>(tableScanEntities.size());
        HashMap<String, DataSourceEntity> mapDsTmp = new HashMap<>(tableScanEntities.size());
        HashMap<String, List<FieldScanEntity>> mapFields = new HashMap<>();

        Set<String> dsIds = tableScanEntities.stream()
                .map(TableScanEntity::getFDataSourceId)
                .collect(Collectors.toSet());
        // 查询所有数据源
        dataSourceMapper.selectBatchIds(dsIds).forEach(dataSourceEntity -> {
            mapDsTmp.put(dataSourceEntity.getFId(), dataSourceEntity);
        });
        tableScanEntities.forEach(t -> {
            mapTab.put(t.getFId(), t);
            // 处理数据源
            DataSourceEntity dataSourceEntity = mapDsTmp.get(t.getFDataSourceId());
            if (dataSourceEntity != null) {
                mapDs.put(t.getFId(), dataSourceEntity);
            }
        });
        // 查询fieldsList
        List<FieldScanEntity> fieldsList = fieldScanMapper.getAllFieldListByTableId(ids);
        fieldsList.forEach(field -> {
            String tableId = field.getFTableId();
            if (!mapFields.containsKey(tableId)) {
                mapFields.put(tableId, new ArrayList<>());
            }
            mapFields.get(tableId).add(field);
        });

        // 判断废弃表
        for (String tableId : ids) {
            boolean exist = true;
            if (!mapDs.containsKey(tableId)) {
                exist = false;
            } else if (!mapTab.containsKey(tableId)) {
                exist = false;
            } else if (!mapFields.containsKey(tableId)) {
                exist = false;
            }
            tableExistMap.put(tableId, exist);
        }
        List<TableAndFieldDto> results = new ArrayList<>(tableExistMap.size());

        for (String tableId : tableExistMap.keySet()) {
            TableAndFieldDto result = new TableAndFieldDto();
            result.setTableId(tableId);
            Boolean isExist = tableExistMap.get(tableId);
            if (isExist) {
                TableScanEntity tableScanEntity = mapTab.get(tableId);
                DataSourceEntity dataSourceEntity = mapDs.get(tableId);
                List<FieldScanEntity> fieldScanEntities = mapFields.get(tableId);
                TableAndFieldDto.TableDto table = new TableAndFieldDto.TableDto(tableScanEntity);
                result.setTable(table);
                TableAndFieldDto.DatasourceDto datasourceDto = new TableAndFieldDto.DatasourceDto(dataSourceEntity);
                result.setDatasource(datasourceDto);
                List<TableAndFieldDto.FieldDto> fields = new ArrayList<>(fieldScanEntities.size());
                for (FieldScanEntity field : fieldScanEntities) {
                    // t_table_field 没有 f_table_name字段，需要补全
                    String fTableName = field.getFTableName();
                    if (CommonUtil.isEmpty(fTableName)) {
                        TableOldEntity tableOldEntity = tableOldMapper.selectById(field.getFTableId());
                        field.setFTableName(tableOldEntity.getFName());
                    }
                    TableAndFieldDto.FieldDto fieldDto = new TableAndFieldDto.FieldDto(field);
                    fields.add(fieldDto);
                }
                result.setFieldList(fields);
            }
            results.add(result);
        }
        response.put("entries", results);
        response.put("total_count", tableIds.size());
        return ResponseEntity.ok(response);
    }


    public ResponseEntity<?> getTableOldListByDsId(String userId, String dsId, String keyword, int limit,
                                                   int offset, String sort, String direction) {
        JSONArray entries = new JSONArray();
        JSONObject response = new JSONObject();
        List<TableOldEntity> dsList = tableOldMapper.getTableListByDsId(dsId, keyword);
        long count = tableOldMapper.selectCount(dsId, keyword);
        if (dsList == null || dsList.size() == 0) {
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        //TODO:这里可以根据userId过滤资源
        Set<Long> ids = dsList.stream().map(TableOldEntity::getFId).collect(Collectors.toSet());
        if ("f_create_time".equals(sort) || "f_operation_time".equals(sort)) {
            sort = "f_update_time";
        }
        List<TableOldEntity> tableScanEntities = tableOldMapper.selectPage(ids, keyword, offset, limit, sort, direction);
        if (tableScanEntities == null || tableScanEntities.size() == 0) {
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        List<TableScanDto> results = tableScanEntities.stream().map(t -> {
            TableScanDto tableScanDto = new TableScanDto(
                    String.valueOf(t.getFId()),
                    t.getFName(),
                    t.getFCreateTime(),
                    t.getFUpdateTime(),
                    null
            );
            String fAdvancedParams = t.getFAdvancedParams();
            String fDataSourceName = t.getFDataSourceName();
            if (CommonUtil.isNotEmpty(fAdvancedParams)) {
                setAdvancedParams(t.getFDataSourceTypeName(),
                        fDataSourceName,
                        fAdvancedParams,
                        t.getFScanSource(),
                        t.getFDescription(),
                        tableScanDto
                );
            }
            return tableScanDto;
        }).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    private void setAdvancedParams(String dataSourceType,
                                   String fDataSourceName,
                                   String fAdvancedParams,
                                   Integer fScanSource,
                                   String comment,
                                   TableScanDto tableScanDto) {
        if (CommonUtil.EXCEL.equals(dataSourceType)) {
            HashMap[] array = JSON.parseObject(fAdvancedParams, HashMap[].class);
            String sheet = "";
            Boolean allSheet = false;
            Boolean sheetAsNewColumn = false;
            Boolean hasHeaders = false;
            String startCell = "";
            String endCell = "";
            String fileName = "";
            for (HashMap map : array) {
                String key = (String) map.get("key");
                if ("sheet".equals(key)) {
                    sheet = (String) map.getOrDefault("value", "");
                } else if ("allSheet".equals(key)) {
                    allSheet = (Boolean) map.getOrDefault("value", false);
                } else if ("sheetAsNewColumn".equals(key)) {
                    sheetAsNewColumn = (Boolean) map.getOrDefault("value", false);
                } else if ("hasHeaders".equals(key)) {
                    hasHeaders = (Boolean) map.getOrDefault("value", false);
                } else if ("startCell".equals(key)) {
                    startCell = (String) map.getOrDefault("value", "");
                } else if ("endCell".equals(key)) {
                    endCell = (String) map.getOrDefault("value", "");
                } else if ("fileName".equals(key)) {
                    fileName = (String) map.getOrDefault("value", "");
                }
            }
            TableScanDto.AdvancedParam param = new TableScanDto.AdvancedParam(sheet,
                    allSheet,
                    sheetAsNewColumn,
                    startCell,
                    endCell,
                    hasHeaders,
                    fileName,
                    fDataSourceName);
            tableScanDto.setAdvancedParams(param);
        } else if (CommonUtil.MYSQL.equals(dataSourceType) || CommonUtil.MARIA.equals(dataSourceType)) {
            HashMap[] array = JSON.parseObject(fAdvancedParams, HashMap[].class);
            String updateTime = "";
            String createTime = "";
            String engine = "";
            String tableRows = null;
            String dataLength = null;
            String indexLength = null;
            for (HashMap map : array) {
                String key = (String) map.get("key");
                if ("update_time".equals(key)) {
                    updateTime = (String) map.getOrDefault("value", "");
                } else if ("create_time".equals(key)) {
                    createTime = (String) map.getOrDefault("value", "");
                } else if ("engine".equals(key)) {
                    engine = (String) map.getOrDefault("value", "");
                } else if ("table_rows".equals(key)) {
                    tableRows = (String) map.getOrDefault("value", null);
                } else if ("data_length".equals(key)) {
                    dataLength = (String) map.getOrDefault("value", null);
                } else if ("index_length".equals(key)) {
                    indexLength = (String) map.getOrDefault("value", null);
                }
            }
            String tableType = "";
            if (fScanSource != null) {
                if (fScanSource == 1) {
                    tableType = "视图";
                } else if (fScanSource == 0) {
                    tableType = "普通表";
                }
            }
            dataLength = CommonUtil.isNotEmpty(dataLength) ? dataLength + "MB" : null;
            indexLength = CommonUtil.isNotEmpty(indexLength) ? indexLength + "MB" : null;
            JSONObject param = new JSONObject();
            param.put("table_type", tableType);
            param.put("init_time", createTime);
            param.put("upgrade_time", updateTime);
            param.put("engine", engine);
            param.put("table_rows", tableRows);
            param.put("data_length", dataLength);
            param.put("index_length", indexLength);
            param.put("comment", comment);
            tableScanDto.setAdvancedParams(param);
        } else if (CommonUtil.OPEN_SEARCH.equals(dataSourceType)) {
            HashMap[] array = JSON.parseObject(fAdvancedParams, HashMap[].class);
            String health = "";
            String status = "";
            String uuid = "";
            String pri = "";
            String rep = "";
            String docsCount = "";
            String docsDeleted = "";
            String storeSize = "";
            String priStoreSize = "";
            for (HashMap map : array) {
                String key = (String) map.get("key");
                if ("health".equals(key)) {
                    health = (String) map.getOrDefault("value", "");
                } else if ("status".equals(key)) {
                    status = (String) map.getOrDefault("value", "");
                } else if ("uuid".equals(key)) {
                    uuid = (String) map.getOrDefault("value", "");
                } else if ("pri".equals(key)) {
                    pri = (String) map.getOrDefault("value", "");
                } else if ("rep".equals(key)) {
                    rep = (String) map.getOrDefault("value", "");
                } else if ("docs.count".equals(key)) {
                    docsCount = (String) map.getOrDefault("value", "");
                } else if ("docs.deleted".equals(key)) {
                    docsDeleted = (String) map.getOrDefault("value", "");
                } else if ("store.size".equals(key)) {
                    storeSize = (String) map.getOrDefault("value", "");
                } else if ("pri.store.size".equals(key)) {
                    priStoreSize = (String) map.getOrDefault("value", "");
                }
            }
            JSONObject param = new JSONObject();
            param.put("health", health);
            param.put("status", status);
            param.put("uuid", uuid);
            param.put("pri", pri);
            param.put("rep", rep);
            param.put("docs.count", docsCount);
            param.put("docs.deleted", docsDeleted);
            param.put("store.size", storeSize);
            param.put("pri.store.size", priStoreSize);
            tableScanDto.setAdvancedParams(param);
        } else {
            JSONObject param = new JSONObject();
            tableScanDto.setAdvancedParams(param);
        }
    }
}
