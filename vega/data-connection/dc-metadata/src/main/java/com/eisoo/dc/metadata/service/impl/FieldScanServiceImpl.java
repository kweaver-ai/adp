package com.eisoo.dc.metadata.service.impl;

import cn.hutool.json.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.eisoo.dc.common.constant.Description;
import com.eisoo.dc.common.constant.Detail;
import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.FieldOldEntity;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableOldEntity;
import com.eisoo.dc.common.metadata.mapper.FieldOldMapper;
import com.eisoo.dc.common.metadata.mapper.FieldScanMapper;
import com.eisoo.dc.common.metadata.mapper.TableOldMapper;
import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.metadata.domain.dto.FieldScanDto;
import com.eisoo.dc.metadata.service.IFieldScanService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * @author Tian.lan
 */
@Service
@Slf4j
public class FieldScanServiceImpl extends ServiceImpl<FieldScanMapper, FieldScanEntity> implements IFieldScanService {
    @Autowired(required = false)
    private FieldScanMapper fieldScanMapper;
    @Autowired(required = false)
    private FieldOldMapper fieldOldMapper;
    @Autowired(required = false)
    private TableOldMapper tableOldMapper;

    @Override
    public List<FieldScanEntity> selectByTableId(String tableId) {
        return fieldScanMapper.selectByTableId(tableId);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public void fieldScanBatch(List<FieldScanEntity> deletes, List<FieldScanEntity> updateList, List<FieldScanEntity> saveList) {
        if (deletes != null && !deletes.isEmpty()) {
            this.updateBatchById(deletes);
        }
        if (!updateList.isEmpty()) {
            updateBatchById(updateList, 100);
        }
        if (!saveList.isEmpty()) {
            this.saveBatch(saveList, 1000);
        }
    }

    @Override
    public ResponseEntity<?> getFieldListByTableId(String userId, String tableId, String keyword, int limit, int offset, String sort, String direction) {
        JSONObject response = new JSONObject();
        List<FieldScanEntity> dsList = fieldScanMapper.getFieldListByTableId(tableId, keyword);
        long count = fieldScanMapper.selectCount(tableId, keyword);
        if (dsList == null || dsList.size() == 0) {
            // 从old查询
            return getFieldOldListByTableId(userId, tableId, keyword, limit, offset, sort, direction);
        }
        //TODO:这里可以根据userId过滤资源
        Set<String> ids = dsList.stream().map(t -> t.getFId()).collect(Collectors.toSet());
        List<FieldScanEntity> tableScanEntities = fieldScanMapper.selectPage(ids, keyword, offset, limit, sort, direction);
        if (tableScanEntities == null || tableScanEntities.size() == 0) {
            response.put("entries", new JSONArray());
            response.put("total_count", count);
            return ResponseEntity.ok(response);
        }
        List<FieldScanDto> results = tableScanEntities.stream().map(t -> new FieldScanDto(tableId,
                t.getFId(),
                t.getFFieldName(),
                t.getFFieldType().toLowerCase(),
                CommonUtil.getVirtualType(t.getFAdvancedParams()),
                t.getFFieldComment()
        )).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }

    public ResponseEntity<?> getFieldOldListByTableId(String userId, String tableId, String keyword, int limit, int offset, String sort, String direction) {
        JSONArray entries = new JSONArray();
        JSONObject response = new JSONObject();
        Long tableIdNew = null;
        try {
            tableIdNew = Long.valueOf(tableId);
        } catch (Exception e) {
            throw new AiShuException(ErrorCodeEnum.BadRequest,
                    Description.TABLE_NOT_FOUND_ERROR,
                    String.format(Detail.TABLE_NOT_EXIST, tableId),
                    Message.MESSAGE_INTERNAL_ERROR);
        }
        TableOldEntity tableOldEntity = tableOldMapper.selectById(tableIdNew);
        if (tableOldEntity == null) {
            // 不存在返回空就行
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        List<FieldOldEntity> dsList = fieldOldMapper.getFieldListByTableId(tableIdNew, keyword);
        long count = fieldOldMapper.selectCount(tableIdNew, keyword);
        if (dsList == null || dsList.size() == 0) {
            // 从old查询
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        //TODO:这里可以根据userId过滤资源
        Set<String> names = dsList.stream().map(FieldOldEntity::getFFieldName).collect(Collectors.toSet());
        if ("f_create_time".equals(sort)) {
            sort = "f_update_time";
        }
        List<FieldOldEntity> tableScanEntities = fieldOldMapper.selectPage(tableIdNew, names, keyword, offset, limit, sort, direction);
        if (tableScanEntities == null || tableScanEntities.size() == 0) {
            // 从old查询
            response.put("entries", entries);
            response.put("total_count", 0);
            return ResponseEntity.ok(response);
        }
        List<FieldScanDto> results = tableScanEntities.stream().map(t -> new FieldScanDto(tableId,
                null,
                t.getFFieldName(),
                t.getFFieldType().toLowerCase(),
                CommonUtil.getVirtualType(t.getFAdvancedParams()),
                t.getFFieldComment())).collect(Collectors.toList());
        response.put("entries", results);
        response.put("total_count", count);
        return ResponseEntity.ok(response);
    }
}
