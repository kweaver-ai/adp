SET SCHEMA vega;


INSERT INTO t_dict(f_dict_type,f_dict_key,f_dict_value,f_extend_property,f_enable_status) SELECT 1,18,'maxcompute','{"dbCatalogName": 关系型数据库}',1 FROM DUAL WHERE NOT EXISTS ( SELECT f_id from t_dict where f_dict_type = 1 AND f_dict_key = 18);
