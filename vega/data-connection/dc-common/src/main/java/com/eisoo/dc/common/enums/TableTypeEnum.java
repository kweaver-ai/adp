package com.eisoo.dc.common.enums;

/**
 * @author Tian.lan
 */

public enum TableTypeEnum {
    TABLE(0, "table"),
    VIEW(1, "view");
    private final int code;
    private final String desc;

    TableTypeEnum(int code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    public int getCode() {
        return code;
    }

    public String getDesc() {
        return desc;
    }

    /**
     * 根据状态编码获取枚举实例
     *
     * @param code 状态编码
     * @return 对应的枚举实例，不存在则返回null
     */
    public static String fromCode(int code) {
        for (TableTypeEnum status : values()) {
            if (status.code == code) {
                return status.desc;
            }
        }
        return null;
    }

    public static Integer fromDesc(String desc) {
        for (TableTypeEnum status : values()) {
            if (status.desc.equals(desc)) {
                return status.code;
            }
        }
        return null;
    }
}
