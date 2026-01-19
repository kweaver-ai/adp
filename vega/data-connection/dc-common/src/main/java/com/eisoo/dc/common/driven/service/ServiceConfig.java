package com.eisoo.dc.common.driven.service;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1Service;
import io.kubernetes.client.openapi.models.V1ServiceList;
import io.kubernetes.client.util.Config;
import lombok.Data;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import javax.annotation.PostConstruct;
import java.io.IOException;
import java.util.Objects;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

@Configuration
@Data
public class ServiceConfig {

    private static final Logger log = LoggerFactory.getLogger(ServiceConfig.class);


    @Value("${services.is-local}")
    private boolean isLocal;

    @Value("${services.vega-calculate-coordinator}")
    private String vegaCalculateCoordinator;

    @Value("${services.hydra-admin}")
    private String hydraAdmin;

    @Value("${services.user-management-private}")
    private String userManagementPrivate;

    @Value("${services.authorization-private}")
    private String authorizationPrivate;

    @Value("${services.efast-public}")
    private String efastPublic;

    @Value("${services.efast-private}")
    private String efastPrivate;

    @Value("${services.vega-gateway}")
    private String vegaGateway;


    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(1);
    private final AtomicReference<String> vegaCalculateCoordinatorFQDN = new AtomicReference<>();
    private String hydraAdminFQDN;
    private String userManagementPrivateFQDN;
    private String authorizationPrivateFQDN;
    private String efastPublicFQDN;
    private String efastPrivateFQDN;

    @PostConstruct
    public void init() throws IOException {
        if (!isLocal) {
            ApiClient client = Config.fromCluster();
            CoreV1Api api = new CoreV1Api(client);

            // 初始化获取所有服务地址
            hydraAdminFQDN = getServiceFQDN(api, hydraAdmin);
            userManagementPrivateFQDN = getServiceFQDN(api, userManagementPrivate);
            authorizationPrivateFQDN = getServiceFQDN(api, authorizationPrivate);
            efastPublicFQDN = getServiceFQDN(api, efastPublic);
            efastPrivateFQDN = getServiceFQDN(api, efastPrivate);

            // 定时更新vegaCalculateCoordinator地址
            updateVegaCalculateCoordinator(api);
            scheduler.scheduleAtFixedRate(() -> updateVegaCalculateCoordinator(api),
                    1, 1, TimeUnit.MINUTES); // 每1分钟更新一次
        }
    }

    private void updateVegaCalculateCoordinator(CoreV1Api api) {
        try {
            String newFQDN = getServiceFQDN(api, vegaCalculateCoordinator);
            vegaCalculateCoordinatorFQDN.set(newFQDN);
        } catch (Exception e) {
            vegaCalculateCoordinatorFQDN.set(null);
        }
    }


    @Bean
    public ServiceEndpoints serviceEndpoints() {
        if (isLocal){
            return new ServiceEndpoints(
                    () -> vegaCalculateCoordinator,
                    () -> hydraAdmin,
                    () -> userManagementPrivate,
                    () -> authorizationPrivate,
                    () -> efastPublic,
                    () -> efastPrivate
            );
        }else{
            return new ServiceEndpoints(
                    () -> vegaCalculateCoordinatorFQDN.get(),
                    () -> hydraAdminFQDN,
                    () -> userManagementPrivateFQDN,
                    () -> authorizationPrivateFQDN,
                    () -> efastPublicFQDN,
                    () -> efastPrivateFQDN
            );
        }
    }

    private String getServiceFQDN(CoreV1Api api, String url) {
        String serviceName = url.replace("http://", "").split(":")[0];
        int port = Integer.parseInt(url.split(":")[2]);

        // 搜索所有namespace查找服务
        V1ServiceList serviceList = null;
        try {
            serviceList = api.listServiceForAllNamespaces(null, null,
                    "metadata.name=" + serviceName, null, null, null, null, null, null, null);
        } catch (ApiException e) {
            log.error("Failed to list services for serviceName: {}, error: {}", serviceName, e.getMessage(), e);
            throw new RuntimeException(e);
        }

        if (serviceList.getItems().isEmpty()) {
            log.error("Service not found in cluster: {}", serviceName);
            throw new RuntimeException("Service not found: " + serviceName);
        }

        V1Service service = serviceList.getItems().get(0);
        String namespace = Objects.requireNonNull(service.getMetadata()).getNamespace();
        log.debug("Successfully got FQDN for service: {}.{}:{}", serviceName, namespace, port);

        return String.format("http://%s.%s.svc.cluster.local:%d",
                serviceName,
                namespace,
                port);
    }
}
