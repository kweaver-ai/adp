package com.eisoo.dc.common.util.jdbc.db.impl;

import java.sql.ResultSet;
import java.sql.SQLException;

/**
 * @author Tian.lan
 */
public class Inceptor2JdbcClient extends JdbcBaseClient {
    @Override
    protected String processDecimalDigits(ResultSet columnSet) throws SQLException {
        return columnSet.getString("DECIMAL_DIGITS");
    }
}
