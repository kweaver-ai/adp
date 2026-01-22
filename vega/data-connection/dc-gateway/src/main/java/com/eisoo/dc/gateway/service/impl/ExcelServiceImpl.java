package com.eisoo.dc.gateway.service.impl;

import com.alibaba.fastjson2.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.baomidou.mybatisplus.core.conditions.query.QueryWrapper;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.connector.ConnectorConfigCache;
import com.eisoo.dc.common.connector.TypeConfig;
import com.eisoo.dc.common.enums.ScanStatusEnum;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.TableScanEntity;
import com.eisoo.dc.common.metadata.mapper.FieldScanMapper;
import com.eisoo.dc.common.metadata.mapper.TableScanMapper;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.eisoo.dc.gateway.common.CatalogConstant;
import com.eisoo.dc.gateway.common.Detail;
import com.eisoo.dc.gateway.common.Message;
import com.eisoo.dc.gateway.domain.dto.ExcelTableConfigDto;
import com.eisoo.dc.gateway.service.ExcelService;
import com.eisoo.dc.gateway.service.GatewayCatalogService;
import com.eisoo.dc.gateway.service.GatewayViewService;
import com.eisoo.dc.gateway.util.*;
import com.monitorjbl.xlsx.StreamingReader;
import com.monitorjbl.xlsx.exceptions.MissingSheetException;
import org.apache.commons.lang3.ArrayUtils;
import org.apache.commons.lang3.StringUtils;
import org.apache.poi.ss.usermodel.Cell;
import org.apache.poi.ss.usermodel.Row;
import org.apache.poi.ss.usermodel.Sheet;
import org.apache.poi.ss.usermodel.Workbook;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import javax.servlet.http.HttpServletRequest;
import java.io.InputStream;
import java.net.URLDecoder;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import static org.apache.poi.ss.usermodel.CellType.STRING;

/**
 * @Author exx
 **/
@Service
public class ExcelServiceImpl implements ExcelService {

    private static final Logger log = LoggerFactory.getLogger(ExcelServiceImpl.class);

    private static final String SCHEMA_DEFAULT = "default";
    private static final String CELL_PATTERN_REG = "^[A-Z]+[1-9][0-9]*$";
    private static final Pattern CELL_PATTERN = Pattern.compile(CELL_PATTERN_REG);
    private static final String[] COLUMN_TYPE = {"bigint", "varchar", "timestamp", "boolean", "double"};

    @Autowired(required = false)
    GatewayCatalogService gatewayCatalogService;

    @Autowired(required = false)
    GatewayViewService gatewayViewService;

    @Autowired(required = false)
    GatewayASUtil gatewayAsUtil;

    @Autowired(required = false)
    ExcelUtil excelUtil;

    @Autowired(required = false)
    TableScanMapper tableScanMapper;

    @Autowired(required = false)
    FieldScanMapper fieldScanMapper;

    @Autowired
    ConnectorConfigCache connectorConfigCache;

    private static Map<String, String> sourceTypeVegaTypeMapping;

    @Value(value = "${services.efast-public}")
    private String efastPublic;

    @Override
    public ResponseEntity<?> files(HttpServletRequest request, String catalog) {
        JSONObject files = getFiles(TokenUtil.getBearerToken(request), catalog);
        JSONObject filesJSONObject = files.getJSONObject("data");
        JSONArray data = new JSONArray();
        data.addAll(filesJSONObject.keySet());
        JSONObject result = new JSONObject();
        result.put("data", data);
        result.put("total", files.getIntValue("total"));
        return ResponseEntity.ok(result);
    }

    public JSONObject getFiles(String token, String catalog) {
        String catalogStr = CheckUtil.checkCatalog(gatewayCatalogService, catalog);
        JSONObject catalogJson = JSONObject.parseObject(catalogStr);
        if (!catalogJson.getString("connector.name").equals(CatalogConstant.EXCEL_CATALOG)
                && !catalogJson.containsKey("excel.password")) {
            log.error("数据源类型错误:{}", catalogJson.getString("connector.name"));
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, String.format(Detail.CATALOG_TYPE_ERROR, catalogJson.getString("connector.name")));
        }
        String base = catalogJson.getString("excel.base");
        String protocol = catalogJson.getString("excel.protocol");
        String url = "";
        String path = "";
        if (protocol.equals(CatalogConstant.STORAGE_PROTOCOL_DOCLIB)) {
            url = efastPublic;
            path = GatewayASUtil.getItemPath(url, token, base);
        } else {
            String host = catalogJson.getString("excel.host");
            String port = catalogJson.getString("excel.port");
            String username = catalogJson.getString("excel.username");
            String password = catalogJson.getString("excel.password");
            url = CommonUtil.getUrl("https", host, port);
            token = GatewayASUtil.getToken(host, port, username, password);
            path = base;
        }

        String catalogFileName = null;
        if (excelUtil.isExcelFile(path)) {
            catalogFileName = path.substring(path.lastIndexOf("/") + 1);
            base = base.substring(0, base.lastIndexOf("/"));
        }
        String docId = "";
        if (protocol.equals(CatalogConstant.STORAGE_PROTOCOL_DOCLIB)) {
            docId = base;
        } else {
            docId = GatewayASUtil.getDocid(url, token, base);
        }
        JSONObject dirJson = GatewayASUtil.loadDir(url, token, docId);
        JSONArray files = dirJson.getJSONArray("files");

        JSONObject result = new JSONObject();
        JSONObject data = new JSONObject();
        int total = 0;
        for (int i = 0; i < files.size(); i++) {
            JSONObject file = files.getJSONObject(i);
            String docid = file.getString("docid");
            String fileName = file.getString("name");
            if (excelUtil.isExcelFile(fileName) && (StringUtils.isEmpty(catalogFileName) || fileName.equals(catalogFileName))) {
                data.put(fileName, docid);
                total++;
            }
        }
        result.put("data", data);
        result.put("total", total);
        return result;
    }

    @Override
    public ResponseEntity<?> sheet(HttpServletRequest request, String catalog, String fileName) {
        fileName = decodeFileName(fileName);
        String token = TokenUtil.getBearerToken(request);
        String docid = checkFileName(token, catalog, fileName);
        String catalogStr = CheckUtil.checkCatalog(gatewayCatalogService, catalog);
        JSONObject catalogJson = JSONObject.parseObject(catalogStr);
        String protocol = catalogJson.getString("excel.protocol");
        String url = "";
        if (protocol.equals(CatalogConstant.STORAGE_PROTOCOL_DOCLIB)) {
            url = efastPublic;
        } else {
            String host = catalogJson.getString("excel.host");
            String port = catalogJson.getString("excel.port");
            String username = catalogJson.getString("excel.username");
            String password = catalogJson.getString("excel.password");
            url = CommonUtil.getUrl("https", host, port);
            token = GatewayASUtil.getToken(host, port, username, password);
        }

        JSONObject result = new JSONObject();
        JSONArray data = new JSONArray();
        result.put("data", data);

        try {
            InputStream inputStream = GatewayASUtil.getInputStream(url, token, docid);
            Workbook workbook = StreamingReader.builder()
                    .bufferSize(8192)
                    .rowCacheSize(100) // 限制缓存行数
                    .open(inputStream);
            Iterator<Sheet> it = workbook.sheetIterator();
            while (it.hasNext()) {
                data.add(it.next().getSheetName());
            }
            result.put("total", workbook.getNumberOfSheets());
        } catch (AiShuException aiShuException) {
            throw aiShuException;
        } catch (Exception e) {
            log.error("excel sheet error", e);
            throw new AiShuException(ErrorCodeEnum.ReadExcelFail, String.format(Detail.READ_FILE_ERROR, fileName), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        return ResponseEntity.ok(result);
    }

    @Override
    public ResponseEntity<?> columns(HttpServletRequest request, ExcelTableConfigDto excelTableConfigDto) {
        String token = TokenUtil.getBearerToken(request);
        String docid = checkFileName(token, excelTableConfigDto.getCatalog(), excelTableConfigDto.getFileName());
        String catalogStr = CheckUtil.checkCatalog(gatewayCatalogService, excelTableConfigDto.getCatalog());
        JSONObject catalogJson = JSONObject.parseObject(catalogStr);
        String protocol = catalogJson.getString("excel.protocol");
        String url = "";
        if (protocol.equals(CatalogConstant.STORAGE_PROTOCOL_DOCLIB)) {
            url = efastPublic;
        } else {
            String host = catalogJson.getString("excel.host");
            String port = catalogJson.getString("excel.port");
            String username = catalogJson.getString("excel.username");
            String password = catalogJson.getString("excel.password");
            url = CommonUtil.getUrl("https", host, port);
            token = GatewayASUtil.getToken(host, port, username, password);
        }

        JSONObject result = new JSONObject();
        JSONArray data = new JSONArray();
        result.put("data", data);

        try {
            InputStream inputStream = GatewayASUtil.getInputStream(url, token, docid);
            Workbook workbook = StreamingReader.builder()
                    .bufferSize(8192)
                    .rowCacheSize(100)
                    .open(inputStream);
            Sheet sheet;
            if (StringUtils.isEmpty(excelTableConfigDto.getSheet()) || excelTableConfigDto.isAllSheet()) {
                sheet = workbook.getSheetAt(0);
            } else {
                String sheetName = excelTableConfigDto.getSheet().split(",")[0];
                sheet = workbook.getSheet(sheetName);
                if (sheet == null) {
                    throw new AiShuException(ErrorCodeEnum.SheetNotExist, String.format(Detail.READ_SHEET_ERROR, sheetName), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
                }
            }

            int[][] rowCells = checkCellRange(excelTableConfigDto);
            int[] startRowCol = rowCells[0];
            int[] endRowCol = rowCells[1];

            boolean sheetNameExist = false;

            // 使用迭代方式安全获取行数据，避免 StreamingReader 的 getRow() 兼容性问题
            Row headerRow = null;
            Row dataRow = null;
            int targetHeaderRowNum = startRowCol[0] - 1;
            int targetDataRowNum = excelTableConfigDto.isHasHeaders() ? startRowCol[0] : targetHeaderRowNum;
            int maxTargetRowNum = Math.max(targetHeaderRowNum, targetDataRowNum);

            Iterator<Row> rowIterator = sheet.rowIterator();
            while (rowIterator.hasNext()) {
                Row currentRow = rowIterator.next();
                int currentRowNum = currentRow.getRowNum();

                if (excelTableConfigDto.isHasHeaders() && currentRowNum == targetHeaderRowNum) {
                    headerRow = currentRow;
                }
                if (currentRowNum == targetDataRowNum) {
                    dataRow = currentRow;
                }

                // 提前终止条件：如果已找到所有需要的行或者超过了最大目标行号
                if ((excelTableConfigDto.isHasHeaders() ? (headerRow != null && dataRow != null) : dataRow != null)
                        || currentRowNum > maxTargetRowNum) {
                    break;
                }
            }

            int number = 1;
            if (headerRow != null && excelTableConfigDto.isHasHeaders()) {
                for (int i = startRowCol[1] - 1; i < endRowCol[1]; i++) {
                    Cell cell = headerRow.getCell(i);
                    JSONObject columnJson = new JSONObject();
                    if (cell == null || cell.getCellType() != STRING) {
                        columnJson.put("column", "column_" + number++);
                    } else {
                        columnJson.put("column", cell.getStringCellValue());
                        if (Objects.equals(cell.getStringCellValue(), "sheet_name")) {
                            sheetNameExist = true;
                        }
                    }
                    Cell dataCell = dataRow != null ? dataRow.getCell(i) : null;
                    columnJson.put("type", excelUtil.guessColumnType(dataCell, workbook));
                    data.add(columnJson);
                }
            } else {
                for (int i = startRowCol[1] - 1; i < endRowCol[1]; i++) {
                    JSONObject columnJson = new JSONObject();
                    columnJson.put("column", "column_" + number++);
                    Cell dataCell = dataRow != null ? dataRow.getCell(i) : null;
                    columnJson.put("type", excelUtil.guessColumnType(dataCell, workbook));
                    data.add(columnJson);
                }
            }

            if (excelTableConfigDto.isSheetAsNewColumn()) {
                String sheetName = "sheet_name";
                if (sheetNameExist) {
                    sheetName += "_1";
                }
                JSONObject columnJson = new JSONObject();
                columnJson.put("column", sheetName);
                columnJson.put("type", "varchar");
                data.add(columnJson);
            }

            result.put("total", data.size());
        } catch (AiShuException aiShuException) {
            throw aiShuException;
        } catch (MissingSheetException e) {
            throw new AiShuException(ErrorCodeEnum.SheetNotExist, String.format(Detail.READ_SHEET_ERROR, e.getMessage()), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        } catch (Exception e) {
            log.error("excel sheet log column --> filename{}", excelTableConfigDto.getFileName(), e);
            throw new AiShuException(ErrorCodeEnum.ReadExcelFail, String.format(Detail.READ_FILE_ERROR, excelTableConfigDto.getFileName()), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }
        return ResponseEntity.ok(result);
    }

    @Override
    public ResponseEntity<?> createTable(HttpServletRequest request, ExcelTableConfigDto excelTableConfigDto) {
        IntrospectInfo introspectInfo = com.eisoo.dc.common.util.CommonUtil.getOrCreateIntrospectInfo(request);
        String userId = com.eisoo.dc.common.util.StringUtils.defaultString(introspectInfo.getSub());

        // 去掉参数首尾空格
        CheckUtil.excelTableConfigDtoTrim(excelTableConfigDto);
        if (StringUtils.isEmpty(excelTableConfigDto.getTableName())) {
            log.error("tableName不能为空");
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NAME_NOT_NULL);
        }
        QueryWrapper<TableScanEntity> wrapper = new QueryWrapper<>();
        wrapper.eq("f_data_source_id", excelTableConfigDto.getDataSourceId())
                .eq("f_name", excelTableConfigDto.getTableName())
                .ne("f_operation_type", 1);
        List<TableScanEntity> tableScanEntityList = tableScanMapper.selectList(wrapper);
        if (tableScanEntityList != null && tableScanEntityList.size() > 0) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NAME_EXIST);
        }
        checkTableConfig(request, excelTableConfigDto);
        initTypeMapping();
        String tableId = createTable(userId, excelTableConfigDto);

        String tableName = excelTableConfigDto.getCatalog() + "." + SCHEMA_DEFAULT + "." + excelTableConfigDto.getTableName();
        JSONObject obj = new JSONObject();
        obj.put("tableId", tableId);
        obj.put("tableName", tableName);
        return ResponseEntity.ok(obj);
    }

    @Override
    @Transactional(rollbackFor = Exception.class)
    public ResponseEntity<?> deleteTable(String tableId) {
        QueryWrapper<TableScanEntity> wrapper = new QueryWrapper<>();
        wrapper.eq("f_id", tableId)
                .ne("f_operation_type", 1);
        TableScanEntity tableScanEntity = tableScanMapper.selectById(tableId);
        if (tableScanEntity == null) {
            throw new AiShuException(ErrorCodeEnum.TableNotExist, tableId, Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
        }

        Date now = new Date();
        tableScanMapper.deleteById(tableId, now);
        fieldScanMapper.deleteByTableId(tableId, now);

        String fullName = tableScanEntity.getFDataSourceName() + "." +
                tableScanEntity.getFSchemaName() + "." +
                tableScanEntity.getFName();
        JSONObject result = new JSONObject();
        result.put("tableId", tableId);
        result.put("tableName", fullName);
        log.info("删除Excel表成功:{}", fullName);

        return ResponseEntity.ok(result);
    }

    public String decodeFileName(String fileName) {
        try {
            return URLDecoder.decode(fileName, "UTF-8");
        } catch (Exception e) {
            log.warn("decode fileName error: {}", fileName);
        }
        return fileName;
    }

    @Transactional(rollbackFor = Exception.class)
    public String createTable(String userId, ExcelTableConfigDto excelTableConfigDto) {
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        // 3. 格式化输出
        String now = LocalDateTime.now().format(formatter);
        TableScanEntity tableScanEntity = new TableScanEntity();
        tableScanEntity.setFId(UUID.randomUUID().toString());
        tableScanEntity.setFName(excelTableConfigDto.getTableName());
        tableScanEntity.setFSchemaName(SCHEMA_DEFAULT);
        tableScanEntity.setFTaskId("-1");
        tableScanEntity.setFDataSourceId(excelTableConfigDto.getDataSourceId());
        tableScanEntity.setFDataSourceName(excelTableConfigDto.getCatalog());
        tableScanEntity.setFVersion(1);
        tableScanEntity.setFCreateTime(now);
        tableScanEntity.setFCreatUser(userId);
        tableScanEntity.setFOperationTime(now);
        tableScanEntity.setFOperationUser(userId);
        tableScanEntity.setFOperationType(0);
        tableScanEntity.setFStatus(ScanStatusEnum.SUCCESS.getCode());
        tableScanEntity.setFStatusChange(1);
        tableScanEntity.setFAdvancedParams(getAdvancedParams(excelTableConfigDto));
        tableScanMapper.insert(tableScanEntity);

        for (int i = 0; i < excelTableConfigDto.getColumns().size(); i++) {
            String sourceType = excelTableConfigDto.getColumns().get(i).getType();
            String vegaType = sourceTypeVegaTypeMapping.get(sourceType);
            FieldScanEntity fieldScanEntity = new FieldScanEntity();
            fieldScanEntity.setFId(UUID.randomUUID().toString());
            fieldScanEntity.setFFieldName(excelTableConfigDto.getColumns().get(i).getColumn());
            fieldScanEntity.setFTableId(tableScanEntity.getFId());
            fieldScanEntity.setFTableName(excelTableConfigDto.getTableName());
            fieldScanEntity.setFFieldType(sourceType);
            fieldScanEntity.setFFieldOrderNo(Integer.toString(i));
            fieldScanEntity.setFAdvancedParams(getFieldAdvancedParams(sourceType, vegaType));
            fieldScanEntity.setFVersion(1);
            fieldScanEntity.setFCreatTime(now);
            fieldScanEntity.setFCreatUser(userId);
            fieldScanEntity.setFOperationTime(now);
            fieldScanEntity.setFOperationUser(userId);
            fieldScanEntity.setFOperationType(0);
            fieldScanEntity.setFOperationType(0);
            fieldScanMapper.insert(fieldScanEntity);
        }

        log.info("创建Excel表成功:" + tableScanEntity);
        return tableScanEntity.getFId();
    }

    public String getFieldAdvancedParams(String sourceType, String vegaType) {
        JSONArray array = new JSONArray();

        JSONObject originFieldTypeJson = new JSONObject();
        originFieldTypeJson.put("key", "originFieldType");
        originFieldTypeJson.put("value", sourceType);
        array.add(originFieldTypeJson);

        JSONObject virtualFieldTypeJson = new JSONObject();
        virtualFieldTypeJson.put("key", "virtualFieldType");
        virtualFieldTypeJson.put("value", vegaType);
        array.add(virtualFieldTypeJson);

        return array.toJSONString();
    }

    public void initTypeMapping() {
        if (sourceTypeVegaTypeMapping == null) {
            sourceTypeVegaTypeMapping = new HashMap<>();
            ConnectorConfig connectorConfig = connectorConfigCache.getConnectorConfig(CatalogConstant.EXCEL_CATALOG);
            for (TypeConfig typeConfig : connectorConfig.getType()) {
                sourceTypeVegaTypeMapping.put(typeConfig.getSourceType(), typeConfig.getVegaType());
            }
        }
    }

    public String checkFileName(String token, String catalog, String fileName) {
        if (!excelUtil.isExcelFile(fileName)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, String.format(Detail.FILE_TYPE_ERROR, fileName));
        }
        JSONObject filesResponseEntity = getFiles(token, catalog);
        JSONObject files = filesResponseEntity.getJSONObject("data");
        String docid = files.getString(fileName);
        if (StringUtils.isEmpty(docid)) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, String.format(Detail.FILE_NOT_EXIST, fileName));
        }
        return docid;
    }

    public int[][] checkCellRange(ExcelTableConfigDto excelTableConfigDto) {
        Matcher matcher = CELL_PATTERN.matcher(excelTableConfigDto.getStartCell());
        if (!matcher.matches()) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.START_CELL_ERROR);
        }

        matcher = CELL_PATTERN.matcher(excelTableConfigDto.getEndCell());
        if (!matcher.matches()) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.END_CELL_ERROR);
        }

        int[] startRowCol = excelTableConfigDto.getRowAndCloumnIndex(excelTableConfigDto.getStartCell());
        if (startRowCol[0] > 1048576 || startRowCol[1] > 16384) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.START_CELL_RANGE_ERROR, Message.MESSAGE_CELL_RANGE_SOLUTION);
        }
        int[] endRowCol = excelTableConfigDto.getRowAndCloumnIndex(excelTableConfigDto.getEndCell());
        if (endRowCol[0] > 1048576 || endRowCol[1] > 16384) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.END_CELL_RANGE_ERROR, Message.MESSAGE_CELL_RANGE_SOLUTION);
        }

        if (endRowCol[0] < startRowCol[0] || endRowCol[1] < startRowCol[1]) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CELL_RANGE_ERROR);
        }

        return new int[][]{startRowCol, endRowCol};
    }

    public void checkTableConfig(HttpServletRequest request, ExcelTableConfigDto excelTableConfigDto) {
        CheckUtil.checkCatalog(gatewayCatalogService, excelTableConfigDto.getCatalog());

        checkFileName(TokenUtil.getBearerToken(request), excelTableConfigDto.getCatalog(), excelTableConfigDto.getFileName());

        if (StringUtils.isNotEmpty(excelTableConfigDto.getTableName())
                && !CheckUtil.checkViewName(excelTableConfigDto.getTableName())) {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.TABLE_NAME_ERROR, Message.MESSAGE_VIEW_NAME_SOLUTION);
        }

        if (StringUtils.isNotEmpty(excelTableConfigDto.getSheet())) {
            String[] sheetParams = excelTableConfigDto.getSheet().split(",");
            ResponseEntity<?> sheetResponseEntity = sheet(request, excelTableConfigDto.getCatalog(), excelTableConfigDto.getFileName());
            JSONArray sheetArr = JSONObject.parseObject(sheetResponseEntity.getBody().toString()).getJSONArray("data");
            for (String sheetParam : sheetParams) {
                if (!sheetArr.stream().anyMatch(sheetParam::equals)) {
                    throw new AiShuException(ErrorCodeEnum.SheetNotExist, String.format(Detail.READ_SHEET_ERROR, excelTableConfigDto.getFileName()), Message.MESSAGE_DATANOTEXIST_ERROR_SOLUTION);
                }
            }
        }

        int[][] rowCells = checkCellRange(excelTableConfigDto);
        int[] startRowCol = rowCells[0];
        int[] endRowCol = rowCells[1];

        if (excelTableConfigDto.getColumns() != null && excelTableConfigDto.getColumns().size() > 0) {
            int colCount;
            colCount = endRowCol[1] - startRowCol[1] + 1;
            if (excelTableConfigDto.isSheetAsNewColumn()) {
                colCount += 1;
            }
            if (excelTableConfigDto.getColumns().size() != colCount) {
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.CELL_RANGE_AND_COLUMNS_INCONSISTENT);
            }

            Set<String> columnNameSet = new HashSet<>();
            for (ExcelTableConfigDto.ColumnType columnType : excelTableConfigDto.getColumns()) {
                if (StringUtils.isEmpty(columnType.getColumn())) {
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.COLUMN_NOT_NULL);
                } else {
                    if (!CheckUtil.checkExcelColumnName(columnType.getColumn())) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter, columnType.getColumn() + ":" + Detail.COLUMN_NAME_FORMAT_ERROR, Message.MESSAGE_COLUMN_NAME_SOLUTION);
                    }
                }

                if (columnNameSet.contains(columnType.getColumn().toLowerCase())) {
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter, String.format(Detail.COLUMN_NAME_ERROR, columnType.getColumn()));
                } else {
                    columnNameSet.add(columnType.getColumn().toLowerCase());
                }

                if (StringUtils.isEmpty(columnType.getType())) {
                    throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.COLUMN_TYPE_NOT_NULL);
                } else {
                    columnType.setType(columnType.getType().toLowerCase());
                    if (!ArrayUtils.contains(COLUMN_TYPE, columnType.getType())) {
                        throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.COLUMN_TYPE_UNSUPPORTED, Message.MESSAGE_COLUMN_TYPE_SOLUTION);
                    }
                }
            }
            columnNameSet.clear();
        } else {
            throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.COLUMN_TYPE_NOT_NULL);
        }

    }

    private String getAdvancedParams(ExcelTableConfigDto excelTableConfigDto) {
        JSONObject sheet = new JSONObject();
        sheet.put("key", "sheet");
        sheet.put("value", excelTableConfigDto.getSheet());

        JSONObject allSheet = new JSONObject();
        allSheet.put("key", "allSheet");
        allSheet.put("value", excelTableConfigDto.isAllSheet());

        JSONObject sheetAsNewColumn = new JSONObject();
        sheetAsNewColumn.put("key", "sheetAsNewColumn");
        sheetAsNewColumn.put("value", excelTableConfigDto.isSheetAsNewColumn());

        JSONObject startCell = new JSONObject();
        startCell.put("key", "startCell");
        startCell.put("value", excelTableConfigDto.getStartCell());

        JSONObject endCell = new JSONObject();
        endCell.put("key", "endCell");
        endCell.put("value", excelTableConfigDto.getEndCell());

        JSONObject hasHeaders = new JSONObject();
        hasHeaders.put("key", "hasHeaders");
        hasHeaders.put("value", excelTableConfigDto.isHasHeaders());

        JSONObject fileName = new JSONObject();
        fileName.put("key", "fileName");
        fileName.put("value", excelTableConfigDto.getFileName());

        return new JSONArray(
                sheet,
                allSheet,
                sheetAsNewColumn,
                startCell,
                endCell,
                hasHeaders,
                fileName).toJSONString();
    }
}
