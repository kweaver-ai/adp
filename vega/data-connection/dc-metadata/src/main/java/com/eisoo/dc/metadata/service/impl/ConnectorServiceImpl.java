package com.eisoo.dc.metadata.service.impl;

import cn.hutool.json.JSONUtil;
import com.eisoo.dc.common.connector.ConnectorConfig;
import com.eisoo.dc.common.connector.ConnectorConfigCache;
import com.eisoo.dc.common.connector.TypeConfig;
import com.eisoo.dc.common.connector.mapping.TypeMapping;
import com.eisoo.dc.common.connector.mapping.TypeMappingFactory;
import com.eisoo.dc.common.constant.CatalogConstant;
import com.eisoo.dc.common.constant.Detail;
import com.eisoo.dc.common.exception.enums.ErrorCodeEnum;
import com.eisoo.dc.common.exception.vo.AiShuException;
import com.eisoo.dc.common.util.StringUtils;
import com.eisoo.dc.metadata.domain.dto.SourceTypeDto;
import com.eisoo.dc.metadata.domain.dto.TypeMappingDto;
import com.eisoo.dc.metadata.domain.vo.TargetMappingVo;
import com.eisoo.dc.metadata.domain.vo.TargetTypeVo;
import com.eisoo.dc.metadata.service.ConnectorService;
import com.google.common.collect.Lists;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

/**
 * @Author zdh
 **/
@Service
public class ConnectorServiceImpl implements ConnectorService {
    private static final Logger log = LoggerFactory.getLogger(ConnectorServiceImpl.class);

    @Autowired
    ConnectorConfigCache connectorConfigCache;

    @Autowired
    TypeMappingFactory mappingTypeFactory;

    @Override
    public ResponseEntity<?> getConnectorConfig(String connectorName) {
        log.info("读取数据源配置,connectorName:{}", connectorName);
        ConnectorConfig connectorConfig = connectorConfigCache.getConnectorConfig(connectorName);
        return ResponseEntity.ok(connectorConfig);
    }

    @Override
    public ResponseEntity<?> getConnectorsMapping(TypeMappingDto mappingDto) {
        log.info("type mapping request:{}", JSONUtil.parseObj(mappingDto));
        String sourceConnector = mappingDto.getSourceConnector();
        String targetConnector = mappingDto.getTargetConnector();

        ConnectorConfig sourceConfig = null;
        ConnectorConfig targetConfig = null;
        if (!sourceConnector.equals(CatalogConstant.CONNECTOR_VEGA)) {
            sourceConfig = connectorConfigCache.getConnectorConfig(sourceConnector);
        }
        if (!targetConnector.equals(CatalogConstant.CONNECTOR_VEGA)) {
            targetConfig = connectorConfigCache.getConnectorConfig(targetConnector);
        }

        TargetMappingVo targetMappingVo = mapping(sourceConfig, targetConfig, mappingDto.getType());
        return ResponseEntity.ok(targetMappingVo);
    }

    private TargetMappingVo mapping(ConnectorConfig sourceConfig, ConnectorConfig targetConfig, List<SourceTypeDto> types) {
        List<TargetTypeVo> targetTypeVos = Lists.newCopyOnWriteArrayList();
        types.stream().forEach(type -> {
            String sourceTypeName = type.getSourceType().toLowerCase();
            Long columnSize = type.getPrecision();
            Long decimalDigits = type.getDecimalDigits();
            Integer index = type.getIndex();
            if (StringUtils.isBlank(sourceTypeName)) {
                log.error("原始数据类型sourceTypeName为空或类型索引index为空,type:{}", type);
                throw new AiShuException(ErrorCodeEnum.InvalidParameter, Detail.SOURCE_TYPE_OR_INDEX_NULL);
            }

            TargetTypeVo targetTypeVo = new TargetTypeVo();
            targetTypeVo.setIndex(index);
            TypeConfig targetType = null;

            //vega类型->vega类型
            if (sourceConfig == null && targetConfig == null){
                if (!ConnectorConfigCache.vegaTypes.contains(sourceTypeName)){
                    targetTypeVo.setTargetTypeName("");
                }else{
                    targetType = new TypeConfig();
                    targetType.setVegaType(sourceTypeName);
                }
            }

            //真实类型->vega类型
            if (sourceConfig != null && targetConfig == null){
                TypeConfig sourceType = sourceConfig.getType().stream()
                        .filter(source -> StringUtils.equals(source.getSourceType(), sourceTypeName))
                        .findFirst().orElse(null);
                if (sourceType == null) {
                    targetTypeVo.setTargetTypeName("");
                }else{
                    targetType = new TypeConfig();
                    targetType.setVegaType(sourceType.getVegaType());
                }
            }

            //真实类型->真实类型
            if (sourceConfig != null && targetConfig != null){
                TypeConfig sourceType = sourceConfig.getType().stream()
                        .filter(source -> StringUtils.equals(source.getSourceType(), sourceTypeName))
                        .findFirst().orElse(null);
                if (sourceType == null) {
                    targetTypeVo.setTargetTypeName("");
                }else{
                    targetType = targetConfig.getType().stream()
                            .filter(target -> StringUtils.equals(target.getVegaType(), sourceType.getVegaType()))
                            .findFirst().orElse(null);
                }
            }

            //vega类型->真实类型
            if (sourceConfig == null && targetConfig != null){
                targetType = targetConfig.getType().stream()
                        .filter(target -> StringUtils.equals(target.getVegaType(), sourceTypeName))
                        .findFirst().orElse(null);
            }

            if (targetType == null){
                targetTypeVo.setTargetTypeName("");
            }else{
                Map<String, Long> mappingType = mappingType(targetType, columnSize, targetConfig);
                String target = mappingType.keySet().stream().findFirst().orElse("");
                if (StringUtils.isNotBlank(target)) {
                    columnSize = mappingType.get(target);
                } else {
                    columnSize = null;
                }
                if (StringUtils.equals("decimal", target)) {
                    if (columnSize != null) {
                        columnSize = Math.min(columnSize, 38L);
                    } else {
                        columnSize = 38L;
                    }
                    if (decimalDigits > 18) {
                        decimalDigits = 18L;
                    }
                }
                targetTypeVo.setTargetTypeName(target);
                targetTypeVo.setPrecision(columnSize);
                targetTypeVo.setDecimalDigits(decimalDigits);
            }
            targetTypeVos.add(targetTypeVo);
        });
        TargetMappingVo targetMappingVo = new TargetMappingVo();
        String targetConnector = StringUtils.isNotNull(targetConfig) ? targetConfig.getConnector() : CatalogConstant.CONNECTOR_VEGA;
        targetMappingVo.setTargetConnector(targetConnector);
        targetMappingVo.setType(targetTypeVos);
        return targetMappingVo;
    }

    private Map<String, Long> mappingType(TypeConfig targetType, Long columnSize, ConnectorConfig targetConfig) {
        String targetName = targetConfig==null ? CatalogConstant.CONNECTOR_VEGA : targetConfig.getConnector();
        TypeMapping connector = mappingTypeFactory.getConnector(targetName);
        return connector.getTypeMapping(targetType, columnSize, targetName);
    }

}


