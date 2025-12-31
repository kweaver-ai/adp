USE dip_mdl;

-- 扫描记录
CREATE TABLE IF NOT EXISTS t_scan_record (
    f_record_id varchar(40) NOT NULL DEFAULT '' COMMENT '扫描记录 id',
    f_data_source_id varchar(40) NOT NULL COMMENT '数据源 id',
    f_scanner varchar(40) NOT NULL COMMENT '扫描器',
    f_scan_time bigint(20) NOT NULL DEFAULT 0 COMMENT '扫描时间',
    f_data_source_status varchar(20) NOT NULL DEFAULT '' COMMENT '数据源状态: available 可用 scanning 扫描中',
    f_metadata_task_id varchar(128)  DEFAULT NULL COMMENT '元数据采集平台任务id',
    PRIMARY KEY (f_record_id),
    UNIQUE KEY uk_scan_record (f_data_source_id, f_scanner)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '数据源扫描记录表';