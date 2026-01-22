package com.eisoo.dc.metadata.domain.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import io.swagger.annotations.ApiModel;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.io.Serializable;

/**
 * @author Tian.lan
 */
@Data
@JsonInclude(JsonInclude.Include.ALWAYS)
@ApiModel
@AllArgsConstructor
@NoArgsConstructor
public class TableScanDto implements Serializable {
    @JsonProperty(value = "id")
    private String id;
    @JsonProperty(value = "name")
    private String name;
    @JsonProperty(value = "create_time")
    private String createTime;
    @JsonProperty(value = "update_time")
    private String updateTime;
    @JsonProperty(value = "advanced_params")
    private Object advancedParams;

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class AdvancedParam {
        @JsonProperty(value = "sheet")
        private String sheet;
        @JsonProperty(value = "all_sheet")
        private Boolean allSheet;
        @JsonProperty(value = "sheet_as_new_column")
        private Boolean sheetAsNewColumn;
        @JsonProperty(value = "start_cell")
        private String startCell;
        @JsonProperty(value = "end_cell")
        private String endCell;
        @JsonProperty(value = "has_headers")
        private Boolean hasHeaders;
        @JsonProperty(value = "file_name")
        private String fileName;
        @JsonProperty(value = "catalog")
        private String catalog;

        public AdvancedParam(AdvancedParamInternal advancedParam) {
            this.sheet = advancedParam.getSheet();
            this.allSheet = advancedParam.getAllSheet();
            this.sheetAsNewColumn = advancedParam.getSheetAsNewColumn();
            this.startCell = advancedParam.getStartCell();
            this.endCell = advancedParam.getEndCell();
            this.hasHeaders = advancedParam.getHasHeaders();
            this.fileName = advancedParam.getFileName();
        }
    }

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class AdvancedParamInternal {
        @JsonProperty(value = "sheet")
        private String sheet;
        @JsonProperty(value = "allSheet")
        private Boolean allSheet;
        @JsonProperty(value = "sheetAsNewColumn")
        private Boolean sheetAsNewColumn;
        @JsonProperty(value = "startCell")
        private String startCell;
        @JsonProperty(value = "endCell")
        private String endCell;
        @JsonProperty(value = "hasHeaders")
        private Boolean hasHeaders;
        @JsonProperty(value = "fileName")
        private String fileName;
        @JsonProperty(value = "catalog")
        private String catalog;
    }
}
