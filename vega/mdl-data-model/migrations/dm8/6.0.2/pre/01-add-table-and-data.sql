SET SCHEMA dip_mdl;

CREATE TABLE IF NOT EXISTS t_scan_record (
    f_record_id VARCHAR(40 CHAR) NOT NULL DEFAULT '',
    f_data_source_id VARCHAR(40 CHAR) NOT NULL,
    f_scanner VARCHAR(40 CHAR) NOT NULL,
    f_scan_time BIGINT NOT NULL DEFAULT 0,
    f_data_source_status VARCHAR(20 CHAR) NOT NULL DEFAULT '',
    f_metadata_task_id VARCHAR(128 CHAR)  DEFAULT NULL,
    CLUSTER PRIMARY KEY (f_record_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS t_scan_record_uk_scan_record ON t_scan_record(f_data_source_id, f_scanner);

