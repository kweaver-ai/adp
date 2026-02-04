package com.eisoo.engine.gateway.service.impl;

import cn.hutool.core.date.StopWatch;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.eisoo.engine.gateway.common.CatalogConstant;
import com.eisoo.engine.gateway.common.QueryConstant;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.utils.common.Detail;
import com.eisoo.engine.utils.common.HttpStatus;
import com.eisoo.engine.utils.common.Message;
import com.eisoo.engine.utils.enums.ErrorCodeEnum;
import com.eisoo.engine.utils.exception.AiShuException;
import com.eisoo.engine.utils.util.StringUtils;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.jayway.jsonpath.JsonPath;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantLock;

/**
 * @Author zdh
 **/
@Service
public class OlkCollectionServiceImpl {

    private static final Logger log = LoggerFactory.getLogger(OlkCollectionServiceImpl.class);

    @Value(value = "${openlookeng.url}")
    private String openlookengUrl;
    @Value(value = "${openlookeng.connect-timeout}")
    private int connectTimeout;
    @Autowired
    TaskServiceImpl taskServiceImpl;
    private final ReentrantLock lock = new ReentrantLock();


    protected String getStatement(String statement, String user) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_STATEMENT;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendPostOlk(urlOpen, statement, user,connectTimeout);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError, e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http getStatement 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getStatement 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

    protected String getStatementURL(String statement, String user) {
        String urlOpen = openlookengUrl + QueryConstant.VIRTUAL_V1_STATEMENT_V1;
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendPostOlk(urlOpen, statement, user,connectTimeout);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http getStatementURL 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http getStatementURL 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }
    /**
     * 执行接口拿取
     *
     * @param json
     * @return
     */
    protected String getnextUri(String json) {
        return getJsonPathToStr(json, "$.nextUri");
    }

    protected String getQueryId(String json) {
        return getJsonPathToStr(json, "$.id");
    }

    protected String getRowCount(String json) {
        return getJsonPathToStr(json, "$.updateCount");
    }

    /**
     * 执行状态
     *
     * @param json
     * @return
     * @throws JsonProcessingException
     */
    protected String checkState(String json) {
        return getJsonPathToStr(json, "$.stats.state");
    }

    protected String getErrorMessage(String json) {
        return getJsonPathToStr(json, "$.error.message");
    }

    protected String getErrorName(String json) {
        return getJsonPathToStr(json, "$.error.errorName");
    }


    /**
     * JsonPath获取路径值
     *
     * @param json
     * @param path
     * @return
     */
    protected String getJsonPathToStr(String json, String path) {
        String result = null;
        try {
            if (StringUtils.equalsAnyIgnoreCase(json, CatalogConstant.ERROR_MSG)) {
                return null;
            }
            JSONObject jsonObject = JSONUtil.parseObj(json);
            Object read = JsonPath.read(jsonObject, path);
            result = String.valueOf(read);
        } catch (Exception e) {
            log.error("JsonPath read error,path:{}", path, e);
        }
        return result;
    }


    /**
     * execute执行
     *
     * @param nextUri
     * @param user
     * @return
     */
    protected String execute(String nextUri, String user) {
        if (StringUtils.equalsAnyIgnoreCase(nextUri, CatalogConstant.ERROR_MSG)) {
            return null;
        }
        lock.lock();
        try {
            String result = executeStage(nextUri, user);
            String msg = checkState(result);
            if (StringUtils.endsWithIgnoreCase(msg, "FAILED") || StringUtils.endsWithIgnoreCase(msg, "FINISHED")) {
                return result;
            }
            //解析json，获取结果集
            //获取下一个nextUri
            String url = getnextUri(result);
            if (StringUtils.isBlank(url)) {
                return result;
            } else {
                return execute(url, user);
            }
        } finally {
            lock.unlock();
        }
    }

//    protected String executeWithQueryId(String nextUri, String user, String queryId, int readTimeOut) {
//        if (StringUtils.equalsAnyIgnoreCase(nextUri, CatalogConstant.ERROR_MSG)) {
//            return null;
//        }
//        lock.lock();
//        try {
//            String result = executeStageWithQueryId(nextUri, user, queryId, readTimeOut);
//            if (null == result) {
//                return null;
//            }
//            String msg = checkState(result);
//            if (StringUtils.endsWithIgnoreCase(msg, "FAILED") || StringUtils.endsWithIgnoreCase(msg, "FINISHED")) {
//                return result;
//            }
//            //解析json，获取结果集
//            //获取下一个nextUri
//            String url = getnextUri(result);
//            if (StringUtils.isBlank(url)) {
//                return result;
//            } else {
//                return executeWithQueryId(url, user, queryId, readTimeOut);
//            }
//        } finally {
//            lock.unlock();
//        }
//    }

    /**
     * openlookeng接口请求
     *
     * @param nextUri
     * @param user
     * @return
     */
    protected String executeStage(String nextUri, String user) {
        StopWatch stopWatch = new StopWatch();
        stopWatch.start();
        HttpResInfo result;
        try{
            result = HttpOpenUtils.sendGetOlk(nextUri, user,connectTimeout);
        }catch (AiShuException e){
            throw e;
        }catch (Exception e){
            throw new AiShuException(ErrorCodeEnum.ETrinoError,e.getMessage(),Message.MESSAGE_ETRINO_SOLUTION);
        }
        stopWatch.stop();

        if (result.getHttpStatus()>=HttpStatus.BAD_REQUEST){
            log.error("Http executeStage 请求失败: httpStatus={}, result={}, 耗时={}ms",
            result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
            throw new AiShuException(ErrorCodeEnum.OpenLooKengError,result.getResult(),Message.MESSAGE_OPENLOOKENG_ERROR);
        }
        log.info("Http executeStage 请求成功: httpStatus={}, result={}, 耗时={}ms",
        result.getHttpStatus(), result.getResult(), stopWatch.getTotal(TimeUnit.MILLISECONDS));
        return result.getResult();
    }

//    protected String executeStageWithQueryId(String nextUri, String user, String queryId, int readTimeOut) {
//        StopWatch stopWatch = new StopWatch();
//        stopWatch.start();
//        String result = null;
//        try {
//            result = HttpOpenUtils.sendGetOlkWithTimeOut(nextUri, user, connectTimeout, readTimeOut);
//        } catch (Exception e) {
//            log.error("executeStageWithQueryId执行报错！现在手动 taskServiceImpl.DeleteTaskReq掉任务！queryId={}", queryId);
//            taskServiceImpl.deleteTaskReq(queryId);
//        } finally {
//            stopWatch.stop();
////            log.info("HttpOpenUtils.sendGetOlk nextUri:{}, 请求时长为：{}ms", nextUri, stopWatch.getTotal(TimeUnit.MILLISECONDS));
//        }
//        return result;
//    }
}
