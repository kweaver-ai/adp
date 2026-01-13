package com.eisoo.dc.common.metadata.entity;


import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;

@JsonInclude(JsonInclude.Include.ALWAYS)
@AllArgsConstructor
public class AdvancedParamsDTO {
    @JsonProperty(value = "key")
    private String key;

    public String getKey() {
        return this.key;
    }
    public void setKey(String key) {
        this.key = key;
    }
    @JsonProperty(value = "value")
    private String value;
    public String getValue() {
        return this.value;
    }

    public void setValue(String value) {
        this.value = value;
    }
}
