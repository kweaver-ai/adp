package com.eisoo.dc.common.util;

public class LockUtil {
    public static final GlobalMultiTaskLock GLOBAL_MULTI_TASK_LOCK = new GlobalMultiTaskLock();

    public static final GlobalMultiTaskLock SCHEDULE_SCAN_TASK_LOCK = new GlobalMultiTaskLock();

}
