package com.eisoo.engine.gateway.service.impl;

import com.alibaba.fastjson2.JSONObject;
import com.eisoo.engine.gateway.service.ClientIdService;
import com.eisoo.engine.metadata.entity.ClientIdEntity;
import com.eisoo.engine.metadata.mapper.ClientIdMapper;
import com.eisoo.engine.utils.common.Constants;
import com.eisoo.engine.utils.util.AFUtil;
import com.eisoo.engine.utils.util.StringUtils;
import com.github.yulichang.base.MPJBaseServiceImpl;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

@Service
public class ClientIdServiceImpl extends MPJBaseServiceImpl<ClientIdMapper, ClientIdEntity> implements ClientIdService {

    private static final Logger log = LoggerFactory.getLogger(ClientIdServiceImpl.class);

    private static ClientIdEntity clientIdEntity;

    @Autowired(required = false)
    ClientIdMapper clientIdMapper;

    @Value(value = "${auth.register-clientid-url}")
    private String registerClientIdUrl;

    @Override
    public ClientIdEntity reRegistClient() {
        clientIdMapper.deleteById(clientIdEntity);
        clientIdEntity = null;
        return getClientIdAndSecret();
    }

    @Override
    public ClientIdEntity getClientIdAndSecret() {
        if (clientIdEntity != null
                && StringUtils.isNotEmpty(clientIdEntity.getClientId())
                && StringUtils.isNotEmpty(clientIdEntity.getClientSecret())) {
            return clientIdEntity;
        }
        // 从数据库中查询client_id
        clientIdEntity = clientIdMapper.selectById(1);
        log.info("query client_id: {}", clientIdEntity);
        // 如果数据库中没有client_id则去注册一个，保存到数据库
        if (clientIdEntity == null) {
            clientIdEntity = new ClientIdEntity();
            clientIdEntity.setId(1);
            clientIdEntity.setClientName(Constants.OAUTH_CLIENT_NAME);
            clientIdEntity.setCreateTime(LocalDateTime.now());
            // 通过主键不允许重复实现分布式锁，防止多实例重复注册
            try {
                clientIdMapper.insert(clientIdEntity);
                log.info("insert db: {}", clientIdEntity);

                AFUtil afUtil = new AFUtil();
                JSONObject result = afUtil.registerClientId(registerClientIdUrl, Constants.OAUTH_CLIENT_NAME);
                log.info("regist client_id: {}", result.toJSONString());

                clientIdEntity.setClientId(result.getString("client_id"));
                clientIdEntity.setClientSecret(result.getString("client_secret"));
                clientIdEntity.setUpdateTime(LocalDateTime.now());
                if (StringUtils.isNotEmpty(clientIdEntity.getClientId())
                        && StringUtils.isNotEmpty(clientIdEntity.getClientSecret())) {
                    clientIdMapper.updateById(clientIdEntity);
                    log.info("update client_id: {}", clientIdEntity);
                } else {
                    clientIdMapper.deleteById(clientIdEntity);
                    clientIdEntity = null;
                }
            } catch (Exception e) {
                log.warn(e.getMessage());
                clientIdEntity = getClientFromDbAfterThreeSeconds();
            }
        } else if (StringUtils.isEmpty(clientIdEntity.getClientId()) ||
                StringUtils.isEmpty(clientIdEntity.getClientSecret())) {
            clientIdMapper.deleteById(clientIdEntity);
            clientIdEntity = null;
        } else {
            if (StringUtils.isEmpty(clientIdEntity.getClientId())) {
                clientIdEntity = getClientFromDbAfterThreeSeconds();
            }
        }
        return clientIdEntity;
    }

    private ClientIdEntity getClientFromDbAfterThreeSeconds() {
        try {
            // 如果查询到记录但是client_id为空，说明其他实例正在注册，休眠3秒后重新查询一次数据库
            Thread.sleep(3000);
            ClientIdEntity clientIdEntity = clientIdMapper.selectById(1);
            log.info("query client_id again: {}", clientIdEntity);
            return clientIdEntity;
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }
    }
}
