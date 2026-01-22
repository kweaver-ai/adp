package com.eisoo.dc.common.enums;

/**
 * @author Tian.lan
 */

public enum OperationTyeEnum {
    INSERT(0, "新增"),
    DELETE(1, "删除"),
    UPDATE(2, "更新"),
    UNKNOWN(3, "未知");

    private final int code;
    private final String desc;

    OperationTyeEnum(int code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    public int getCode() {
        return code;
    }

    public String getDesc() {
        return desc;
    }
}
