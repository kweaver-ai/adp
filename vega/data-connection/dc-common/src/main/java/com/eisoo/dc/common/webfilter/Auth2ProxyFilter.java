package com.eisoo.dc.common.webfilter;

import com.eisoo.dc.common.auditLog.enums.OperatorType;
import com.eisoo.dc.common.constant.Detail;
import com.eisoo.dc.common.constant.Message;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.constant.Constants;
import com.eisoo.dc.common.constant.Description;
import com.eisoo.dc.common.driven.Hydra;
import com.eisoo.dc.common.vo.Ext;
import com.eisoo.dc.common.vo.IntrospectInfo;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.google.gson.Gson;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.*;

import javax.servlet.*;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.*;
import java.util.*;

@Slf4j
public class Auth2ProxyFilter implements Filter {

    private String  url;
    private boolean authEnabled;
    private Gson gson = new Gson();
    public Auth2ProxyFilter(String  url, boolean authEnabled) {
        this.url=url;
        this.authEnabled=authEnabled;
    }

    @Override
    public void doFilter(ServletRequest servletRequest,
                         ServletResponse servletResponse,
                         FilterChain filterChain) throws IOException, ServletException {
        HttpServletRequest request = (HttpServletRequest) servletRequest;
        HttpServletResponse response = (HttpServletResponse) servletResponse;
        IntrospectInfo introspectInfo = new IntrospectInfo();
        introspectInfo.setExt(new Ext());
        String auth2Token = null;
        // uri
        String uri = request.getRequestURI();
        if (authEnabled && (uri.startsWith("/api/data-connection/v1/datasource")
                || uri.startsWith("/api/data-connection/v1/metadata")
                || uri.startsWith("/api/data-connection/v1/gateway"))) {

            auth2Token = request.getHeader(Constants.HEADER_TOKEN_KEY);
            if(auth2Token == null)
            {
                response.setStatus(401);
                response.setContentType("application/json");
                AiShuException aiShuException = new AiShuException(ErrorCodeEnum.UnauthorizedError, Detail.AUTHORIZATION_NOT_EXIST);
                write(response, null, gson.toJson(aiShuException));
                return ;
            }
            try {
                if (!auth2Token.startsWith("Bearer ")) {
                    //格式错误
                    response.setStatus(400);
                    response.setContentType("application/json");
                    AiShuException aiShuException = new AiShuException(ErrorCodeEnum.UnauthorizedError, Detail.AUTHORIZATION_FORMAT_ERROR);
                    write(response, null, gson.toJson(aiShuException));
                    return;
                }
                String token= auth2Token.substring("Bearer ".length()).trim();
                introspectInfo= Hydra.getIntrospectInfoByToken(url,token,"all");
                if(!introspectInfo.isActive())
                {
                    response.setStatus(401);
                    response.setContentType("application/json");
                    AiShuException aiShuException = new AiShuException(ErrorCodeEnum.UnauthorizedError);
                    write(response, null, gson.toJson(aiShuException));
                    return ;
                }
            } catch (Exception e) {
                response.setStatus(500);
                response.setContentType("application/json");
                AiShuException aiShuException = new AiShuException(ErrorCodeEnum.InternalServerError, Description.HYDRA_SERVICE_ERROR, e.getMessage(), Message.MESSAGE_AUTH_SERVICE_ERROR_SOLUTION);
                write(response, null, gson.toJson(aiShuException));
                return ;
            }
        }
        servletRequest.setAttribute("Authorization", auth2Token);
        servletRequest.setAttribute("introspectInfo", introspectInfo);
        filterChain.doFilter(servletRequest, servletResponse);
    }

    private void write(HttpServletResponse response, HttpHeaders headers, String body) throws IOException {
        if (headers != null) {
            Set<String> headerKeys = headers.keySet();
            for (String key : headerKeys) {
                response.setHeader(key, headers.get(key).toString());
            }
        }
        if (headers!= null && headers.getContentType() != null) {
            response.setContentType(headers.getContentType().toString());
        }
        if (body != null) {
            // 在getWriterz之前执行，否则不生效
            response.setCharacterEncoding("UTF-8");
            PrintWriter writer = response.getWriter();
            writer.write(body);
            writer.flush();
        }
    }


    @Override
    public void destroy() {}

}