package com.eisoo.dc.datasource.service;

import com.eisoo.dc.datasource.domain.vo.DataSourceVo;
import com.eisoo.dc.common.vo.TestDataSourceVo;
import org.springframework.http.ResponseEntity;

import javax.servlet.http.HttpServletRequest;


public interface CatalogService {

    ResponseEntity<?> createDatasource(HttpServletRequest request, DataSourceVo params);

    ResponseEntity<?> getDatasourceList(String userId, String userType, String keyword,String types,int limit,int offset,String sort,String direction);

    ResponseEntity<?> getAssignableDatasourceList(String userId, String userType, String id, String keyword,int limit,int offset,String sort,String direction);

    ResponseEntity<?> getDatasource(String userId, String userType, String id);

    ResponseEntity<?> updateDatasource(HttpServletRequest request, DataSourceVo params,String id);

    ResponseEntity<?> deleteDatasource(HttpServletRequest request, String id);

    ResponseEntity<?> testDataSource(HttpServletRequest request, TestDataSourceVo catalogDto);
    ResponseEntity<?> connectorList(String type);
}
