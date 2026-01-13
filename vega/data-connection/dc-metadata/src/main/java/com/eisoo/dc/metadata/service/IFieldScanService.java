package com.eisoo.dc.metadata.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import org.springframework.http.ResponseEntity;

import java.util.List;

/**
 * @author Tian.lan
 */
public interface IFieldScanService extends IService<FieldScanEntity> {
    List<FieldScanEntity> selectByTableId(String tableId);
    void fieldScanBatch(List<FieldScanEntity> deletes, List<FieldScanEntity> updateList, List<FieldScanEntity> saveList);
    ResponseEntity<?> getFieldListByTableId(String userId, String tableId, String keyword, int limit, int offset, String sort, String direction);
}
