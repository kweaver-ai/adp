package com.eisoo.dc.datasource.controller;

import com.eisoo.dc.common.util.CommonUtil;
import com.eisoo.dc.common.util.StringUtils;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.eisoo.dc.datasource.domain.vo.DataSourceVo;
import com.eisoo.dc.common.vo.TestDataSourceVo;
import com.eisoo.dc.datasource.service.CatalogService;
import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.validation.constraints.Min;
import javax.validation.constraints.Pattern;
import javax.validation.constraints.Size;

@Api(tags = "数据源管理")
@RestController
@Validated
@RequestMapping("/api")
public class CatalogController {

    @Autowired(required = false)
    private CatalogService catalogService;


    @ApiOperation(value = "新增数据源", notes = "新增数据源接口")
    @PostMapping("/data-connection/v1/datasource")
    public ResponseEntity<?> createDatasource(HttpServletRequest request, @Validated @RequestBody DataSourceVo req){
        return catalogService.createDatasource(request,req);
    }

    @ApiOperation(value = "查询数据源列表", notes = "查询数据源列表接口")
    @GetMapping("/data-connection/v1/datasource")
    public ResponseEntity<?> getDatasourceList(HttpServletRequest request,
                                          @RequestParam(value = "limit", required = false, defaultValue = "-1") @Min(value = -1) int limit,
                                          @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                          @RequestParam(value = "keyword", required = false) String keyword,
                                          @RequestParam(value = "type", required = false)
                                             @Pattern(regexp = "^(|structured|no-structured|other)(,(structured|no-structured|other))*$", message = "可选参数值：structured、no-structured、other，多个由逗号分割")
                                             String types,
                                          @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                             @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc")
                                             String direction,
                                          @RequestParam(value = "sort", required = false, defaultValue = "created_at")
                                             @Pattern(regexp = "updated_at|created_at|name", message = "可选参数值：updated_at、created_at、name")
                                             String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String userType = introspectInfo.getAccountType();
        return catalogService.getDatasourceList(userId,userType,keyword,types,limit,offset,sort,direction);
    }

    @ApiOperation(value = "查询数据源列表（内部）", notes = "查询数据源列表（内部）")
    @GetMapping("/internal/data-connection/v1/datasource")
    public ResponseEntity<?> getDatasourceListByInternal(@RequestHeader(name = "x-account-id") String accountId,
                                                         @RequestHeader(name = "x-account-type")
                                                         @Pattern(regexp = "user|app|anonymous", message = "可选参数值：user、app、anonymous")
                                                         String accountType,
                                               @RequestParam(value = "limit", required = false, defaultValue = "-1") @Min(value = -1) int limit,
                                               @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                               @RequestParam(value = "keyword", required = false) String keyword,
                                               @RequestParam(value = "type", required = false)
                                                 @Pattern(regexp = "^(|structured|no-structured|other)(,(structured|no-structured|other))*$", message = "可选参数值：structured、no-structured、other，多个由逗号分割")
                                                 String types,
                                               @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                               @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc")
                                               String direction,
                                               @RequestParam(value = "sort", required = false, defaultValue = "created_at")
                                               @Pattern(regexp = "updated_at|created_at|name", message = "可选参数值：updated_at、created_at、name")
                                               String sort) {
        return catalogService.getDatasourceList(accountId,accountType,keyword,types,limit,offset,sort,direction);
    }

    @ApiOperation(value = "获取数据源列表", notes = "获取数据源列表接口")
    @GetMapping("/data-connection/v1/datasource/assignable-catalog")
    public ResponseEntity<?> getAssignableDatasourceList(HttpServletRequest request,
                                               @RequestParam(value = "id", required = false) String id,
                                               @RequestParam(value = "limit", required = false, defaultValue = "50") @Min(value = -1) int limit,
                                               @RequestParam(value = "offset", required = false, defaultValue = "0") @Min(value = 0) int offset,
                                               @RequestParam(value = "keyword", required = false) String keyword,
                                               @RequestParam(value = "direction", required = false, defaultValue = "desc")
                                               @Pattern(regexp = "asc|desc", flags = Pattern.Flag.CASE_INSENSITIVE, message = "可选参数值：asc、desc")
                                               String direction,
                                               @RequestParam(value = "sort", required = false, defaultValue = "created_at")
                                               @Pattern(regexp = "updated_at|created_at|name", message = "可选参数值：updated_at、created_at、name")
                                               String sort) {
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String userType = introspectInfo.getAccountType();
        return catalogService.getAssignableDatasourceList(userId,userType,id,keyword,limit,offset,sort,direction);
    }

    @ApiOperation(value = "查询数据源详情", notes = "根据id查询数据源详情")
    @GetMapping("/data-connection/v1/datasource/{id}")
    public ResponseEntity<?> getDatasource(HttpServletRequest request, @PathVariable("id") @Size(max = 36) String id){
        IntrospectInfo introspectInfo = CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = StringUtils.defaultString(introspectInfo.getSub());
        String userType = introspectInfo.getAccountType();
        return catalogService.getDatasource(userId,userType, id);
    }

    @ApiOperation(value = "查询数据源详情（内部）", notes = "根据id查询数据源详情（内部）")
    @GetMapping("/internal/data-connection/v1/datasource/{id}")
    public ResponseEntity<?> getDatasourceByInternal(@RequestHeader(name = "x-account-id") String accountId,
                                                     @RequestHeader(name = "x-account-type")
                                                     @Pattern(regexp = "user|app|anonymous", message = "可选参数值：user、app、anonymous")
                                                     String accountType,
                                                     @PathVariable("id") @Size(max = 36) String id){
        return catalogService.getDatasource(accountId,accountType, id);
    }

    @ApiOperation(value = "更新数据源", notes = "更新数据源接口")
    @PutMapping("/data-connection/v1/datasource/{id}")
    public ResponseEntity<?> updateDatasource(HttpServletRequest request, @Validated @RequestBody DataSourceVo req, @PathVariable("id") @Size(max = 36) String id){
        return catalogService.updateDatasource(request, req, id);
    }

    @ApiOperation(value = "删除数据源", notes = "删除数据源接口")
    @DeleteMapping("/data-connection/v1/datasource/{id}")
    public ResponseEntity<?> deleteDatasource(HttpServletRequest request, @PathVariable("id") @Size(max = 36) String id){
        return catalogService.deleteDatasource(request, id);
    }

    @ApiOperation(value = "测试数据源连接", notes = "测试数据源连接接口")
    @PostMapping("/data-connection/v1/datasource/test")
    public ResponseEntity<?> testDataSource(HttpServletRequest request, @Validated @RequestBody TestDataSourceVo dto){
        return catalogService.testDataSource(request, dto);
    }

    @ApiOperation(value = "查询所有支持数据源", notes = "查询所有支持数据源接口")
    @GetMapping("/data-connection/v1/datasource/connectors")
    public ResponseEntity<?> connectorList(@RequestParam(value = "type", required = false)
                                           @Pattern(regexp = "structured|no-structured|other", message = "可选参数值：structured、no-structured、other") String type) {
        return catalogService.connectorList(type);
    }


}
