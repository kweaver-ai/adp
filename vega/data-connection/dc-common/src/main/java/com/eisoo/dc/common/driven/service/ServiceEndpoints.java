package com.eisoo.dc.common.driven.service;

import java.util.function.Supplier;
import java.util.concurrent.ConcurrentHashMap;

public class ServiceEndpoints {
    private final ConcurrentHashMap<String, String> cache = new ConcurrentHashMap<>();

    private final Supplier<String> vegaCalculateCoordinator;
    private final Supplier<String> hydraAdmin;
    private final Supplier<String> userManagementPrivate;
    private final Supplier<String> authorizationPrivate;
    private final Supplier<String> efastPublic;
    private final Supplier<String> efastPrivate;

    public ServiceEndpoints(
            Supplier<String> vegaCalculateCoordinator,
            Supplier<String> hydraAdmin,
            Supplier<String> userManagementPrivate,
            Supplier<String> authorizationPrivate,
            Supplier<String> efastPublic,
            Supplier<String> efastPrivate) {
        this.vegaCalculateCoordinator = vegaCalculateCoordinator;
        this.hydraAdmin = memoize(hydraAdmin, "hydraAdmin");
        this.userManagementPrivate = memoize(userManagementPrivate, "userManagementPrivate");
        this.authorizationPrivate = memoize(authorizationPrivate, "authorizationPrivate");
        this.efastPublic = memoize(efastPublic, "efastPublic");
        this.efastPrivate = memoize(efastPrivate, "efastPrivate");
    }

    private Supplier<String> memoize(Supplier<String> supplier, String key) {
        return () -> cache.computeIfAbsent(key, k -> supplier.get());
    }

    // Getter方法
    public String getVegaCalculateCoordinator() { return vegaCalculateCoordinator.get(); }
    public String getHydraAdmin() { return hydraAdmin.get(); }
    public String getUserManagementPrivate() { return userManagementPrivate.get(); }
    public String getAuthorizationPrivate() { return authorizationPrivate.get(); }
    public String getEfastPublic() { return efastPublic.get(); }
    public String getEfastPrivate() { return efastPrivate.get(); }
}
