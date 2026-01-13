package com.eisoo.dc.common.util.jdbc.db.impl;

import com.eisoo.dc.common.metadata.entity.FieldScanEntity;

import java.sql.ResultSet;
import java.sql.SQLException;

/**
 * @author Tian.lan
 */
public class MaxComputeJdbcClient extends JdbcBaseClient {
    @Override
    protected boolean needProcessColumnSet() {
        return true;
    }

    @Override
    protected void processColumnSet(ResultSet columnSet, FieldScanEntity fieldScanEntity) throws Exception {
        String typeName = columnSet.getString("TYPE_NAME");
        String type = fieldScanEntity.getFFieldType().toLowerCase();
        Integer length = fieldScanEntity.getFFieldLength();
        Integer precision = fieldScanEntity.getFFieldPrecision();
        if (null != typeName) {
            if ("char".equalsIgnoreCase(type) && 0 == length) {
                length = Integer.valueOf(typeName.split("CHAR\\(")[1].split("\\)")[0]);
            } else if ("varchar".equalsIgnoreCase(type) && 0 == length) {
                length = Integer.valueOf(typeName.split("VARCHAR\\(")[1].split("\\)")[0]);
            } else if ("decimal".equalsIgnoreCase(type)) {
                if (typeName.contains("DECIMAL(") && typeName.contains(",")) {
                    String[] arr = typeName.split("DECIMAL\\(")[1].split("\\)")[0].split(",");
                    length = Integer.valueOf(arr[0]);
                    precision = Integer.valueOf(arr[1]);
                }
            }
            fieldScanEntity.setFFieldLength(length);
            fieldScanEntity.setFFieldPrecision(precision);
        }
    }
    @Override
    protected String processDecimalDigits(ResultSet columnSet) throws SQLException {
        return columnSet.getLong("DECIMAL_DIGITS") + "";
    }
}
