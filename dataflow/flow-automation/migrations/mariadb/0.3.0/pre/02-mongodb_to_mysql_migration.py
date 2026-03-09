#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
import logging
import os
import sys
import threading
import time
from dataclasses import dataclass, field
from typing import Any, Callable, Dict, Iterable, Iterator, List, Optional, Sequence


logger = logging.getLogger(__name__)
_snowflake_lock = threading.Lock()
_snowflake_last_ms = 0
_snowflake_sequence = 0


def configure_logging() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s - %(levelname)s - %(message)s",
        handlers=[logging.StreamHandler(sys.stdout)],
    )


def json_default(value: Any) -> str:
    if hasattr(value, "isoformat"):
        return value.isoformat()
    return str(value)


def json_string(value: Any) -> str:
    if value is None:
        return ""
    return json.dumps(value, ensure_ascii=False, default=json_default)


def stored_text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value
    return json.dumps(value, ensure_ascii=False, default=json_default)


def pick(document: Dict[str, Any], *keys: str, default: Any = None) -> Any:
    for key in keys:
        if key in document and document[key] is not None:
            return document[key]
    return default


def to_bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if value in (1, "1", "true", "True", "yes", "on"):
        return True
    return False


def to_int(value: Any, default: int = 0) -> int:
    if value is None or value == "":
        return default
    if isinstance(value, bool):
        return int(value)
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return int(value)
    text = str(value).strip()
    if not text:
        return default
    return int(text)


def to_uint64(value: Any) -> int:
    result = to_int(value)
    if result < 0:
        raise ValueError(f"negative unsigned integer: {value}")
    return result


def stable_uint64(*parts: Any) -> int:
    raw = "::".join(str(part) for part in parts)
    digest = hashlib.sha1(raw.encode("utf-8")).digest()
    value = int.from_bytes(digest[:8], "big", signed=False)
    return value or 1


def generate_snowflake_id() -> int:
    global _snowflake_last_ms
    global _snowflake_sequence

    with _snowflake_lock:
        current_ms = int(time.time() * 1000)
        if current_ms == _snowflake_last_ms:
            _snowflake_sequence = (_snowflake_sequence + 1) & 0xFFF
            if _snowflake_sequence == 0:
                while current_ms <= _snowflake_last_ms:
                    current_ms = int(time.time() * 1000)
        else:
            _snowflake_sequence = 0

        _snowflake_last_ms = current_ms
        epoch = 1577808000000
        timestamp = max(current_ms - epoch, 0)
        worker_id = 1
        datacenter_id = 1
        return ((timestamp & ((1 << 41) - 1)) << 22) | ((datacenter_id & 0x1F) << 17) | ((worker_id & 0x1F) << 12) | _snowflake_sequence


def version_string(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        return value.strip()
    if isinstance(value, dict):
        if {"major", "minor", "patch"}.issubset(value.keys()):
            return f"v{value['major']}.{value['minor']}.{value['patch']}"
    return stored_text(value)


def collection_name(prefix: str, suffix: str) -> str:
    clean = prefix.strip("_")
    return f"{clean}_{suffix}" if clean else suffix


def chunked(items: Sequence[Any], size: int) -> Iterator[Sequence[Any]]:
    for index in range(0, len(items), size):
        yield items[index:index + size]


def has_non_empty_command(value: Any) -> bool:
    if value is None:
        return False
    if isinstance(value, dict):
        return bool(value)
    if isinstance(value, (list, tuple, set)):
        return bool(value)
    if isinstance(value, str):
        return value.strip() != ""
    return True


def batch_run_id_from_vars(vars_data: Any) -> str:
    if not isinstance(vars_data, dict):
        return ""
    value = vars_data.get("batch_run_id")
    if isinstance(value, dict):
        return stored_text(value.get("value", ""))
    return stored_text(value)


def build_dag_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_user_id": stored_text(document.get("userid")),
        "f_name": stored_text(document.get("name")),
        "f_desc": stored_text(document.get("desc")),
        "f_trigger": stored_text(document.get("trigger")),
        "f_cron": stored_text(document.get("cron")),
        "f_vars": json_string(document.get("vars")),
        "f_status": stored_text(document.get("status")),
        "f_tasks": json_string(document.get("tasks")),
        "f_steps": json_string(document.get("steps")),
        "f_description": stored_text(document.get("description")),
        "f_shortcuts": json_string(document.get("shortcuts")),
        "f_accessors": json_string(document.get("accessors")),
        "f_type": stored_text(document.get("type")),
        "f_policy_type": stored_text(document.get("policy_type")),
        "f_appinfo": json_string(document.get("appinfo")),
        "f_priority": stored_text(document.get("priority")),
        "f_removed": to_bool(document.get("removed")),
        "f_emails": json_string(document.get("emails")),
        "f_template": stored_text(document.get("template")),
        "f_published": to_bool(pick(document, "publish", "published", default=False)),
        "f_trigger_config": json_string(document.get("trigger_config")),
        "f_sub_ids": json_string(document.get("sub_ids")),
        "f_exec_mode": stored_text(document.get("exec_mode")),
        "f_category": stored_text(document.get("category")),
        "f_outputs": json_string(document.get("outputs")),
        "f_instructions": json_string(document.get("instructions")),
        "f_operator_id": stored_text(document.get("operator_id")),
        "f_inc_values": json_string(document.get("inc_values")),
        "f_version": version_string(document.get("version")),
        "f_version_id": stored_text(document.get("versionId")),
        "f_modify_by": stored_text(document.get("modify_by")),
        "f_is_debug": to_bool(document.get("is_debug")),
        "f_debug_id": stored_text(document.get("debug_id")),
        "f_biz_domain_id": stored_text(document.get("biz_domain_id")),
    }


def build_dag_var_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = to_uint64(document["_id"])
    vars_data = document.get("vars") or {}
    if not isinstance(vars_data, dict):
        return []
    rows: List[Dict[str, Any]] = []
    for var_name, payload in vars_data.items():
        payload = payload or {}
        default_value = payload.get("defaultValue", "") if isinstance(payload, dict) else ""
        description = payload.get("desc", "") if isinstance(payload, dict) else ""
        rows.append(
            {
                "f_id": stable_uint64("dag_var", dag_id, var_name),
                "f_dag_id": dag_id,
                "f_var_name": stored_text(var_name),
                "f_default_value": stored_text(default_value),
                "f_var_type": "string",
                "f_description": stored_text(description),
            }
        )
    return rows


def build_dag_step_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = to_uint64(document["_id"])
    rows: List[Dict[str, Any]] = []

    def add_row(path: str, operator: str, has_datasource: bool) -> None:
        rows.append(
            {
                "f_id": stable_uint64("dag_step", dag_id, path, operator, has_datasource),
                "f_dag_id": dag_id,
                "f_operator": stored_text(operator),
                "f_source_id": "",
                "f_has_datasource": has_datasource,
            }
        )

    def walk(steps: Any, prefix: str) -> None:
        if not isinstance(steps, list):
            return
        for index, step in enumerate(steps):
            if not isinstance(step, dict):
                continue
            step_id = stored_text(step.get("id")) or str(index)
            path = f"{prefix}.{index}.{step_id}" if prefix else f"{index}.{step_id}"
            operator = stored_text(step.get("operator"))
            has_datasource = isinstance(step.get("dataSource"), dict)
            if operator:
                add_row(path, operator, has_datasource)
            elif has_datasource:
                add_row(path, "", True)
            walk(step.get("steps"), f"{path}.steps")
            branches = step.get("branches")
            if isinstance(branches, list):
                for branch_index, branch in enumerate(branches):
                    if isinstance(branch, dict):
                        walk(branch.get("steps"), f"{path}.branches.{branch_index}")

    walk(document.get("steps"), "")
    return rows


def build_dag_trigger_config_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = to_uint64(document["_id"])
    config = document.get("trigger_config") or {}
    if not isinstance(config, dict):
        return []
    operator = stored_text(config.get("operator"))
    data_source = config.get("dataSource") or {}
    source_id = stored_text(data_source.get("id")) if isinstance(data_source, dict) else ""
    if not operator and not source_id:
        return []
    return [
        {
            "f_id": stable_uint64("dag_trigger", dag_id, operator, source_id),
            "f_dag_id": dag_id,
            "f_operator": operator,
            "f_source_id": source_id,
        }
    ]


def build_dag_accessor_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_id = to_uint64(document["_id"])
    accessors = document.get("accessors") or []
    if not isinstance(accessors, list):
        return []
    rows: List[Dict[str, Any]] = []
    for accessor in accessors:
        if not isinstance(accessor, dict):
            continue
        accessor_id = stored_text(accessor.get("id"))
        if not accessor_id:
            continue
        rows.append(
            {
                "f_id": stable_uint64("dag_accessor", dag_id, accessor_id),
                "f_dag_id": dag_id,
                "f_accessor_id": accessor_id,
            }
        )
    return rows


def build_dag_version_row(document: Dict[str, Any]) -> Dict[str, Any]:
    config = document.get("config")
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_dag_id": stored_text(document.get("dagId")),
        "f_user_id": stored_text(document.get("userid")),
        "f_version": version_string(document.get("version")),
        "f_version_id": stored_text(document.get("versionId")),
        "f_change_log": stored_text(document.get("changeLog")),
        "f_config": stored_text(config),
        "f_sort_time": to_int(document.get("sortTime")),
    }


def build_dag_instance_row(document: Dict[str, Any]) -> Dict[str, Any]:
    vars_data = document.get("vars")
    cmd = document.get("cmd")
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_dag_id": to_uint64(document.get("dagId")),
        "f_trigger": stored_text(document.get("trigger")),
        "f_worker": stored_text(document.get("worker")),
        "f_source": stored_text(document.get("source")),
        "f_vars": json_string(vars_data),
        "f_keywords": json_string(document.get("keywords")),
        "f_event_persistence": to_int(document.get("eventPersistence")),
        "f_event_oss_path": stored_text(document.get("eventOssPath")),
        "f_share_data": json_string(document.get("shareData")),
        "f_share_data_ext": json_string(document.get("shareDataExt")),
        "f_status": stored_text(document.get("status")),
        "f_reason": stored_text(document.get("reason")),
        "f_cmd": json_string(cmd),
        "f_has_cmd": has_non_empty_command(cmd),
        "f_batch_run_id": batch_run_id_from_vars(vars_data),
        "f_user_id": stored_text(document.get("userid")),
        "f_ended_at": to_int(document.get("endedAt")),
        "f_dag_type": stored_text(document.get("dag_type")),
        "f_policy_type": stored_text(document.get("policy_type")),
        "f_appinfo": json_string(document.get("appinfo")),
        "f_priority": stored_text(document.get("priority")),
        "f_mode": to_int(document.get("mode")),
        "f_dump": stored_text(document.get("dump")),
        "f_dump_ext": json_string(document.get("dumpExt")),
        "f_success_callback": stored_text(document.get("success_callback")),
        "f_error_callback": stored_text(document.get("error_callback")),
        "f_call_chain": json_string(document.get("call_chain")),
        "f_resume_data": stored_text(document.get("resume_data")),
        "f_resume_status": stored_text(document.get("resume_status")),
        "f_version": version_string(document.get("version")),
        "f_version_id": stored_text(document.get("versionId")),
        "f_biz_domain_id": stored_text(document.get("biz_domain_id")),
    }


def build_dag_instance_keyword_rows(document: Dict[str, Any]) -> List[Dict[str, Any]]:
    dag_ins_id = to_uint64(document["_id"])
    keywords = document.get("keywords") or []
    if not isinstance(keywords, list):
        return []
    rows: List[Dict[str, Any]] = []
    for keyword in keywords:
        keyword_text = stored_text(keyword)
        if not keyword_text:
            continue
        rows.append(
            {
                "f_id": stable_uint64("dag_keyword", dag_ins_id, keyword_text),
                "f_dag_ins_id": dag_ins_id,
                "f_keyword": keyword_text,
            }
        )
    return rows


def build_task_instance_row(document: Dict[str, Any]) -> Dict[str, Any]:
    updated_at = to_int(document.get("updatedAt"))
    timeout_secs = to_int(document.get("timeoutSecs"))
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": updated_at,
        "f_expired_at": updated_at + timeout_secs,
        "f_task_id": stored_text(document.get("taskId")),
        "f_dag_ins_id": to_uint64(pick(document, "dagInsId", "dagInsID", default=0)),
        "f_name": stored_text(document.get("name")),
        "f_depend_on": json_string(document.get("dependOn")),
        "f_action_name": stored_text(document.get("actionName")),
        "f_timeout_secs": timeout_secs,
        "f_params": json_string(document.get("params")),
        "f_traces": json_string(document.get("traces")),
        "f_status": stored_text(document.get("status")),
        "f_reason": json_string(document.get("reason")),
        "f_pre_checks": json_string(pick(document, "preChecks", "preCheck")),
        "f_results": json_string(document.get("results")),
        "f_steps": json_string(document.get("steps")),
        "f_last_modified_at": to_int(document.get("lastModifiedAt")),
        "f_rendered_params": json_string(document.get("renderedParams")),
        "f_hash": stored_text(document.get("hash")),
        "f_settings": json_string(document.get("settings")),
        "f_metadata": json_string(document.get("metadata")),
    }


def build_token_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_user_id": stored_text(document.get("userid")),
        "f_user_name": stored_text(document.get("username")),
        "f_refresh_token": stored_text(document.get("refresh_token")),
        "f_token": stored_text(document.get("token")),
        "f_expires_in": to_int(document.get("expires_in")),
        "f_login_ip": stored_text(document.get("login_ip")),
        "f_is_app": to_bool(document.get("isapp")),
    }


def build_inbox_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_msg": json_string(document.get("msg")),
        "f_topic": stored_text(document.get("topic")),
        "f_docid": stored_text(document.get("docid")),
        "f_dag": json_string(pick(document, "dag", "dags")),
    }


def build_client_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": generate_snowflake_id(),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_client_name": stored_text(document.get("client_name")),
        "f_client_id": stored_text(document.get("client_id")),
        "f_client_secret": stored_text(document.get("client_secret")),
    }


def build_switch_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_name": stored_text(document.get("name")),
        "f_status": to_bool(document.get("status")),
    }


def build_log_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_ossid": stored_text(document.get("ossid")),
        "f_key": stored_text(document.get("key")),
        "f_filename": stored_text(document.get("filename")),
    }


def build_outbox_row(document: Dict[str, Any]) -> Dict[str, Any]:
    return {
        "f_id": to_uint64(document["_id"]),
        "f_created_at": to_int(document.get("createdAt")),
        "f_updated_at": to_int(document.get("updatedAt")),
        "f_msg": stored_text(document.get("msg")),
        "f_topic": stored_text(document.get("topic")),
    }


@dataclass(frozen=True)
class ChildTableConfig:
    table_name: str
    pk_column: str
    row_builder: Callable[[Dict[str, Any]], List[Dict[str, Any]]]


@dataclass(frozen=True)
class CollectionMapping:
    name: str
    collection_suffix: str
    table_name: str
    pk_column: str
    row_builder: Callable[[Dict[str, Any]], Dict[str, Any]]
    identity_column: Optional[str] = None
    child_tables: Sequence[ChildTableConfig] = field(default_factory=tuple)


@dataclass
class TableStats:
    scanned: int = 0
    inserted: int = 0
    skipped: int = 0
    failed: int = 0


MAPPINGS: Sequence[CollectionMapping] = (
    CollectionMapping(
        name="dag",
        collection_suffix="dag",
        table_name="t_flow_dag",
        pk_column="f_id",
        row_builder=build_dag_row,
        child_tables=(
            ChildTableConfig("t_flow_dag_var", "f_id", build_dag_var_rows),
            ChildTableConfig("t_flow_dag_step", "f_id", build_dag_step_rows),
            ChildTableConfig("t_flow_dag_trigger_config", "f_id", build_dag_trigger_config_rows),
            ChildTableConfig("t_flow_dag_accessor", "f_id", build_dag_accessor_rows),
        ),
    ),
    CollectionMapping("dag_version", "dag_version", "t_flow_dag_version", "f_id", build_dag_version_row),
    CollectionMapping(
        name="dag_instance",
        collection_suffix="dag_instance",
        table_name="t_flow_dag_instance",
        pk_column="f_id",
        row_builder=build_dag_instance_row,
        child_tables=(ChildTableConfig("t_flow_dag_instance_keyword", "f_id", build_dag_instance_keyword_rows),),
    ),
    CollectionMapping("task_instance", "task_instance", "t_flow_task_instance", "f_id", build_task_instance_row),
    CollectionMapping("token", "token", "t_flow_token", "f_id", build_token_row),
    CollectionMapping("inbox", "inbox", "t_flow_inbox", "f_id", build_inbox_row),
    CollectionMapping("client", "client", "t_flow_client", "f_id", build_client_row, identity_column="f_client_name"),
    CollectionMapping("switch", "switch", "t_flow_switch", "f_id", build_switch_row),
    CollectionMapping("log", "log", "t_flow_log", "f_id", build_log_row),
    CollectionMapping("outbox", "outbox", "t_flow_outbox", "f_id", build_outbox_row),
)


class DatabaseManager:
    def __init__(self, mongo_database: str, mysql_database: str, mongo_prefix: str) -> None:
        self.mongo_database_name = mongo_database
        self.mysql_database_name = mysql_database
        self.mongo_prefix = mongo_prefix
        self.mongo_client = None
        self.mongo_db = None
        self.mysql_conn = None

    def connect_mongodb(self):
        try:
            from pymongo import MongoClient
        except ImportError as exc:
            raise RuntimeError("missing dependency: pymongo") from exc

        uri = os.getenv("MONGODB_URI")
        if not uri:
            host = os.getenv("MONGODB_HOST", "127.0.0.1")
            port = os.getenv("MONGODB_PORT", "27017")
            user = os.getenv("MONGODB_USER", "")
            password = os.getenv("MONGODB_PASSWORD", "")
            auth_source = os.getenv("MONGODB_AUTH_SOURCE", "admin")
            if user:
                uri = f"mongodb://{user}:{password}@{host}:{port}?authSource={auth_source}"
            else:
                uri = f"mongodb://{host}:{port}"

        client = MongoClient(uri, serverSelectionTimeoutMS=5000)
        client.admin.command("ping")
        self.mongo_client = client
        self.mongo_db = client[self.mongo_database_name]
        logger.info("MongoDB connected: %s", self.mongo_database_name)
        return self.mongo_db

    def connect_mysql(self):
        driver = None
        for module_name in ("rdsdriver", "pymysql"):
            try:
                driver = __import__(module_name)
                break
            except ImportError:
                continue
        if driver is None:
            raise RuntimeError("missing dependency: rdsdriver or pymysql")

        params = {
            "host": os.getenv("DB_HOST", "127.0.0.1"),
            "port": int(os.getenv("DB_PORT", "3306")),
            "user": os.getenv("DB_USER", "root"),
            "password": os.getenv("DB_PASSWD", os.getenv("DB_PASSWORD", "")),
            "charset": "utf8mb4",
            "autocommit": True,
        }
        try:
            self.mysql_conn = driver.connect(**params)
        except TypeError:
            params["passwd"] = params.pop("password")
            self.mysql_conn = driver.connect(**params)
        logger.info("MySQL connected: %s:%s", params["host"], params["port"])
        return self.mysql_conn

    def close(self) -> None:
        if self.mysql_conn:
            self.mysql_conn.close()
        if self.mongo_client:
            self.mongo_client.close()


class Migrator:
    def __init__(self, db_manager: DatabaseManager, batch_size: int, dry_run: bool) -> None:
        self.db_manager = db_manager
        self.batch_size = batch_size
        self.dry_run = dry_run
        self.stats: Dict[str, TableStats] = {}

    def table_stats(self, table_name: str) -> TableStats:
        if table_name not in self.stats:
            self.stats[table_name] = TableStats()
        return self.stats[table_name]

    def qualify_table(self, table_name: str) -> str:
        return f"`{self.db_manager.mysql_database_name}`.`{table_name}`"

    def validate_target_tables(self, mappings: Sequence[CollectionMapping]) -> None:
        target_tables = {mapping.table_name for mapping in mappings}
        for mapping in mappings:
            for child in mapping.child_tables:
                target_tables.add(child.table_name)

        cursor = self.db_manager.mysql_conn.cursor()
        try:
            sql = (
                "SELECT table_name FROM information_schema.tables "
                "WHERE table_schema = %s AND table_name IN ("
                + ",".join(["%s"] * len(target_tables))
                + ")"
            )
            params = [self.db_manager.mysql_database_name, *sorted(target_tables)]
            cursor.execute(sql, params)
            existing = {row[0] for row in cursor.fetchall()}
        finally:
            cursor.close()

        missing = sorted(target_tables - existing)
        if missing:
            raise RuntimeError(f"target tables not found in `{self.db_manager.mysql_database_name}`: {', '.join(missing)}")

    def iter_documents(self, mapping: CollectionMapping) -> Iterator[List[Dict[str, Any]]]:
        collection = self.db_manager.mongo_db[collection_name(self.db_manager.mongo_prefix, mapping.collection_suffix)]
        cursor = collection.find({}, no_cursor_timeout=True).sort("_id", 1).batch_size(self.batch_size)
        batch: List[Dict[str, Any]] = []
        try:
            for document in cursor:
                batch.append(document)
                if len(batch) >= self.batch_size:
                    yield batch
                    batch = []
            if batch:
                yield batch
        finally:
            cursor.close()

    def insert_rows(self, table_name: str, rows: List[Dict[str, Any]], pk_column: str) -> None:
        stats = self.table_stats(table_name)
        if not rows:
            return
        if self.dry_run:
            stats.inserted += len(rows)
            return

        columns = list(rows[0].keys())
        placeholders = ", ".join(["%s"] * len(columns))
        column_list = ", ".join(f"`{column}`" for column in columns)
        sql = f"INSERT IGNORE INTO {self.qualify_table(table_name)} ({column_list}) VALUES ({placeholders})"

        for chunk in chunked(rows, self.batch_size):
            values = [tuple(row[column] for column in columns) for row in chunk]
            cursor = self.db_manager.mysql_conn.cursor()
            try:
                cursor.executemany(sql, values)
                inserted = int(cursor.rowcount or 0)
                stats.inserted += inserted
                stats.skipped += len(chunk) - inserted
            finally:
                cursor.close()

    def load_existing_values(self, table_name: str, column_name: str, values: Sequence[Any]) -> set:
        existing = set()
        if not values:
            return existing
        for chunk in chunked(list(values), self.batch_size):
            cursor = self.db_manager.mysql_conn.cursor()
            try:
                sql = (
                    f"SELECT `{column_name}` FROM {self.qualify_table(table_name)} "
                    f"WHERE `{column_name}` IN ({', '.join(['%s'] * len(chunk))})"
                )
                cursor.execute(sql, list(chunk))
                existing.update(row[0] for row in cursor.fetchall())
            finally:
                cursor.close()
        return existing

    def filter_existing_rows(self, table_name: str, rows: List[Dict[str, Any]], identity_column: Optional[str]) -> List[Dict[str, Any]]:
        if not rows or not identity_column:
            return rows

        identities = []
        seen = set()
        for row in rows:
            identity = row.get(identity_column)
            if identity in (None, "") or identity in seen:
                continue
            identities.append(identity)
            seen.add(identity)

        existing = self.load_existing_values(table_name, identity_column, identities)
        filtered: List[Dict[str, Any]] = []
        batch_seen = set()
        stats = self.table_stats(table_name)
        for row in rows:
            identity = row.get(identity_column)
            if identity in (None, ""):
                filtered.append(row)
                continue
            if identity in existing or identity in batch_seen:
                stats.skipped += 1
                continue
            batch_seen.add(identity)
            filtered.append(row)
        return filtered

    def migrate_mapping(self, mapping: CollectionMapping) -> None:
        logger.info("migrating `%s` from `%s`", mapping.table_name, collection_name(self.db_manager.mongo_prefix, mapping.collection_suffix))
        for documents in self.iter_documents(mapping):
            main_rows: List[Dict[str, Any]] = []
            child_rows: Dict[str, List[Dict[str, Any]]] = {child.table_name: [] for child in mapping.child_tables}

            for document in documents:
                self.table_stats(mapping.table_name).scanned += 1
                try:
                    main_rows.append(mapping.row_builder(document))
                    for child in mapping.child_tables:
                        child_rows[child.table_name].extend(child.row_builder(document))
                except Exception as exc:
                    self.table_stats(mapping.table_name).failed += 1
                    logger.exception("transform failed for `%s` document `%s`: %s", mapping.name, document.get("_id"), exc)

            main_rows = self.filter_existing_rows(mapping.table_name, main_rows, mapping.identity_column)
            self.insert_rows(mapping.table_name, main_rows, mapping.pk_column)
            for child in mapping.child_tables:
                self.insert_rows(child.table_name, child_rows[child.table_name], child.pk_column)

    def log_summary(self) -> None:
        logger.info("migration summary:")
        for table_name in sorted(self.stats):
            stats = self.stats[table_name]
            logger.info(
                "  %s scanned=%s inserted=%s skipped=%s failed=%s",
                table_name,
                stats.scanned,
                stats.inserted,
                stats.skipped,
                stats.failed,
            )


def build_arg_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Migrate Flow Automation data from MongoDB to MySQL.")
    parser.add_argument("--mongo-database", default=os.getenv("MONGODB_DATABASE", os.getenv("MONGO_DATABASE", "automation")))
    parser.add_argument("--mongo-prefix", default=os.getenv("MONGODB_PREFIX", os.getenv("MONGO_PREFIX", os.getenv("STORE_PREFIX", "flow"))))
    parser.add_argument("--mysql-database", default=os.getenv("DB_NAME", os.getenv("DB_DATABASE", os.getenv("MYSQL_DATABASE", "adp"))))
    parser.add_argument("--batch-size", type=int, default=500)
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--tables", nargs="*", choices=[mapping.name for mapping in MAPPINGS], default=[mapping.name for mapping in MAPPINGS])
    return parser


def selected_mappings(names: Sequence[str]) -> List[CollectionMapping]:
    selected = set(names)
    return [mapping for mapping in MAPPINGS if mapping.name in selected]


def main() -> int:
    configure_logging()
    args = build_arg_parser().parse_args()

    db_manager = DatabaseManager(
        mongo_database=args.mongo_database,
        mysql_database=args.mysql_database,
        mongo_prefix=args.mongo_prefix,
    )

    try:
        db_manager.connect_mongodb()
        db_manager.connect_mysql()
        mappings = selected_mappings(args.tables)
        migrator = Migrator(db_manager=db_manager, batch_size=args.batch_size, dry_run=args.dry_run)
        migrator.validate_target_tables(mappings)
        for mapping in mappings:
            migrator.migrate_mapping(mapping)
        migrator.log_summary()
        return 0
    except Exception as exc:
        logger.exception("migration failed: %s", exc)
        return 1
    finally:
        db_manager.close()


if __name__ == "__main__":
    sys.exit(main())
