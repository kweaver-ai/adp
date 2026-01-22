package com.eisoo.dc.common.enums;

import lombok.Getter;

/**
 * @author Tian.lan
 */

@Getter
public enum ScheduleJobStatusEnum {
    OPEN(1, "open"),
    CLOSE(0, "close");
    private final Integer code;
    private final String desc;

    ScheduleJobStatusEnum(Integer code, String desc) {
        this.code = code;
        this.desc = desc;
    }

    // 根据编码获取枚举
    public static ScheduleJobStatusEnum getByCode(Integer code) {
        for (ScheduleJobStatusEnum status : values()) {
            if (status.getCode().equals(code)) {
                return status;
            }
        }
        return null;
    }

    public static String fromCode(int code) {
        for (ScheduleJobStatusEnum status : values()) {
            if (status.code == code) {
                return status.desc;
            }
        }
        return null;
    }

    public static Integer fromDesc(String desc) {
        for (ScheduleJobStatusEnum status : values()) {
            if (status.desc.equals(desc)) {
                return status.code;
            }
        }
        return null;
    }
}
