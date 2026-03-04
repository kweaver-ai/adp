package com.eisoo.dc.datasource.config;

import com.eisoo.dc.common.driven.service.ServiceEndpoints;
import com.eisoo.dc.common.webfilter.ArTraceFilter;
import com.eisoo.dc.common.webfilter.Auth2ProxyFilter;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.Ordered;

/**
 *
 */
@Configuration
public class FilterConfig {

    @Autowired
    private ServiceEndpoints serviceEndpoints;
    @Autowired
    AnyRobotConfigruation anyRobotConfigruation;

    @Value(value = "#{systemEnvironment['AUTH_ENABLED'] ?: false}")
    private boolean authEnabled;

    @Bean
    public FilterRegistrationBean<ArTraceFilter> arTraceFilter() {
        FilterRegistrationBean<ArTraceFilter> bean = new FilterRegistrationBean<>();
        String traceEndpointUrl = anyRobotConfigruation.getTraceEnabled() == false ? null : anyRobotConfigruation.getTraceEndpointUrl();
        bean.setFilter(new ArTraceFilter(traceEndpointUrl, anyRobotConfigruation.getServiceName(), anyRobotConfigruation.getServiceVersion()));
        bean.setName("ArTraceFilter");
        bean.setOrder(-1);
        return bean;

    }

    @Bean
    public FilterRegistrationBean<Auth2ProxyFilter> auth2ProxyFilterRegistration() {
        FilterRegistrationBean<Auth2ProxyFilter> bean = new FilterRegistrationBean<>(
                new Auth2ProxyFilter(serviceEndpoints.getHydraAdmin(), authEnabled));
        bean.setName("Auth2ProxyFilter");
        bean.setOrder(Ordered.HIGHEST_PRECEDENCE);
        return bean;
    }
}
