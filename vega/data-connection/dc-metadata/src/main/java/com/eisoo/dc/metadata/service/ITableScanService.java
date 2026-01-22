package com.eisoo.dc.metadata.service;

import com.baomidou.mybatisplus.extension.service.IService;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.metadata.domain.vo.DataSourceIdsVo;
import com.eisoo.dc.metadata.domain.vo.TableIdsVo;
import org.springframework.http.ResponseEntity;

import java.util.Collection;
import java.util.List;

/**
 * @author Tian.lan
 */
public interface ITableScanService extends IService<TableScanEntity> {
    List<TableScanEntity> selectByDsId(String dsId);

    boolean updateTableScanBatchById(Collection<TableScanEntity> entityList);

    boolean updateBatchTableScan(Collection<TableScanEntity> entityList, int batch);

    boolean saveBatchTableScan(Collection<TableScanEntity> entityList, int batch);

    int updateScanStatusById(String id, int status);
    int updateScanStatusByIdNewRequires(String id, String taskId, int status);
    int updateScanStatusAndOperationTimeById(String tableId, String taskId, int status);


    void tableScanBatch(List<TableScanEntity> deletes, List<TableScanEntity> updateList, List<TableScanEntity> saveList);

    ResponseEntity<?> getTableListByDsId(String userId, String dsId, String keyword, int limit, int offset, String sort, String direction);

    ResponseEntity<?> getTableListByDsIdsBatch(String userId, String accountType, DataSourceIdsVo req,String updateTime, String keyword, int limit, int offset, String sort, String direction);

    ResponseEntity<?> getTableAndFieldDetailBatch(String accountId, String accountType, TableIdsVo req);

    List<TableScanEntity> selectBatchIds(List<String> tables);
}
