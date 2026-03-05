package com.eisoo.engine.gateway.controller;

import cn.hutool.core.io.IORuntimeException;
import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.JSONArray;
import com.alibaba.fastjson2.JSONObject;
import com.eisoo.engine.gateway.domain.vo.HttpResInfo;
import com.eisoo.engine.gateway.util.HttpOpenUtils;
import com.eisoo.engine.metadata.mapper.ClientIdMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.util.AFUtil;
import com.eisoo.engine.utils.util.HttpsUtil;
import com.eisoo.engine.utils.vo.AuthTokenInfo;
import com.eisoo.engine.utils.vo.RowColumnRuleVo;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.junit.jupiter.api.DisplayName;
import org.junit.runner.RunWith;
import org.mockito.MockedStatic;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.*;
import org.springframework.test.context.junit4.SpringRunner;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockMvcRequestBuilders;
import org.springframework.test.web.servlet.result.MockMvcResultMatchers;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.context.WebApplicationContext;
import com.eisoo.engine.utils.common.HttpStatus;

import java.util.ArrayList;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@SpringBootTest
@AutoConfigureMockMvc
@RunWith(SpringRunner.class)
@DisplayName("数据查询单元测试")
public class FetchControllerTest {
    @Autowired
    private WebApplicationContext context;

    private MockMvc mockMvc;

    private MockedStatic<HttpOpenUtils> httpClient;

    private MockedStatic<HttpsUtil> httpsUtil;

    private MockedStatic<RequestEntity> requestEntity;

    @Value(value = "${auth.clientid-token}")
    private String clientIdTokenUrl;

    @Value(value = "${auth.row-column-rule}")
    private String rowColumnRuleUrl;

    @Value(value = "${auth.register-clientid-url}")
    private String registerClientIdUrl;

    @MockBean
    private ClientIdMapper clientIdMapper;

    @MockBean
    private AFUtil afUtil;

    @MockBean
    private RequestEntity.BodyBuilder bodyBuilder;

    @MockBean
    private RestTemplate restTemplate;

    @Before
    public void setUp() throws Exception {
        mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
        this.httpClient = mockStatic(HttpOpenUtils.class);
        this.httpsUtil = mockStatic(HttpsUtil.class);
        this.requestEntity = mockStatic(RequestEntity.class);
    }

    @After
    public void teardown() {
        this.httpClient.close();
        this.httpsUtil.close();
        this.requestEntity.close();
    }

    @Test
    public void fetchTest_ok() throws Exception {

        httpClient.when(()->HttpOpenUtils.sendGet(anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"connector.name\":\"vdm\"}\n"));

        when(clientIdMapper.selectById(any())).thenReturn(null);
        when(clientIdMapper.insert(any())).thenReturn(1);

        //mock afUtil.registerClientId
        JSONObject requestBody = new JSONObject();
        JSONArray grantTypes = new JSONArray();
        grantTypes.add("urn:ietf:params:oauth:grant-type:jwt-bearer");
        grantTypes.add("client_credentials");
        JSONArray responseTypes = new JSONArray();
        responseTypes.add("token");
        JSONArray redirectUris = new JSONArray();
        redirectUris.add("https://127.0.0.1:9010/callback");
        JSONArray postLogoutRedirectUris = new JSONArray();
        postLogoutRedirectUris.add("https://127.0.0.1:9010/successful-logout");
        requestBody.put("client_name", Constants.OAUTH_CLIENT_NAME);
        requestBody.put("grant_types", grantTypes);
        requestBody.put("response_types", responseTypes);
        requestBody.put("scope", "all");
        requestBody.put("redirect_uris", redirectUris);
        requestBody.put("post_logout_redirect_uris", postLogoutRedirectUris);
        HttpEntity<Object> entity = new HttpEntity<>(requestBody);
        ResponseEntity<String> responseStrEntity = new ResponseEntity<>("{\"client_id\":\"xxx\",\"client_secret\":\"xxx\"}", org.springframework.http.HttpStatus.valueOf(HttpStatus.SUCCESS));
        when(restTemplate.postForEntity(registerClientIdUrl, entity, String.class)).thenReturn(responseStrEntity);
        when(afUtil.registerClientId(any(),any())).thenReturn(JSON.parseObject(responseStrEntity.getBody()));

        when(clientIdMapper.updateById(any())).thenReturn(1);



        //mock afUtil.login
        AuthTokenInfo authTokenInfo = new AuthTokenInfo();
        authTokenInfo.setStatusCode(200);
        ResponseEntity<AuthTokenInfo> responseEntity = new ResponseEntity<>(authTokenInfo, org.springframework.http.HttpStatus.valueOf(HttpStatus.SUCCESS));
        requestEntity.when(()->RequestEntity.post(anyString())).thenReturn(bodyBuilder);
        when(bodyBuilder.contentType(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.accept(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.acceptCharset(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.body(any())).thenReturn(null);
        when(restTemplate.postForEntity(clientIdTokenUrl, null, AuthTokenInfo.class)).thenReturn(responseEntity);
        httpsUtil.when(()->HttpsUtil.getRestTemplateAndSetBaisAuth(anyString(),anyString())).thenReturn(restTemplate);
        when(afUtil.login(anyString(),anyString(),anyString())).thenReturn(authTokenInfo);

        //mock afUtil.getRowColumnRule
        httpsUtil.when(HttpsUtil::getRestTemplate).thenReturn(restTemplate);
        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", null);
        HttpEntity<MultiValueMap<String, Object>> httpEntity = new HttpEntity<>(headers);
        RowColumnRuleVo rowColumnRuleVo = new RowColumnRuleVo();
        rowColumnRuleVo.setStatusCode(200);
        ArrayList<RowColumnRuleVo.Entry> entries = new ArrayList<>();
        RowColumnRuleVo.Entry entry = new RowColumnRuleVo.Entry();
        ArrayList<String> columns = new ArrayList<>();
        columns.add("test");
        entry.setColumns(columns);
        entry.setRowRule("xxx");
        entries.add(entry);
        rowColumnRuleVo.setEntries(entries);
        rowColumnRuleVo.setTargetSql("select * from mysql1.test.test1 limit 10000");
        ResponseEntity<RowColumnRuleVo> exchangeResponseEntity = new ResponseEntity<>(rowColumnRuleVo, org.springframework.http.HttpStatus.valueOf(HttpStatus.SUCCESS));
        when(restTemplate.exchange(rowColumnRuleUrl+ "?logic_view_name=mysql1.test.test1&user_id=admin", HttpMethod.GET, httpEntity, RowColumnRuleVo.class)).thenReturn(exchangeResponseEntity);
        when(afUtil.getRowColumnRule(anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(rowColumnRuleVo);

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FINISHED\"},\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}],\"data\":[[\"jack\",2],[\"tom\",2]]}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/fetch")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .header("X-Presto-Session","xxx=xxx")
                        .param("user_id","admin")
                        .param("type","1")
                        .content("select * from mysql1.test.test1 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().isOk());
    }

    @Test
    public void fetchTest_headerError() throws Exception {
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/fetch")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void fetchTest_getStatementError() throws Exception {

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.BAD_REQUEST,"Query not found"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/fetch")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .content("select * from xxx.xxx.xxx"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
    }

    @Test
    public void fetchTest_queuedError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet(any(),any())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"xxx\",\"errorName\": \"xxx\",\"semanticErrorName\": \"MISSING_ATTRIBUTE\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/fetch")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .header("X-Presto-Session","xxx=xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 10000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void fetchTest_executingError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"xxx\",\"errorName\": \"xxx\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/fetch")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .header("X-Presto-Session","xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void nextFetchTest_ok() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendGet(anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"data\": [[{\"xxx\": \"xxx\"}]],\"nextUri\": \"xxx\",\"columns\": [{\"type\": \"numeric_v1\",\"name\": \"numeric_v1\"}]}"));
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/statement/executing/{queryId}/{slug}/{token}","xxx","xxx",0L)
                        .header("X-Presto-User","admin"))
                .andExpect(MockMvcResultMatchers.status().isOk());
    }

    @Test
    public void sampleTest_ok() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendGet(any(),any()))
//                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"connector.name\":\"vdm\"}\n"))
                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"))
                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FINISHED\"},\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}],\"data\":[[\"jack\",2],[\"tom\",2]],\"total_count\":2}"));

        when(clientIdMapper.selectById(any())).thenReturn(null);
        when(clientIdMapper.insert(any())).thenReturn(1);

        //mock afUtil.registerClientId
        JSONObject requestBody = new JSONObject();
        JSONArray grantTypes = new JSONArray();
        grantTypes.add("urn:ietf:params:oauth:grant-type:jwt-bearer");
        grantTypes.add("client_credentials");
        JSONArray responseTypes = new JSONArray();
        responseTypes.add("token");
        JSONArray redirectUris = new JSONArray();
        redirectUris.add("https://127.0.0.1:9010/callback");
        JSONArray postLogoutRedirectUris = new JSONArray();
        postLogoutRedirectUris.add("https://127.0.0.1:9010/successful-logout");
        requestBody.put("client_name", Constants.OAUTH_CLIENT_NAME);
        requestBody.put("grant_types", grantTypes);
        requestBody.put("response_types", responseTypes);
        requestBody.put("scope", "all");
        requestBody.put("redirect_uris", redirectUris);
        requestBody.put("post_logout_redirect_uris", postLogoutRedirectUris);
        HttpEntity<Object> entity = new HttpEntity<>(requestBody);
        ResponseEntity<String> responseStrEntity = new ResponseEntity<>("{\"client_id\":\"xxx\",\"client_secret\":\"xxx\"}", org.springframework.http.HttpStatus.valueOf(HttpStatus.SUCCESS));
        when(restTemplate.postForEntity(registerClientIdUrl, entity, String.class)).thenReturn(responseStrEntity);
        when(afUtil.registerClientId(any(),any())).thenReturn(JSON.parseObject(responseStrEntity.getBody()));

        when(clientIdMapper.updateById(any())).thenReturn(1);


        //mock afUtil.login
        AuthTokenInfo authTokenInfo = new AuthTokenInfo();
        authTokenInfo.setStatusCode(HttpStatus.BAD_REQUEST);
        authTokenInfo.setErrorMsg("not allowed to use authorization grant 'client_credentials'");
        ResponseEntity<AuthTokenInfo> responseEntity = new ResponseEntity<>(authTokenInfo, org.springframework.http.HttpStatus.valueOf(HttpStatus.BAD_REQUEST));
        requestEntity.when(()->RequestEntity.post(anyString())).thenReturn(bodyBuilder);
        when(bodyBuilder.contentType(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.accept(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.acceptCharset(any())).thenReturn(bodyBuilder);
        when(bodyBuilder.body(any())).thenReturn(null);
        when(restTemplate.postForEntity(clientIdTokenUrl, null, AuthTokenInfo.class)).thenReturn(responseEntity);
        httpsUtil.when(()->HttpsUtil.getRestTemplateAndSetBaisAuth(anyString(),anyString())).thenReturn(restTemplate);
        when(clientIdMapper.deleteById(any())).thenReturn(1);
        when(afUtil.login(anyString(),anyString(),anyString())).thenReturn(authTokenInfo);


        //mock afUtil.getRowColumnRule
        httpsUtil.when(HttpsUtil::getRestTemplate).thenReturn(restTemplate);
        HttpHeaders headers = new HttpHeaders();
        headers.set("Authorization", null);
        HttpEntity<MultiValueMap<String, Object>> httpEntity = new HttpEntity<>(headers);
        RowColumnRuleVo rowColumnRuleVo = new RowColumnRuleVo();
        rowColumnRuleVo.setStatusCode(200);
        ArrayList<RowColumnRuleVo.Entry> entries = new ArrayList<>();
        RowColumnRuleVo.Entry entry = new RowColumnRuleVo.Entry();
        ArrayList<String> columns = new ArrayList<>();
        columns.add("col1");
        entry.setColumns(columns);
        entries.add(entry);
        rowColumnRuleVo.setEntries(entries);
        rowColumnRuleVo.setTargetSql("select * from xxx.xxx.xxx limit 10000");
        ResponseEntity<RowColumnRuleVo> exchangeResponseEntity = new ResponseEntity<>(rowColumnRuleVo, org.springframework.http.HttpStatus.valueOf(HttpStatus.SUCCESS));
        when(restTemplate.exchange(rowColumnRuleUrl+ "?logic_view_name=xxx.xxx.xxx&user_id=admin", HttpMethod.GET, httpEntity, RowColumnRuleVo.class)).thenReturn(exchangeResponseEntity);
        when(afUtil.getRowColumnRule(anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(rowColumnRuleVo);


        httpClient.when(()->HttpOpenUtils.sendPostWithTabInfo(anyString(),anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/preview/{catalog}/{schema}/{table}","xxx","xxx","xxx")
                        .header("X-Presto-User","admin")
                        .param("limit","10000")
                        .param("user_id","admin")
                        .param("columns","col1"))
                .andExpect(MockMvcResultMatchers.status().isOk());
    }
    @Test
    public void sampleTest_headerError() throws Exception {
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/preview/{catalog}/{schema}/{table}","xxx","xxx","xxx")
                        .param("limit","10000")
                        .param("columns","col1"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void sampleTest_getStatementError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPostWithTabInfo(anyString(),anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.BAD_REQUEST,"Query not found"));
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/preview/{catalog}/{schema}/{table}","xxx","xxx","xxx")
                        .header("X-Presto-User","admin")
                        .param("limit","10000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
    }

    @Test
    public void sampleTest_queuedError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPostWithTabInfo(anyString(),anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet(any(),any())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"Schema xxx does not exist\",\"errorName\": \"xxx\",\"semanticErrorName\": \"VIEW_ANALYSIS_ERROR\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/preview/{catalog}/{schema}/{table}","xxx","xxx","xxx")
                        .header("X-Presto-User","admin")
                        .param("limit","10000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void sampleTest_executingError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPostWithTabInfo(anyString(),anyString(),anyString(),anyString(),anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"xxx\",\"errorName\": \"xxx\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.get("/api/virtual_engine_service/v1/preview/{catalog}/{schema}/{table}","xxx","xxx","xxx")
                        .header("X-Presto-User","admin")
                        .param("limit","10000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void downloadTest_ok() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendGet(any(),any()))
                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"connection-password\":\"xxx\",\"connector.name\":\"mysql\",\"connection-url\":\"jdbc:mysql://127.0.0.1:3306\",\"connection-user\":\"root\",\"jdbc.pushdown-module\":\"FULL_PUSHDOWN\"}\n"))
                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FINISHED\"},\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}]}"));

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/download")
                        .contentType(MediaType.APPLICATION_JSON)
                        .header("X-Presto-User",Constants.DEFAULT_ASYNC_TASK_USER)
                        .param("user_id","admin")
                        .content("{\n" +
                                "    \"catalog\": \"xxx\",\n" +
                                "    \"schema\": \"xxx\",\n" +
                                "    \"table\": \"xxx\",\n" +
                                "    \"columns\": \"col1\",\n" +
                                "    \"row_rules\": \"col1 like '%%'\",\n" +
                                "    \"order_by\": \"col1\",\n" +
                                "    \"offset\": 0,\n" +
                                "    \"limit\": 10,\n" +
                                "    \"action\": \"download\"\n" +
                                "}"))
                .andExpect(MockMvcResultMatchers.status().isOk());
    }
//    @Test
//    public void downloadTest_headerError() throws Exception {
//        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/download")
//                        .contentType(MediaType.APPLICATION_JSON)
//                        .content("{\n" +
//                                "    \"catalog\": \"xxx\",\n" +
//                                "    \"schema\": \"xxx\",\n" +
//                                "    \"table\": \"xxx\",\n" +
//                                "    \"columns\": \"col1\",\n" +
//                                "    \"row_rules\": \"col1 like '%%'\",\n" +
//                                "    \"order_by\": \"col1\",\n" +
//                                "    \"offset\": 0,\n" +
//                                "    \"limit\": 10,\n" +
//                                "    \"action\": \"download\"\n" +
//                                "}"))
//                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
//    }

    @Test
    public void downloadTest_getStatementError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.BAD_REQUEST,"Query not found"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/download")
                        .contentType(MediaType.APPLICATION_JSON)
                        .header("X-Presto-User",Constants.DEFAULT_ASYNC_TASK_USER)
                        .content("{\n" +
                                "    \"catalog\": \"mysql1\",\n" +
                                "    \"schema\": \"anydata\",\n" +
                                "    \"table\": \"test\",\n" +
                                "    \"columns\": \"name\",\n" +
                                "    \"rowRules\": \"\",\n" +
                                "    \"orderBy\": \"\",\n" +
                                "    \"offset\": 0,\n" +
                                "    \"limit\": 10,\n" +
                                "    \"action\": \"download\"\n" +
                                "}"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
    }

    @Test
    public void downloadTest_queuedError() throws Exception {

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx",Constants.DEFAULT_ASYNC_TASK_USER)).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\":{\"state\":\"FAILED\"},\"error\":{\"message\":\"xxx\",\"errorName\": \"xxx\",\"semanticErrorName\": \"xxx\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/download")
                        .contentType(MediaType.APPLICATION_JSON)
                        .header("X-Presto-User",Constants.DEFAULT_ASYNC_TASK_USER)
                        .content("{\n" +
                                "    \"catalog\": \"mysql1\",\n" +
                                "    \"schema\": \"anydata\",\n" +
                                "    \"table\": \"test\",\n" +
                                "    \"columns\": \"name\",\n" +
                                "    \"rowRules\": \"\",\n" +
                                "    \"orderBy\": \"\",\n" +
                                "    \"offset\": 0,\n" +
                                "    \"limit\": 10,\n" +
                                "    \"action\": \"download\"\n" +
                                "}"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void downloadTest_executingError() throws Exception {

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx",Constants.DEFAULT_ASYNC_TASK_USER)).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1",Constants.DEFAULT_ASYNC_TASK_USER)).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\":{\"state\":\"FAILED\"},\"error\":{\"message\":\"xxx\",\"errorName\": \"xxx\",\"semanticErrorName\": \"MISSING_CATALOG\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/download")
                        .contentType(MediaType.APPLICATION_JSON)
                        .header("X-Presto-User",Constants.DEFAULT_ASYNC_TASK_USER)
                        .content("{\n" +
                                "    \"catalog\": \"mysql1\",\n" +
                                "    \"schema\": \"anydata\",\n" +
                                "    \"table\": \"test\",\n" +
                                "    \"columns\": \"name\",\n" +
                                "    \"rowRules\": \"\",\n" +
                                "    \"orderBy\": \"\",\n" +
                                "    \"offset\": 0,\n" +
                                "    \"limit\": 10,\n" +
                                "    \"action\": \"download\"\n" +
                                "}"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void queryTest_ok() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendGet(anyString(),anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"connection-password\":\"xxx\",\"connector.name\":\"mysql\",\"connection-url\":\"jdbc:mysql://127.0.0.1:3306\",\"connection-user\":\"root\",\"jdbc.pushdown-module\":\"FULL_PUSHDOWN\"}\n"));

        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        //1
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/0\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/0","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"RUNNING\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\",\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}],\"data\":[[\"jack\",2],[\"tom\",2]]}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","admin")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FINISHED\"},\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}]}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .param("user_id","admin")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().isOk());

        //2
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FINISHED\"},\"columns\": [{\"name\": \"xxx\",\"type\": \"xxx\"},{\"name\": \"total\",\"type\": \"bigint\"}],\"data\":[[\"jack\",2],[\"tom\",2]]}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .param("user_id","xxx")
                        .content("select * from xxx.xxx.xxx limit 1000"))
//                .andDo(MockMvcResultHandlers.print())
                .andExpect(MockMvcResultMatchers.status().isOk());
    }

    @Test
    public void queryTest_headerError() throws Exception {
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void queryTest_paramError() throws Exception {

        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .content("delete from xxx.xxx.xxx"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));

        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 10000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void queryTest_getStatementError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.BAD_REQUEST,"Query not found"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .content("select * from xxx.xxx.xxx"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
    }

    @Test
    public void queryTest_queuedError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet(any(),any())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"Table xxx does not exist\",\"errorName\": \"xxx\",\"semanticErrorName\": \"VIEW_ANALYSIS_ERROR\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","admin")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void queryTest_executingError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString())).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"QUEUED\"},\"nextUri\": \"http://127.0.0.1:8090/v1/statement/executing/xxx/1\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("http://127.0.0.1:8090/v1/statement/executing/xxx/1","xxx")).thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"stats\": {\"state\": \"FAILED\"},\"error\": {\"message\": \"xxx\",\"errorName\": \"xxx\"}}"));
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.BAD_REQUEST));
    }

    @Test
    public void queryTest_openlookengError() throws Exception {
        httpClient.when(()->HttpOpenUtils.sendPost(anyString(),anyString(),anyString(), (JSONObject) any(), anyString()))
                .thenThrow(new IORuntimeException(" SocketTimeoutException: connect timed out"))
                .thenReturn(new HttpResInfo(HttpStatus.SUCCESS,"{\"nextUri\": \"xxx\"}"));
        httpClient.when(()->HttpOpenUtils.sendGet("xxx","xxx"))
                .thenThrow(new IORuntimeException(" SocketTimeoutException: connect timed out"));
        //1
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
        //2
        mockMvc.perform(MockMvcRequestBuilders.post("/api/virtual_engine_service/v1/query")
                        .contentType(MediaType.TEXT_PLAIN_VALUE)
                        .header("X-Presto-User","xxx")
                        .content("select * from xxx.xxx.xxx offset 0 limit 1000"))
                .andExpect(MockMvcResultMatchers.status().is(HttpStatus.ERROR));
    }
}
