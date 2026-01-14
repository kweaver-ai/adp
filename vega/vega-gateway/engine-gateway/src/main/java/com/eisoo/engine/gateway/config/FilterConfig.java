package com.eisoo.engine.gateway.config;

import com.eisoo.engine.gateway.configuration.AnyRobotConfigruation;
import com.eisoo.engine.utils.webfilter.ArTraceFilter;
import com.eisoo.engine.gateway.proxyfilter.Auth2ProxyFilter;
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
    @Value(value = "${af-auth.token-introspect}")
    private String tokenInstropect;
    @Value(value = "${af-auth.is-open}")
    private boolean isOpen;



    @Autowired
    AnyRobotConfigruation anyRobotConfigruation;

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
                new Auth2ProxyFilter( isOpen,tokenInstropect));
        bean.setName("Auth2ProxyFilter");
        bean.setOrder(Ordered.HIGHEST_PRECEDENCE);
        return bean;
    }
}
