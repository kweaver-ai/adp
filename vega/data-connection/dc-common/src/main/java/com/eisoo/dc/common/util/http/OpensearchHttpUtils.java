package com.eisoo.dc.common.util.http;

import cn.hutool.http.HttpRequest;
import cn.hutool.http.HttpResponse;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.metadata.entity.DataSourceEntity;
import com.eisoo.dc.common.util.RSAUtil;
import com.eisoo.dc.common.util.StringUtils;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;

import java.util.List;

/**
 * @author Tian.lan
 */
@Slf4j
public class OpensearchHttpUtils {
    public static String queryStatement(DataSourceEntity config, String index, String dsl) throws Exception {
        String url = config.getFConnectProtocol() + "://" + config.getFHost() + ":" + config.getFPort() + "/" + index + "/_search";
        if (StringUtils.isEmpty(index)) {
            url = config.getFConnectProtocol() + "://" + config.getFHost() + ":" + config.getFPort() + "/_search";
        }
        log.info("url:{}", url);
        HttpResponse response = null;
        try {
            // 1. 创建Hutool HttpRequest并配置认证
            HttpRequest request = HttpRequest.post(url)
                    .body(dsl) // 设置DSL请求体
                    .timeout(5000) // 超时时间
                    .header("Content-Type", "application/json"); // 设置JSON类型
            // 2. 添加Basic认证（如果配置了用户名密码）
            if (config.getFAccount() != null && !config.getFPassword().isEmpty()) {
                request.basicAuth(config.getFAccount(), RSAUtil.decrypt(config.getFPassword()));
            }
            // 3. 发送请求并获取响应
            response = request.execute();
            // 4. 处理响应状态
            if (!response.isOk()) {
                log.error("OpenSearchUtil查询失败: httpStatus={}, response={}", response.getStatus(), response);
                String error = response.body();
                try {
                    JSONObject jsonObject = JSONObject.parseObject(error);
                    error = jsonObject.getJSONObject("error").toString();
                } catch (Exception e) {
                    log.error("OpenSearchUtil response parse fail:body={}", error, e);
                }
                throw new Exception(error);
            } else {
                return response.body();
            }
        } catch (Exception e) {
            throw new Exception(e);
        } finally {
            if (response != null) {
                response.close();
            }
        }
    }

    @Data
    public static class SearchResult<T> {
        private boolean success;
        private String message;
        private long totalHits; // 总命中数
        private int took; // 耗时（毫秒）
        private List<T> data; // 结果数据
    }
}
