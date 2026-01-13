package com.eisoo.dc.common.util.jdbc.db.impl;

import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import org.apache.commons.lang3.StringUtils;

import java.sql.ResultSet;
import java.sql.SQLException;

/**
 * @author Tian.lan
 */
public class SqlServerJdbcClient extends JdbcBaseClient {
    @Override
    protected String processDecimalDigits(ResultSet columnSet) throws SQLException {
        return columnSet.getString("DECIMAL_DIGITS");
    }
    @Override
    protected boolean needProcessColumnSet()
    {
        return true;
    }

    @Override
    protected void processColumnSet(ResultSet columnSet, FieldScanEntity fieldScanEntity) throws Exception {
        String type = fieldScanEntity.getFFieldType().toLowerCase();
        Integer length = fieldScanEntity.getFFieldLength();
        Integer precision = fieldScanEntity.getFFieldPrecision();
        if ("decimal".equals(type) || "numeric".equals(type)) {
            length = columnSet.getInt("DECIMAL_DIGITS");
            precision = columnSet.getInt("SCALE");
        }
        if ("float".equals(type) || "real".equals(type)) {
            length = columnSet.getInt("DECIMAL_DIGITS");
            precision = length / 2;
        }
        if (StringUtils.isNotBlank(type) && type.contains("int")) {
            length = columnSet.getInt("DECIMAL_DIGITS");
            precision = 0;
        }
        fieldScanEntity.setFFieldLength(length);
        fieldScanEntity.setFFieldPrecision(precision);
    }
}
