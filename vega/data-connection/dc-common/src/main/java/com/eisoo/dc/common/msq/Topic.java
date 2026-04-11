package com.eisoo.dc.common.msq;

public enum Topic {
    ISF_AUDIT_LOG_LOG("isf.audit_log.log", "新增审计日志"),
    AUTHORIZATION_RESOURCE_NAME_MODIFY("authorization.resource.name.modify", "修改资源名称"),
    AF_DATASOURCE_MESSAGE_TOPIC("kweaver.datasource", "数据源消息");

    private final String topicName;
    private final String description;

    Topic(String topicName, String description) {
        this.topicName = topicName;
        this.description = description;
    }

    public String getTopicName() {
        return topicName;
    }

    public String getDescription() {
        return description;
    }

    @Override
    public String toString() {
        return topicName;
    }
}
