package com.eisoo.dc.common.enums;

/**
 * @author Tian.lan
 */

public enum ScanTaskEnm {
    IMMEDIATE_DS(0, "即时快速:数据源"),
    IMMEDIATE_TABLES(1, "即时:多表"),
    SCHEDULE_DS(2, "定时:数据源");
    //    SCHEDULE_FAST_DS(4, "定时快速:数据源");
    private final int code;
    private final String desc;

    ScanTaskEnm(int code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    public int getCode() {
        return code;
    }

    public String getDesc() {
        return desc;
    }

    public static ScanTaskEnm getByCode(int code) {
        for (ScanTaskEnm value : ScanTaskEnm.values()) {
            if (value.code == code) {
                return value;
            }
        }
        return null;
    }

}
