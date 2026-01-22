package com.eisoo.dc.common.util;

import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.dc.common.config.OpenSearchClientCfg;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.connector.TypeConfig;
import com.eisoo.dc.common.constant.Constants;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.metadata.entity.FieldScanEntity;
import com.eisoo.dc.common.metadata.entity.OpenSearchEntity;
import com.eisoo.dc.common.msq.GlobalConfig;
import com.eisoo.dc.common.vo.Ext;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.hc.client5.http.auth.AuthScope;
import org.apache.hc.client5.http.auth.UsernamePasswordCredentials;
import org.apache.hc.client5.http.config.RequestConfig;
import org.apache.hc.client5.http.impl.async.HttpAsyncClientBuilder;
import org.apache.hc.client5.http.impl.auth.BasicCredentialsProvider;
import org.apache.hc.core5.http.HttpHost;
import org.apache.hc.core5.util.Timeout;
import org.apache.http.conn.ssl.NoopHostnameVerifier;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.ssl.SSLContexts;
import org.apache.http.ssl.TrustStrategy;
import org.opensearch.client.RestClient;
import org.opensearch.client.RestClientBuilder;
import org.opensearch.client.json.jackson.JacksonJsonpMapper;
import org.opensearch.client.opensearch.OpenSearchClient;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.opensearch.client.transport.OpenSearchTransport;
import org.opensearch.client.transport.rest_client.RestClientTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.net.ssl.SSLContext;
import javax.servlet.http.HttpServletRequest;
import java.io.IOException;
import java.lang.reflect.Array;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.ThreadFactory;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 工具类
 */
public class CommonUtil {

    private final static Logger logger = LoggerFactory.getLogger(CommonUtil.class);
    private final static DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    public final static String INSERT = "insert";
    public final static String UPDATE = "update";
    public final static String DELETE = "delete";
    public final static Set<String> OPERATON_TYPES = new HashSet<>(Arrays.asList(INSERT, UPDATE, DELETE));


    public static String getNowTime() {
        return LocalDateTime.now().format(formatter);
    }

    public static final String OPEN_SEARCH = "opensearch";
    public static final String EXCEL = "excel";
    public static final String MYSQL = "mysql";

    public static final String MARIA = "maria";

    public static String obj2json(Object o) {
        try {
            return new ObjectMapper().writeValueAsString(o);
        } catch (IOException e) {
            logger.error("序列化错误：", e);
        }
        return null;
    }

    public static CloseableHttpClient getHttpClient() {
        try {
            SSLContext sslContext = SSLContexts.custom().loadTrustMaterial(null, new TrustStrategy() {
                @Override
                public boolean isTrusted(X509Certificate[] x509Certificates, String s) throws CertificateException {
                    return true;
                }
            }).build();

            CloseableHttpClient httpClient = HttpClients.custom().setSSLContext(sslContext).setSSLHostnameVerifier(new NoopHostnameVerifier()).build();
            return httpClient;
        } catch (Exception e) {
            throw new AiShuException(ErrorCodeEnum.BadRequest, e.getMessage());
        }
    }

    public static IntrospectInfo getOrCreateIntrospectInfo(HttpServletRequest request) {
        IntrospectInfo introspectInfo = new IntrospectInfo();
        if (request != null) {
            introspectInfo = (IntrospectInfo) request.getAttribute("introspectInfo");
        }
        if (introspectInfo.getExt() == null) {
            introspectInfo.setExt(new Ext());
        }
        return introspectInfo;
    }

    public static String getToken(HttpServletRequest request) {
        String token = "";
        if (request != null) {
            token = (String) request.getAttribute(Constants.HEADER_TOKEN_KEY);
        }
        return token;
    }

    public static ThreadFactory threadFactory(final String name) {
        final AtomicLong count = new AtomicLong();
        return new ThreadFactory() {
            @Override
            public Thread newThread(Runnable runnable) {
                Thread thread = Executors.defaultThreadFactory().newThread(runnable);
                thread.setName(name + "-" + Long.toString(count.getAndIncrement()));
                // doesn't catch everything, just for extra safety
                thread.setUncaughtExceptionHandler(new Thread.UncaughtExceptionHandler() {
                    @Override
                    public void uncaughtException(Thread t, Throwable e) {
                        logger.error("uncaught error", e);
                    }
                });
                return thread;
            }
        };
    }

    // 字符串转 int 失败使用默认值
    public static int parseInt(String string, int def) {
        int value = def;

        if (string == null) {
            return value;
        }
        try {
            value = Integer.parseInt(string);
        } catch (Exception e) {
            return def;
        }

        return value;
    }

    public static int validPollIntervalms(Properties config) {
        int pollIntervalms = parseInt(config.getProperty(GlobalConfig.POll_INTERVAL_MILLISECONDS),
                GlobalConfig.DEFAULT_POll_INTERVAL_MILLISECONDS);

        pollIntervalms = pollIntervalms > GlobalConfig.DEFAULT_POll_INTERVAL_MILLISECONDS
                ? GlobalConfig.DEFAULT_POll_INTERVAL_MILLISECONDS
                : pollIntervalms;

        pollIntervalms = pollIntervalms < 1 ? 1 : pollIntervalms;

        return pollIntervalms;
    }

    public static int validMaxInFlight(Properties config) {
        int maxInFlight = parseInt(config.getProperty(GlobalConfig.MAX_INFLIGHT),
                GlobalConfig.DEFAULT_MAX_INFLIGHT);

        maxInFlight = maxInFlight > GlobalConfig.DEFAULT_MAX_INFLIGHT ? GlobalConfig.DEFAULT_MAX_INFLIGHT : maxInFlight;
        maxInFlight = maxInFlight < 1 ? 1 : maxInFlight;

        return maxInFlight;
    }

    public static int validMsgTimeout(Properties config) {
        int msgTimeout = parseInt(config.getProperty(GlobalConfig.MSG_TIMEOUT_SECONDS),
                GlobalConfig.DEFAULT_MSG_TIMEOUT_SECONDS);

        msgTimeout = msgTimeout < 1 ? 1 : msgTimeout;

        return msgTimeout;
    }

    public static int validRetryTimes(Properties config) {

        int retryTimes = parseInt(config.getProperty(GlobalConfig.MSG_FAILED_RETRY_TIMES),
                GlobalConfig.DEFAULT_MSG_FAILED_RETRY_TIMES);

        retryTimes = retryTimes < 1 ? GlobalConfig.DEFAULT_MSG_FAILED_RETRY_TIMES : retryTimes;
        return retryTimes;
    }

    public static OpenSearchClient getOpenSearchClient(OpenSearchClientCfg openSearchClientCfg) {
        final HttpHost host = new HttpHost(openSearchClientCfg.getProtocol(),
                openSearchClientCfg.getHost(),
                openSearchClientCfg.getPort());
        final BasicCredentialsProvider credentialsProvider = new BasicCredentialsProvider();
        credentialsProvider.setCredentials(new AuthScope(host), new UsernamePasswordCredentials(openSearchClientCfg.getUserName(),
                openSearchClientCfg.getPassWord().toCharArray()));
        final RestClient restClient = RestClient.builder(host).setHttpClientConfigCallback(new RestClientBuilder.HttpClientConfigCallback() {
            @Override
            public HttpAsyncClientBuilder customizeHttpClient(HttpAsyncClientBuilder httpClientBuilder) {
                return httpClientBuilder.setDefaultCredentialsProvider(credentialsProvider);
            }
        }).setRequestConfigCallback(new RestClientBuilder.RequestConfigCallback() {
            @Override
            public RequestConfig.Builder customizeRequestConfig(RequestConfig.Builder builder) {
                return builder.setConnectTimeout(Timeout.ofSeconds(5))
                        .setResponseTimeout(Timeout.ofSeconds(30));
            }
        }).build();
        final OpenSearchTransport transport = new RestClientTransport(restClient, new JacksonJsonpMapper());
        OpenSearchClient openSearchClient = new OpenSearchClient(transport);
        return openSearchClient;
    }

    public static String getVirtualType(String advancedParams) {
        String virtualFieldType = "";
        if (isEmpty(advancedParams)) {
            return virtualFieldType;
        }
        HashMap[] array = JSON.parseObject(advancedParams, HashMap[].class);
        int size = array.length;
        for (int i = 0; i < size; i++) {
            HashMap map = array[i];
            String key = (String) map.get("key");
            if ("virtualFieldType".equals(key)) {
                virtualFieldType = (String) map.getOrDefault("value", "");
                break;
            }
        }
        return virtualFieldType.toLowerCase();
    }

    public static String getOpenSearchFieldParam(ConnectorConfig connectorConfig, OpenSearchEntity.OpenSearchField field) {
//        {
//            "key": "originFieldType",
//                "value": "bpchar"
//        },
//        {
//            "key": "virtualFieldType",
//                "value": "char"
//        }
        JSONArray result = new JSONArray();
        HashMap<String, String> typeMap = new HashMap<>();
        List<TypeConfig> type = connectorConfig.getType();
        for (TypeConfig typeConfig : type) {
            typeMap.put(typeConfig.getSourceType(), typeConfig.getVegaType());
        }
        JSONObject o1 = new JSONObject();
        o1.put("key", "originFieldType");
        o1.put("value", field.getType().toLowerCase());
        JSONObject o2 = new JSONObject();
        o2.put("key", "virtualFieldType");
        o2.put("value", typeMap.get(field.getType().toLowerCase()));
        result.add(o1);
        result.add(o2);

        String keywordType = field.getKeywordType();
        if (isNotEmpty(keywordType)) {
            JSONObject o = new JSONObject();
            o.put("key", "fields.keyword.type");
            o.put("value", keywordType);
            result.add(o);
        }
        Integer ignoreAbove = field.getIgnoreAbove();
        if (isNotEmpty(ignoreAbove)) {
            JSONObject o = new JSONObject();
            o.put("key", "fields.keyword.ignore_above");
            o.put("value", ignoreAbove);
            result.add(o);
        }
        Boolean norms = field.getNorms();
        if (isNotEmpty(norms)) {
            JSONObject o = new JSONObject();
            o.put("key", "norms");
            o.put("value", norms);
            result.add(o);
        }
        String analyzer = field.getAnalyzer();
        if (isNotEmpty(analyzer)) {
            JSONObject o = new JSONObject();
            o.put("key", "analyzer");
            o.put("value", analyzer);
            result.add(o);
        }
        return result.toJSONString();


    }

    public static boolean judgeTwoFiledIsChane(FieldScanEntity newFieldScanEntity, FieldScanEntity oldFieldScanEntity) {
        // opensearch 没有f_field_length  f_field_precision  f_field_comment
        String fieldTypeNew = newFieldScanEntity.getFFieldType();
        String fieldTypeOld = oldFieldScanEntity.getFFieldType();
        // 类型变化
        if (!fieldTypeNew.equals(fieldTypeOld)) {
            return true;
        }
        if ("text".equalsIgnoreCase(fieldTypeNew)) {
            String paramsNew = newFieldScanEntity.getFAdvancedParams();
            String paramsOld = oldFieldScanEntity.getFAdvancedParams();
            if (isNotEmpty(paramsNew) && !paramsNew.equals(paramsOld)) {
                return true;
            }
            if (isEmpty(paramsNew) && isNotEmpty(paramsOld)) {
                return true;
            }
        }
        return false;
    }

    public static FieldScanEntity makeFieldScanEntity(String fieldName) {
        FieldScanEntity fieldScanEntity = new FieldScanEntity();
        fieldScanEntity.setFId(UUID.randomUUID().toString());
        fieldScanEntity.setFFieldName(fieldName);
        fieldScanEntity.setFFieldType("keyword");
        JSONArray result = new JSONArray();
        JSONObject o1 = new JSONObject();
        o1.put("key", "originFieldType");
        o1.put("value", "Keyword".toLowerCase());
        JSONObject o2 = new JSONObject();
        o2.put("key", "virtualFieldType");
        o2.put("value", "string");
        result.add(o1);
        result.add(o2);
        fieldScanEntity.setFAdvancedParams(result.toJSONString());
        return fieldScanEntity;

    }

    public static boolean isNotEmpty(Object obj) {
        return !isEmpty(obj);
    }

    public static boolean isEmpty(Object obj) {
        if (obj == null) {
            return true;
        } else if (obj instanceof CharSequence) {
            return ((CharSequence) obj).length() == 0;
        } else if (obj instanceof Collection) {
            return ((Collection) obj).isEmpty();
        } else if (obj instanceof Map) {
            return ((Map) obj).isEmpty();
        } else if (obj.getClass().isArray()) {
            return Array.getLength(obj) == 0;
        } else {
            return false;
        }
    }

    public static String getOpenSearchParam(IndicesRecord record) {
        JSONObject health = new JSONObject();
        health.put("key", "health");
        health.put("value", record.health());

        JSONObject status = new JSONObject();
        status.put("key", "status");
        status.put("value", record.status());

        JSONObject uuid = new JSONObject();
        uuid.put("key", "uuid");
        uuid.put("value", record.uuid());

        JSONObject pri = new JSONObject();
        pri.put("key", "pri");
        pri.put("value", record.pri());

        JSONObject rep = new JSONObject();
        rep.put("key", "rep");
        rep.put("value", record.rep());

        JSONObject docsCount = new JSONObject();
        docsCount.put("key", "docs.count");
        docsCount.put("value", record.docsCount());

        JSONObject docsCountDel = new JSONObject();
        docsCountDel.put("key", "docs.deleted");
        docsCountDel.put("value", record.docsDeleted());

        JSONObject storeSize = new JSONObject();
        storeSize.put("key", "store.size");
        storeSize.put("value", record.storeSize());

        JSONObject priStoreSize = new JSONObject();
        priStoreSize.put("key", "pri.store.size");
        priStoreSize.put("value", record.priStoreSize());

        return new JSONArray(health, status, uuid, pri, rep, docsCount, docsCountDel, storeSize, priStoreSize).toJSONString();
    }

    public static final Map<String, ScheduledFuture<?>> SCHEDULE_JOB_MAP = new ConcurrentHashMap<>();
}
