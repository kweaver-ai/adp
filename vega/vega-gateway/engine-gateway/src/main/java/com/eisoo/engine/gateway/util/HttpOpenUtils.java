package com.eisoo.engine.gateway.util;

import cn.hutool.core.lang.Validator;
import cn.hutool.http.HttpRequest;
import cn.hutool.http.HttpResponse;
import cn.hutool.http.HttpUtil;
import cn.hutool.http.Method;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.HashMap;

/**
 * @Author zdh
 **/
public class HttpOpenUtils {
    private static final Logger log = LoggerFactory.getLogger(HttpOpenUtils.class);

    //重试次数
    private static final int MAX_RETRIES = 1;

    //超时时间
    private static int connectTimeout = -1;
    private static int readTimeout = -1;

    /**
     * 向指定 URL 发送POST方法的请求
     *
     * @param url   发送请求的 URL
     * @param param 请求参数，请求参数key,value形式。
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendPost(String url, HashMap param, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPost url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.POST, url)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "multipart/form-data")
                        .removeHeader("Content-Encoding")
                        .formStr(param)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPost Exception, url=" + url + ",param=" + param.toString(), e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }


    /**
     * 向指定 URL 发送PUT方法的请求
     *
     * @param url   发送请求的 URL
     * @param param 请求参数，请求参数key,value形式。
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendPut(String url, HashMap param, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPut url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.PUT, url)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "multipart/form-data")
                        .removeHeader("Content-Encoding")
                        .formStr(param)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPut Exception, url=" + url + ",param=" + param.toString(), e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    /**
     * 向指定 URL 发送GET方法的请求
     *
     * @param url         发送请求的 URL
     * @param user 用户
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendGet(String url, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendGet url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.GET, url)
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendGet Exception, url=" + url, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    /**
     * 向指定 URL 发送Delete方法的请求
     *
     * @param url         发送请求的 URL
     * @param user 用户
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendDelete(String url, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendDelete url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.DELETE, url)
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendDelete Exception, url=" + url, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }


    /**
     * 向指定 URL 发送POST方法的请求
     *
     * @param url 发送请求的 URL
     * @param sql 请求参数，请求参数key,value形式。
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendPost(String url, String sql, String user, JSONObject clientInfo, String xPrestoSession) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPost url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                HttpRequest request = HttpUtil.createRequest(Method.POST, url)
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .header("X-Presto-Client-Info", clientInfo.toString())
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .body(sql);
                if (StringUtils.isNotEmpty(xPrestoSession)) {
                    xPrestoSession = xPrestoSession.replaceAll("&", ",");
                    request.header("X-Presto-Session", xPrestoSession);
                }
                execute = request.execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPost Exception, url=" + url + ",param=" + sql, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }


    /**
     * 向指定 URL 发送POST方法的请求
     *
     * @param url 发送请求的 URL
     * @param body 请求参数，请求参数key,value形式。
     * @return 所代表远程资源的响应结果
     */
    public static HttpResInfo sendPost(String url, String body, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPost url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.POST, url)
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .body(body)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPost Exception, url=" + url + ",param=" + body, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    /***
     * 增加一个携带catalog schema table的post请求用于触发引擎的缓存
     * @param url
     * @param sql
     * @param user
     * @param catalog
     * @param schema
     * @param table
     * @return
     */
    public static HttpResInfo sendPostWithTabInfo(String url, String sql, String user,String catalog,String schema,String table) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPost url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                JSONObject jsonObject = new JSONObject();
                jsonObject.put("catalog",catalog);
                jsonObject.put("schema",schema);
                jsonObject.put("table",table);
                execute = HttpUtil.createRequest(Method.POST, url)
                                  .contentType(Constants.UTF8)
                                  .setConnectionTimeout(connectTimeout)
                                  .setReadTimeout(readTimeout)
                                  .header("X-Presto-User", user)
                                  .header("Content-Type", "application/json")
                                  .header("Accept", "application/json")
                                  .header("Case-Query-Info", jsonObject.toString())
                                  .removeHeader("Content-Encoding")
                                  .body(sql)
                                  .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPost Exception, url=" + url + ",param=" + sql, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }
    public static HttpResInfo sendPost(String url, String sql, String topic, String user) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPost url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        HashMap param = new HashMap<>();
        param.put("statement", sql);
        param.put("topic", topic);
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.POST, url)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "multipart/form-data")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .formStr(param)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPost Exception, url=" + url + ",param=" + sql, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    public static HttpResInfo sendPostWithParams(String url, String type, String sql, String user) throws URISyntaxException {
        HashMap params =new HashMap<>();
        params.put("type",type);
        // 构建带参数的完整URL
        URI uri = new URI(url + "?type=" + type);

        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.POST, uri.toString())
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .body(sql)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPostWithParams Exception, url=" + uri + ",param=" + sql, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        if (execute != null) {
            execute.close();
        }
        return execute != null ? new HttpResInfo(execute.getStatus(), execute.body()) : null;
    }

    /**
     * Post请求设置连接超时时间、指定user
     *
     * @param url
     * @param sql
     * @param user
     * @param connectionTimeOut
     * @return
     */
    public static HttpResInfo sendPostOlk(String url, String sql, String user, int connectionTimeOut) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendPostOlk url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.POST, url)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .header("Accept", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .body(sql)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendPostOlk Exception, url:{},param:{}", url, sql, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    /**
     * Get请求设置连接超时时间、指定user
     *
     * @param url
     * @param user
     * @param timeOut
     * @return
     */
    public static HttpResInfo sendGetOlk(String url, String user, int timeOut) {
        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendGetOlk url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.GET, url)
                        .header("X-Presto-User", user)
                        .header("Content-Type", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendGetOlk Exception, url:{}", url, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }

    /**
     * Get请求设置连接指定user
     *
     * @param url
     * @param user
     * @return
     */
    public static HttpResInfo sendGetOlk(String url, String user) {
        // 输出当前线程 ID
        log.info("Current thread ID: " + Thread.currentThread().getId());

        if (!Validator.isUrl(url)) {
            log.error("调用HttpOpenUtils.sendGetOlk url error=" + url);
            throw new AiShuException(ErrorCodeEnum.InternalError, String.format(Detail.URL_ERROR, url), Message.MESSAGE_INTERNAL_ERROR);
        }
        HttpResponse execute = null;
        for (int retryCount = 0; retryCount < MAX_RETRIES; retryCount++) {
            try {
                execute = HttpUtil.createRequest(Method.GET, url)
                        .charset(Constants.UTF8)
                        .setConnectionTimeout(connectTimeout)
                        .setReadTimeout(readTimeout)
                        .header("X-Presto-User", user)
                        .header("Accept", "application/json")
                        .header("Content-Type", "application/json")
                        .removeHeader("Content-Encoding")
                        .removeHeader("Accept-Encoding")
                        .execute();
                break;
            } catch (Exception e) {
                log.error("调用HttpOpenUtils.sendGetOlk Exception, url:{}", url, e);
                if (retryCount == MAX_RETRIES - 1) {
                    throw e;
                }
            }
        }
        execute.close();
        return new HttpResInfo(execute.getStatus(), execute.body());
    }
}
