{
  "toolbox": {
    "configs": [
      {
        "box_id": "db4da399-af91-4214-afd9-1762d07c942d",
        "box_name": "contextloader工具集_030",
        "box_desc": "contextloader工具集_030",
        "box_svc_url": "http://agent-retrieval:30779",
        "status": "published",
        "category_type": "other_category",
        "category_name": "未分类",
        "is_internal": false,
        "source": "custom",
        "tools": [
          {
            "tool_id": "d2635960-2ad3-4608-85eb-a5011eee6022",
            "name": "get_action_info",
            "description": "根据对象实例标识召回关联行动，返回 _dynamic_tools。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "a312cfb0-0c1f-4485-857f-b7632e154e81",
              "summary": "get_action_info",
              "description": "根据对象实例标识召回关联行动，返回 _dynamic_tools。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/get_action_info",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770801820449627400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户 ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    },
                    "example": "bdb78b62-6c48-11f0-af96-fa8dcc0a06b2"
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "enum": [
                        "user",
                        "system"
                      ],
                      "type": "string"
                    },
                    "example": "user"
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "examples": {
                        "disease_example": {
                          "summary": "疾病对象示例",
                          "value": {
                            "_instance_identity": {
                              "disease_id": "disease_000001"
                            },
                            "at_id": "generate_treatment_plan",
                            "kn_id": "kn_medical"
                          }
                        }
                      },
                      "schema": {
                        "$ref": "#/components/schemas/ActionRecallRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回动态工具列表",
                    "content": {
                      "application/json": {
                        "examples": {
                          "成功": {
                            "value": {
                              "_dynamic_tools": [
                                {
                                  "api_url": "http://...",
                                  "description": "工具描述",
                                  "name": "工具名",
                                  "parameters": {}
                                }
                              ]
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ActionRecallResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "请求参数错误",
                    "content": {
                      "application/json": {
                        "examples": {
                          "invalid_request": {
                            "value": {
                              "code": "INVALID_REQUEST",
                              "description": "_instance_identity 格式错误"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "502",
                    "description": "上游服务不可用",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ActionRecallResponse": {
                      "properties": {
                        "headers": {
                          "type": "object"
                        },
                        "_dynamic_tools": {
                          "description": "Function Call 格式的工具列表",
                          "items": {
                            "properties": {
                              "fixed_params": {
                                "type": "object"
                              },
                              "name": {
                                "type": "string"
                              },
                              "parameters": {
                                "type": "object"
                              },
                              "api_url": {
                                "type": "string"
                              },
                              "description": {
                                "type": "string"
                              }
                            },
                            "type": "object"
                          },
                          "type": "array"
                        }
                      },
                      "type": "object",
                      "required": [
                        "_dynamic_tools"
                      ]
                    },
                    "ErrorResponse": {
                      "properties": {
                        "description": {
                          "type": "string"
                        },
                        "code": {
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "ActionRecallRequest": {
                      "properties": {
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络ID"
                        },
                        "_instance_identity": {
                          "type": "object",
                          "description": "对象实例标识（主键键值对）。**必须从 query_object_instance 或 query_instance_subgraph 返回的 _instance_identity 字段提取，不可臆造。** 当前仅支持单个对象。"
                        },
                        "at_id": {
                          "type": "string",
                          "description": "行动类ID（从 Schema 获取）"
                        }
                      },
                      "type": "object",
                      "required": [
                        "kn_id",
                        "at_id",
                        "_instance_identity"
                      ]
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "action-recall"
                ],
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770801820452636200,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "a312cfb0-0c1f-4485-857f-b7632e154e81",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "512c6f3b-a450-4c70-9b32-6ebf1d63a06e",
            "name": "get_logic_properties_values",
            "description": "根据 query 生成 dynamic_params，批量查询指定对象的逻辑属性值。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "7cc79f6c-70b9-4b76-9796-d382310846a7",
              "summary": "get_logic_properties_values",
              "description": "根据 query 生成 dynamic_params，批量查询指定对象的逻辑属性值。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/logic-property-resolver",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770802144554596400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "用户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "examples": {
                        "示例": {
                          "value": {
                            "_instance_identities": [
                              {
                                "company_id": "company_000001"
                              }
                            ],
                            "kn_id": "kn_medical",
                            "ot_id": "company",
                            "properties": [
                              "approved_drug_count",
                              "business_health_score"
                            ],
                            "query": "最近一年这些药企的药品上市数量和健康度"
                          }
                        }
                      },
                      "schema": {
                        "$ref": "#/components/schemas/ResolveLogicPropertiesRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "ok",
                    "content": {
                      "application/json": {
                        "examples": {
                          "成功示例": {
                            "value": {
                              "datas": [
                                {
                                  "approved_drug_count": {
                                    "datas": [
                                      {
                                        "values": [
                                          12
                                        ]
                                      }
                                    ]
                                  },
                                  "business_health_score": {
                                    "score": 0.82
                                  },
                                  "company_id": "company_000001"
                                }
                              ]
                            }
                          },
                          "缺参示例": {
                            "value": {
                              "error_code": "MISSING_INPUT_PARAMS",
                              "message": "dynamic_params 缺少必需参数",
                              "missing": [
                                {
                                  "params": [
                                    {
                                      "hint": "在 additional_context 中补充明确时间范围（start/end 或 “最近半年”），或在 query 中明确时间范围",
                                      "name": "start",
                                      "type": "INTEGER"
                                    }
                                  ],
                                  "property": "approved_drug_count"
                                }
                              ],
                              "trace_id": "3f5d6c1c-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                            }
                          }
                        },
                        "schema": {
                          "$ref": "#/components/schemas/ResolveLogicPropertiesResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "bad request",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/Error"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "internal error",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/Error"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ResolveLogicPropertiesResponse": {
                      "oneOf": [
                        {
                          "$ref": "#/components/schemas/ObjectPropertiesValuesResponse"
                        },
                        {
                          "$ref": "#/components/schemas/MissingParamsError"
                        }
                      ],
                      "description": "成功返回 datas；缺参时返回 error_code、missing（含 hint）"
                    },
                    "ObjectPropertiesValuesResponse": {
                      "type": "object",
                      "required": [
                        "datas"
                      ],
                      "properties": {
                        "datas": {
                          "type": "array",
                          "description": "与 _instance_identities 顺序对齐，每项含主键和请求的 properties",
                          "items": {
                            "type": "object"
                          }
                        },
                        "debug": {
                          "$ref": "#/components/schemas/ResolveDebugInfo"
                        }
                      }
                    },
                    "ResolveDebugInfo": {
                      "type": "object",
                      "properties": {
                        "trace_id": {
                          "type": "string"
                        },
                        "warnings": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array"
                        },
                        "dynamic_params": {
                          "type": "object"
                        },
                        "now_ms": {
                          "type": "integer"
                        }
                      }
                    },
                    "MissingParamsError": {
                      "type": "object",
                      "properties": {
                        "error_code": {
                          "type": "string"
                        },
                        "message": {
                          "type": "string"
                        },
                        "missing": {
                          "type": "array",
                          "items": {
                            "properties": {
                              "property": {
                                "type": "string"
                              },
                              "params": {
                                "items": {
                                  "type": "object",
                                  "properties": {
                                    "name": {
                                      "type": "string"
                                    },
                                    "hint": {
                                      "type": "string"
                                    }
                                  }
                                },
                                "type": "array"
                              }
                            },
                            "type": "object"
                          }
                        }
                      }
                    },
                    "Error": {
                      "type": "object",
                      "properties": {
                        "error_code": {
                          "type": "string"
                        },
                        "message": {
                          "type": "string"
                        }
                      }
                    },
                    "ResolveLogicPropertiesRequest": {
                      "type": "object",
                      "required": [
                        "kn_id",
                        "ot_id",
                        "query",
                        "_instance_identities",
                        "properties"
                      ],
                      "properties": {
                        "options": {
                          "$ref": "#/components/schemas/ResolveOptions"
                        },
                        "ot_id": {
                          "description": "对象类ID。例 company、drug",
                          "type": "string"
                        },
                        "properties": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "逻辑属性名列表（metric/operator）。自动生成 dynamic_params 并查询。"
                        },
                        "query": {
                          "description": "用户查询，需含时间（如\"最近一年\"）、统计维度、业务上下文，用于生成 dynamic_params",
                          "type": "string"
                        },
                        "_instance_identities": {
                          "type": "array",
                          "description": "对象实例标识数组。**必须从上游提取，不可臆造。** 流程：先调 query_object_instance 或 query_instance_subgraph → 从每个对象的 _instance_identity 字段取值 → 按原顺序组成数组传入。",
                          "items": {
                            "type": "object"
                          }
                        },
                        "additional_context": {
                          "description": "可选。补充上下文，如 timezone、instant、step、对象属性等，帮助生成 dynamic_params。",
                          "type": "string"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络ID。例 kn_medical"
                        }
                      }
                    },
                    "ResolveOptions": {
                      "type": "object",
                      "description": "【可选配置】控制接口行为的高级选项\n",
                      "properties": {
                        "return_debug": {
                          "type": "boolean",
                          "description": "是否返回 debug（dynamic_params、warnings 等）。默认 false"
                        },
                        "max_concurrency": {
                          "type": "integer",
                          "description": "多 property 时的 Agent 并发数。默认 4"
                        },
                        "max_repair_rounds": {
                          "type": "integer",
                          "description": "Agent JSON 非法时的修复轮次。默认 1"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770802144555141600,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "7cc79f6c-70b9-4b76-9796-d382310846a7",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "bfb36322-7e83-4c4a-b427-aab2233185ca",
            "name": "kn_schema_search",
            "description": "基于用户查询意图，返回业务知识网络中相关的概念信息",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "b0ad1932-549f-44d7-9963-ac42dc7e27d8",
              "summary": "kn_schema_search",
              "description": "基于用户查询意图，返回业务知识网络中相关的概念信息",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/semantic-search",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770868142271157800,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-user",
                    "in": "header",
                    "description": "userId",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-visitor-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/SemanticSearchRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "200",
                    "description": "成功返回相关概念信息",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/SemanticSearchResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "ResourceInfo": {
                      "type": "object",
                      "description": "数据来源信息",
                      "properties": {
                        "type": {
                          "type": "string",
                          "description": "数据来源类型"
                        },
                        "id": {
                          "type": "string",
                          "description": "数据视图id"
                        },
                        "name": {
                          "type": "string",
                          "description": "视图名称"
                        }
                      }
                    },
                    "ActionTypeDetail": {
                      "type": "object",
                      "description": "行动类概念详情",
                      "properties": {
                        "object_type_id": {
                          "description": "行动类所绑定的对象类ID",
                          "type": "string"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签",
                          "items": {
                            "type": "string"
                          }
                        },
                        "_score": {
                          "type": "number",
                          "format": "float",
                          "description": "分数"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "id": {
                          "type": "string",
                          "description": "行动类ID"
                        },
                        "module_type": {
                          "description": "模块类型",
                          "type": "string"
                        },
                        "name": {
                          "type": "string",
                          "description": "行动类名称"
                        }
                      }
                    },
                    "ObjectTypeDetail": {
                      "description": "对象类概念详情",
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "对象id"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签",
                          "items": {
                            "type": "string"
                          }
                        },
                        "logic_properties": {
                          "type": "array",
                          "description": "逻辑属性",
                          "items": {
                            "type": "object"
                          }
                        },
                        "module_type": {
                          "type": "string",
                          "description": "模块类型"
                        },
                        "name": {
                          "description": "对象名称",
                          "type": "string"
                        },
                        "primary_keys": {
                          "type": "array",
                          "description": "主键字段",
                          "items": {
                            "type": "string"
                          }
                        },
                        "_score": {
                          "type": "number",
                          "format": "float",
                          "description": "分数"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "data_properties": {
                          "type": "array",
                          "description": "数据属性",
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          }
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/ResourceInfo"
                        }
                      },
                      "type": "object"
                    },
                    "DataProperty": {
                      "type": "object",
                      "description": "数据属性结构定义",
                      "properties": {
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名称"
                        },
                        "mapped_field": {
                          "description": "视图字段信息"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称"
                        },
                        "type": {
                          "type": "string",
                          "description": "属性数据类型"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "condition_operations": {
                          "type": "array",
                          "description": "该数据属性支持的查询条件操作符列表。\n",
                          "items": {
                            "type": "string",
                            "enum": [
                              "==",
                              "!=",
                              ">",
                              "<",
                              ">=",
                              "<=",
                              "in",
                              "not_in",
                              "like",
                              "not_like",
                              "range",
                              "out_range",
                              "exist",
                              "not_exist",
                              "regex",
                              "match",
                              "knn"
                            ]
                          }
                        }
                      }
                    },
                    "Condition": {
                      "type": "object",
                      "properties": {
                        "value": {
                          "type": "string",
                          "description": "值"
                        },
                        "field": {
                          "type": "string",
                          "description": "字段名称"
                        },
                        "operation": {
                          "type": "string",
                          "description": "操作类型"
                        }
                      }
                    },
                    "QueryUnderstanding": {
                      "type": "object",
                      "properties": {
                        "processed_query": {
                          "type": "string",
                          "description": "LLM处理后的标准化查询"
                        },
                        "query_strategy": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/QueryStrategy"
                          }
                        },
                        "intent": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/QueryIntent"
                          }
                        },
                        "origin_query": {
                          "description": "用户原始查询",
                          "type": "string"
                        }
                      }
                    },
                    "QueryStrategy": {
                      "type": "object",
                      "properties": {
                        "strategy_type": {
                          "type": "string",
                          "description": "策略类型",
                          "enum": [
                            "concept_get",
                            "concept_discovery",
                            "object_instance_discovery"
                          ]
                        },
                        "filter": {
                          "$ref": "#/components/schemas/ConceptFilter"
                        }
                      }
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "code": {
                          "description": "错误码",
                          "type": "string"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        },
                        "detail": {
                          "description": "错误详情",
                          "type": "object"
                        },
                        "link": {
                          "type": "string",
                          "description": "错误链接"
                        },
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        }
                      }
                    },
                    "SemanticSearchRequest": {
                      "type": "object",
                      "required": [
                        "query",
                        "kn_id"
                      ],
                      "properties": {
                        "max_concepts": {
                          "description": "最大返回概念数量",
                          "default": 10,
                          "type": "integer"
                        },
                        "query": {
                          "type": "string",
                          "description": "用户自然语言查询"
                        },
                        "rerank_action": {
                          "type": "string",
                          "description": "重排动作",
                          "default": "default",
                          "enum": [
                            "default",
                            "vector",
                            "llm"
                          ]
                        },
                        "search_scope": {
                          "$ref": "#/components/schemas/SearchScope"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        }
                      }
                    },
                    "SemanticSearchResponse": {
                      "type": "object",
                      "properties": {
                        "concepts": {
                          "items": {
                            "$ref": "#/components/schemas/Concept"
                          },
                          "type": "array"
                        },
                        "query_understanding": {
                          "$ref": "#/components/schemas/QueryUnderstanding"
                        }
                      }
                    },
                    "SearchScope": {
                      "type": "object",
                      "description": "【可选】搜索域配置\n",
                      "properties": {
                        "concept_groups": {
                          "type": "array",
                          "description": "限定的概念分组",
                          "items": {
                            "type": "string"
                          }
                        },
                        "include_action_types": {
                          "type": "boolean",
                          "description": "是否包含行作类"
                        },
                        "include_object_types": {
                          "description": "是否包含对象类",
                          "type": "boolean"
                        },
                        "include_relation_types": {
                          "description": "是否包含关系类",
                          "type": "boolean"
                        }
                      }
                    },
                    "RelationTypeDetail": {
                      "properties": {
                        "type": {
                          "type": "string",
                          "description": "关系类型"
                        },
                        "source_object_type_id": {
                          "type": "string",
                          "description": "起点对象类ID"
                        },
                        "target_object_type_id": {
                          "type": "string",
                          "description": "目标对象类ID"
                        },
                        "_score": {
                          "type": "number",
                          "format": "float",
                          "description": "分数"
                        },
                        "module_type": {
                          "type": "string",
                          "description": "模块类型"
                        },
                        "comment": {
                          "type": "string",
                          "description": "备注"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签",
                          "items": {
                            "type": "string"
                          }
                        },
                        "name": {
                          "type": "string",
                          "description": "关系类名称"
                        },
                        "id": {
                          "type": "string",
                          "description": "关系类id"
                        }
                      },
                      "type": "object",
                      "description": "关系类概念详情"
                    },
                    "QueryIntent": {
                      "type": "object",
                      "properties": {
                        "query_segment": {
                          "type": "string",
                          "description": "对应的查询片段"
                        },
                        "reasoning": {
                          "type": "string",
                          "description": "简要识别推理过程"
                        },
                        "related_concepts": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/RelatedConcept"
                          }
                        },
                        "requires_reasoning": {
                          "type": "boolean",
                          "description": "是否需要进一步推理",
                          "default": false
                        },
                        "confidence": {
                          "type": "number",
                          "format": "float",
                          "description": "置信度"
                        }
                      }
                    },
                    "RelatedConcept": {
                      "type": "object",
                      "properties": {
                        "concept_id": {
                          "type": "string",
                          "description": "概念类ID"
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念类名称"
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型",
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ]
                        }
                      }
                    },
                    "Concept": {
                      "type": "object",
                      "properties": {
                        "rerank_score": {
                          "description": "重排序得分",
                          "type": "number",
                          "format": "float"
                        },
                        "samples": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "实例样本列表"
                        },
                        "concept_detail": {
                          "description": "概念类详情，根据concept_type返回不同结构：\n- 当concept_type为\"object_type\"时，返回ObjectTypeDetail结构，包含对象类的完整信息\n- 当concept_type为\"relation_type\"时，返回RelationTypeDetail结构，包含关系类的完整信息\n- 当concept_type为\"action_type\"时，返回ActionTypeDetail结构，包含行动类的完整信息\n",
                          "oneOf": [
                            {
                              "$ref": "#/components/schemas/ObjectTypeDetail"
                            },
                            {
                              "$ref": "#/components/schemas/RelationTypeDetail"
                            },
                            {
                              "$ref": "#/components/schemas/ActionTypeDetail"
                            }
                          ]
                        },
                        "concept_id": {
                          "type": "string",
                          "description": "概念类ID"
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念类名称"
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型",
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ]
                        },
                        "intent_score": {
                          "type": "number",
                          "format": "float",
                          "description": "意图得分"
                        },
                        "match_score": {
                          "format": "float",
                          "description": "匹配得分",
                          "type": "number"
                        }
                      }
                    },
                    "ConceptFilter": {
                      "properties": {
                        "concept_type": {
                          "enum": [
                            "object_type",
                            "relation_type",
                            "action_type"
                          ],
                          "type": "string",
                          "description": "概念类型"
                        },
                        "conditions": {
                          "type": "array",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          }
                        }
                      },
                      "type": "object"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "SemanticSearch"
                ],
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770868142271694300,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "b0ad1932-549f-44d7-9963-ac42dc7e27d8",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "1b02f55a-03fb-4501-bcf6-f0c1c3858c15",
            "name": "kn_search",
            "description": "基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息。\n支持概念召回、语义实例召回、多轮对话等功能。\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "75edbf4b-d7c2-475c-b714-1bf9abd1dc65",
              "summary": "kn_search",
              "description": "基于知识网络的智能检索工具，支持传入完整的问题或一个或多个关键词，能够检索问题或关键词的属性信息和上下文信息。\n支持概念召回、语义实例召回、多轮对话等功能。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/kn_search",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770801820449627400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "kn_search 请求体",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/KnSearchRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "成功返回检索结果",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/KnSearchResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "LogicProperty": {
                      "type": "object",
                      "properties": {
                        "name": {
                          "description": "属性名称",
                          "type": "string"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名称"
                        }
                      }
                    },
                    "ConceptRetrievalConfig": {
                      "type": "object",
                      "description": "概念召回/概念流程配置参数（原最外层参数已收敛到此处）",
                      "properties": {
                        "include_sample_data": {
                          "type": "boolean",
                          "description": "是否获取对象类型的样例数据。True会为每个召回对象类型获取一条样例数据。",
                          "default": false
                        },
                        "schema_brief": {
                          "default": true,
                          "type": "boolean",
                          "description": "概念召回时是否返回精简schema。True仅返回必要字段（概念ID/名称/关系source&target），不返回大字段。"
                        },
                        "top_k": {
                          "description": "概念召回返回最相关关系类型数量（对象类型会随关系类型自动过滤）。",
                          "default": 10,
                          "type": "integer"
                        }
                      }
                    },
                    "RelationType": {
                      "required": [
                        "concept_id",
                        "concept_name",
                        "source_object_type_id",
                        "target_object_type_id"
                      ],
                      "properties": {
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型: relation_type"
                        },
                        "source_object_type_id": {
                          "type": "string",
                          "description": "源对象类型ID"
                        },
                        "target_object_type_id": {
                          "type": "string",
                          "description": "目标对象类型ID"
                        },
                        "concept_id": {
                          "type": "string",
                          "description": "概念ID"
                        },
                        "concept_name": {
                          "description": "概念名称",
                          "type": "string"
                        }
                      },
                      "type": "object"
                    },
                    "ErrorResponse": {
                      "properties": {
                        "message": {
                          "description": "错误详情",
                          "type": "string"
                        },
                        "error": {
                          "type": "string",
                          "description": "错误信息"
                        }
                      },
                      "type": "object"
                    },
                    "KnSearchResponse": {
                      "description": "检索结果，返回object_types/relation_types/action_types，并返回语义实例nodes/message。\n多轮时由concept_retrieval.return_union控制 nodes 的并集/增量。\n",
                      "properties": {
                        "object_types": {
                          "type": "array",
                          "description": "对象类型列表（概念召回时返回）。\n当schema_brief=True时，仅包含：concept_id, concept_name, comment, data_properties（仅name和display_name）, logic_properties（仅name和display_name）, sample_data（当include_sample_data=True时）。\n当schema_brief=False时，包含完整字段（包括primary_keys, display_key, sample_data等）\n",
                          "items": {
                            "$ref": "#/components/schemas/ObjectType"
                          }
                        },
                        "relation_types": {
                          "items": {
                            "$ref": "#/components/schemas/RelationType"
                          },
                          "type": "array",
                          "description": "关系类型列表（概念召回时返回）。\n精简模式和完整模式均包含：concept_id, concept_name, source_object_type_id, target_object_type_id\n"
                        },
                        "action_types": {
                          "type": "array",
                          "description": "操作类型列表（概念召回时返回）。\n当schema_brief=True时，每个action_type仅包含以下字段：id, name, action_type, object_type_id, object_type_name, comment, tags, kn_id\n",
                          "items": {
                            "$ref": "#/components/schemas/ActionType"
                          }
                        },
                        "message": {
                          "type": "string",
                          "description": "提示信息（例如未召回到实例数据时返回原因说明）"
                        },
                        "nodes": {
                          "description": "语义实例召回结果（当不提供conditions且召回到实例时返回），与条件召回节点风格对齐的扁平列表。\n每个节点至少包含 object_type_id、<object_type_id>_name、unique_identities\n",
                          "items": {
                            "$ref": "#/components/schemas/Node"
                          },
                          "type": "array"
                        }
                      },
                      "type": "object"
                    },
                    "Node": {
                      "type": "object",
                      "description": "节点数据，至少包含 object_type_id、<object_type_id>_name、unique_identities",
                      "properties": {
                        "unique_identities": {
                          "type": "object",
                          "description": "对象的唯一标识信息"
                        },
                        "object_type_id": {
                          "type": "string"
                        }
                      }
                    },
                    "ActionType": {
                      "description": "操作类型信息。精简模式（schema_brief=True）下仅包含：id, name, action_type, object_type_id, object_type_name, comment, tags, kn_id",
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "操作类型名称"
                        },
                        "object_type_id": {
                          "type": "string",
                          "description": "对象类型ID"
                        },
                        "object_type_name": {
                          "description": "对象类型名称",
                          "type": "string"
                        },
                        "tags": {
                          "type": "array",
                          "description": "标签列表",
                          "items": {
                            "type": "string"
                          }
                        },
                        "action_type": {
                          "type": "string",
                          "description": "操作类型（如：add, modify等）"
                        },
                        "comment": {
                          "description": "注释说明",
                          "type": "string"
                        },
                        "id": {
                          "description": "操作类型ID",
                          "type": "string"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "知识网络ID"
                        }
                      },
                      "type": "object"
                    },
                    "ObjectType": {
                      "properties": {
                        "comment": {
                          "description": "概念描述",
                          "type": "string"
                        },
                        "concept_id": {
                          "type": "string",
                          "description": "概念ID"
                        },
                        "concept_type": {
                          "type": "string",
                          "description": "概念类型: object_type"
                        },
                        "data_properties": {
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          },
                          "type": "array",
                          "description": "对象属性列表。精简模式下仅包含name和display_name字段（数量不截断）"
                        },
                        "primary_keys": {
                          "type": "array",
                          "description": "主键字段列表（支持多个主键）。仅当schema_brief=False时返回",
                          "items": {
                            "type": "string"
                          }
                        },
                        "concept_name": {
                          "type": "string",
                          "description": "概念名称"
                        },
                        "logic_properties": {
                          "items": {
                            "$ref": "#/components/schemas/LogicProperty"
                          },
                          "type": "array",
                          "description": "逻辑属性列表（指标等）。精简模式下仅包含name和display_name字段（数量不截断）"
                        },
                        "sample_data": {
                          "type": "object",
                          "description": "样例数据（当include_sample_data=True时返回，无论schema_brief是否为True）"
                        },
                        "display_key": {
                          "description": "显示字段名（用于获取instance_name）。仅当schema_brief=False时返回",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "required": [
                        "concept_id",
                        "concept_name"
                      ]
                    },
                    "DataProperty": {
                      "type": "object",
                      "properties": {
                        "comment": {
                          "type": "string",
                          "description": "属性描述（非精简模式）"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名称"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称"
                        }
                      }
                    },
                    "KnSearchRequest": {
                      "type": "object",
                      "required": [
                        "query",
                        "kn_id"
                      ],
                      "properties": {
                        "enable_rerank": {
                          "default": true,
                          "type": "boolean",
                          "description": "是否启用重排序。如果为true，则启用重排序。"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "指定的知识网络ID，必须传递"
                        },
                        "only_schema": {
                          "type": "boolean",
                          "description": "是否只召回概念（schema），不召回语义实例。如果为True，则只返回object_types、relation_types和action_types，不返回nodes。",
                          "default": false
                        },
                        "query": {
                          "type": "string",
                          "description": "用户查询问题或关键词，多个关键词之间用空格隔开"
                        },
                        "retrieval_config": {
                          "properties": {
                            "concept_retrieval": {
                              "$ref": "#/components/schemas/ConceptRetrievalConfig"
                            }
                          },
                          "type": "object",
                          "description": "召回配置参数，用于控制不同类型的召回场景（概念召回、语义实例召回、属性过滤）。如果不提供，将使用系统默认配置。"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "kn-search"
                ],
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770801820452636200,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "75edbf4b-d7c2-475c-b714-1bf9abd1dc65",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "03b24cbb-e43a-40ab-950a-06f9343aec26",
            "name": "query_instance_subgraph",
            "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\n",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "3409150e-71c0-45c8-8cb7-eb6ed5ceebf9",
              "summary": "query_instance_subgraph",
              "description": "基于预定义的关系路径查询知识图谱中的对象子图。支持多条路径查询，每条路径返回独立子图。对象以map形式返回，支持过滤条件和排序。query_type需设为\"relation_path\"。\n",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_instance_subgraph",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770801820449627400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  }
                ],
                "request_body": {
                  "description": "子图查询请求体",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/SubGraphQueryBaseOnTypePath"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "对象子图查询响应体",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/PathEntries"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "Condition": {
                      "properties": {
                        "field": {
                          "description": "字段名称，也即对象类的属性名称",
                          "type": "string"
                        },
                        "operation": {
                          "enum": [
                            "and",
                            "or",
                            "==",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<=",
                            "in",
                            "not_in",
                            "like",
                            "not_like",
                            "exist",
                            "not_exist",
                            "match"
                          ],
                          "type": "string",
                          "description": "查询条件操作符。**注意：** 虽然这里列出了所有可能的操作符，但每个对象类实际支持的操作符列表以对象类定义中的 `condition_operations` 字段为准。"
                        },
                        "sub_conditions": {
                          "type": "array",
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          }
                        },
                        "value": {
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n",
                          "oneOf": [
                            {
                              "type": "string"
                            },
                            {
                              "type": "number"
                            },
                            {
                              "type": "boolean"
                            },
                            {
                              "type": "array",
                              "items": {
                                "oneOf": [
                                  {
                                    "type": "string"
                                  },
                                  {
                                    "type": "number"
                                  },
                                  {
                                    "type": "boolean"
                                  }
                                ]
                              }
                            }
                          ]
                        },
                        "value_from": {
                          "type": "string",
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ]
                        }
                      },
                      "type": "object",
                      "description": "过滤条件结构，用于构建对象实例的查询筛选条件。\n\n**重要规则：**\n- `value_from` 和 `value` 必须同时使用，不能单独使用\n- `value_from` 当前仅支持 \"const\"（常量值）\n- 当使用 `value_from: \"const\"` 时，必须同时提供 `value` 字段\n",
                      "required": [
                        "operation"
                      ]
                    },
                    "ObjectSubGraphResponse": {
                      "type": "object",
                      "description": "对象子图",
                      "required": [
                        "objects",
                        "relation_paths",
                        "total_count",
                        "search_after"
                      ],
                      "properties": {
                        "objects": {
                          "description": "子图中的对象map，格式为：\n{\n  \"对象ID1\": {ObjectInfoInSubgraph对象1},\n  \"对象ID2\": {ObjectInfoInSubgraph对象2}\n}\n其中key是ObjectInfoInSubgraph中的id属性，value是完整的ObjectInfoInSubgraph对象。\n动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty\n",
                          "type": "object"
                        },
                        "relation_paths": {
                          "items": {
                            "$ref": "#/components/schemas/RelationPath"
                          },
                          "type": "array",
                          "description": "对象的关系路径集合"
                        },
                        "search_after": {
                          "type": "array",
                          "description": "表示返回的最后一个起点类对象的排序值，获取这个用于下一次 search_after 分页",
                          "items": {}
                        },
                        "total_count": {
                          "type": "integer",
                          "description": "起点对象类的总条数"
                        }
                      }
                    },
                    "RelationPath": {
                      "type": "object",
                      "description": "对象的关系路径",
                      "required": [
                        "relations",
                        "length"
                      ],
                      "properties": {
                        "length": {
                          "type": "integer",
                          "description": "当前路径的长度"
                        },
                        "relations": {
                          "type": "array",
                          "description": "路径的边集合，沿着路径顺序出现的边",
                          "items": {
                            "$ref": "#/components/schemas/Relation"
                          }
                        }
                      }
                    },
                    "SubGraphQueryBaseOnTypePath": {
                      "properties": {
                        "relation_type_paths": {
                          "type": "array",
                          "description": "关系类路径集合,数组中可以包含多条不同的关系路径，系统会同时查询并返回所有路径的结果。每条路径必须符合严格的顺序和方向要求。",
                          "items": {
                            "$ref": "#/components/schemas/RelationTypePath"
                          }
                        }
                      },
                      "type": "object",
                      "description": "查询请求的顶层结构。用于基于关系类路径查询对象子图。relation_type_paths数组中可以包含多条不同的关系路径，系统会同时查询并返回所有路径的结果。每条路径必须符合严格的顺序和方向要求。",
                      "required": [
                        "relation_type_paths"
                      ]
                    },
                    "Relation": {
                      "properties": {
                        "target_object_id": {
                          "type": "string",
                          "description": "终点对象id"
                        },
                        "relation_type_id": {
                          "description": "关系类id",
                          "type": "string"
                        },
                        "relation_type_name": {
                          "type": "string",
                          "description": "关系类名称"
                        },
                        "source_object_id": {
                          "description": "起点对象id",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "一度关系（边）",
                      "required": [
                        "relation_type_id",
                        "relation_type_name",
                        "source_object_id",
                        "target_object_id"
                      ]
                    },
                    "PathEntries": {
                      "type": "object",
                      "description": "路径子图返回体",
                      "required": [
                        "entries"
                      ],
                      "properties": {
                        "entries": {
                          "type": "array",
                          "description": "路径子图",
                          "items": {
                            "$ref": "#/components/schemas/ObjectSubGraphResponse"
                          }
                        }
                      }
                    },
                    "TypeEdge": {
                      "properties": {
                        "target_object_type_id": {
                          "type": "string",
                          "description": "路径的终点对象类id"
                        },
                        "relation_type_id": {
                          "type": "string",
                          "description": "关系类id"
                        },
                        "source_object_type_id": {
                          "description": "路径的起点对象类id",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "路径中的边信息。**方向和顺序极其重要**！通过关系类id确定边，通过路径的起点对象类id和终点对象类id来确定当前路径的方向为正向还是反向，与关系类的起终点一致为正向，相反则为反向。每个TypeEdge必须与路径中的前后对象类型严格对应，这直接影响查询结果的正确性。",
                      "required": [
                        "relation_type_id",
                        "source_object_type_id",
                        "target_object_type_id"
                      ]
                    },
                    "ObjectTypeOnPath": {
                      "required": [
                        "id",
                        "condition",
                        "limit"
                      ],
                      "properties": {
                        "limit": {
                          "type": "integer",
                          "description": "对象类获取对象数量的限制"
                        },
                        "sort": {
                          "description": "对当前对象类的排序字段",
                          "items": {
                            "$ref": "#/components/schemas/Sort"
                          },
                          "type": "array"
                        },
                        "condition": {
                          "$ref": "#/components/schemas/Condition"
                        },
                        "id": {
                          "type": "string",
                          "description": "对象类id"
                        }
                      },
                      "type": "object",
                      "description": "路径中的对象类信息"
                    },
                    "RelationTypePath": {
                      "type": "object",
                      "description": "基于路径获取对象子图。**这是查询的核心结构**！用于定义完整的关系路径查询模板，包括路径中的所有对象类型和关系类型。object_types和relation_types数组的顺序**必须严格对应**，共同构成一个完整的关系路径。",
                      "required": [
                        "relation_types",
                        "object_types"
                      ],
                      "properties": {
                        "limit": {
                          "type": "integer",
                          "description": "当前路径返回的路径数量的限制。"
                        },
                        "object_types": {
                          "description": "路径中的对象类集合，**顺序必须严格**与路径中节点出现顺序保持一致。对于n跳路径，object_types数组长度应为n+1，且必须按照source_object_type → 中间节点 → target_object_type的顺序排列。如果某个节点没有过滤条件或者排序或者限制数量，也必须保留其id字段以确保顺序正确。",
                          "items": {
                            "$ref": "#/components/schemas/ObjectTypeOnPath"
                          },
                          "type": "array"
                        },
                        "relation_types": {
                          "type": "array",
                          "description": "路径的边集合，**顺序必须严格**按照路径中关系出现的顺序排列。对于n跳路径，relation_types数组长度应为n，且必须与object_types数组中的对象类型严格对应：第i个relation_type的source_object_type_id必须等于object_types数组中第i个对象的id，target_object_type_id必须等于object_types数组中第i+1个对象的id。",
                          "items": {
                            "$ref": "#/components/schemas/TypeEdge"
                          }
                        }
                      }
                    },
                    "Sort": {
                      "required": [
                        "field",
                        "direction"
                      ],
                      "properties": {
                        "direction": {
                          "enum": [
                            "desc",
                            "asc"
                          ],
                          "type": "string",
                          "description": "排序方向"
                        },
                        "field": {
                          "description": "排序字段",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "排序字段"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770801820452636200,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "3409150e-71c0-45c8-8cb7-eb6ed5ceebf9",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "0f2b2b86-8af3-4fcc-9d8f-810a4f3fa6ce",
            "name": "query_object_instance",
            "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "e378f603-3325-4752-b852-0c409b7fcbaf",
              "summary": "query_object_instance",
              "description": "根据单个对象类查询对象实例，该接口基于业务知识网络语义检索接口返回的对象类定义，查询具体的对象实例数据。",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/query_object_instance",
              "method": "POST",
              "create_time": 1770801820449627400,
              "update_time": 1770801820449627400,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "ot_id",
                    "in": "query",
                    "description": "对象类ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "include_type_info",
                    "in": "query",
                    "description": "是否包含对象类信息, 默认false，不包含",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  },
                  {
                    "name": "include_logic_params",
                    "in": "query",
                    "description": "包含逻辑属性的计算参数，默认false，返回结果不包含逻辑属性的字段和值",
                    "required": false,
                    "schema": {
                      "type": "boolean"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "schema": {
                        "$ref": "#/components/schemas/FirstQueryWithSearchAfter"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "200",
                    "description": "ok",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ObjectDataResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "Sort": {
                      "description": "排序字段",
                      "required": [
                        "field",
                        "direction"
                      ],
                      "properties": {
                        "direction": {
                          "enum": [
                            "desc",
                            "asc"
                          ],
                          "type": "string",
                          "description": "排序方向"
                        },
                        "field": {
                          "type": "string",
                          "description": "排序字段"
                        }
                      },
                      "type": "object"
                    },
                    "VectorConfig": {
                      "type": "object",
                      "description": "向量索引的配置",
                      "required": [
                        "dimension"
                      ],
                      "properties": {
                        "dimension": {
                          "type": "integer",
                          "description": "向量维度"
                        }
                      }
                    },
                    "DataSource": {
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "数据视图ID"
                        },
                        "name": {
                          "description": "名称。查看详情时返回。",
                          "type": "string"
                        },
                        "type": {
                          "description": "数据来源类型为数据视图",
                          "enum": [
                            "data_view"
                          ],
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "数据来源",
                      "required": [
                        "type",
                        "id"
                      ]
                    },
                    "LogicSource": {
                      "description": "数据来源",
                      "required": [
                        "type",
                        "id"
                      ],
                      "properties": {
                        "type": {
                          "description": "数据来源类型",
                          "enum": [
                            "metric",
                            "operator"
                          ],
                          "type": "string"
                        },
                        "id": {
                          "type": "string",
                          "description": "数据来源ID"
                        },
                        "name": {
                          "type": "string",
                          "description": "名称。查看详情时返回。"
                        }
                      },
                      "type": "object"
                    },
                    "ConceptGroup": {
                      "required": [
                        "id",
                        "name"
                      ],
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "概念分组ID"
                        },
                        "name": {
                          "description": "概念分组名称",
                          "type": "string"
                        }
                      },
                      "type": "object",
                      "description": "概念分组"
                    },
                    "FirstQueryWithSearchAfter": {
                      "type": "object",
                      "description": "分页查询的第一次查询请求",
                      "required": [
                        "limit"
                      ],
                      "properties": {
                        "properties": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "指定返回的对象属性字段列表，默认返回所有属性。"
                        },
                        "sort": {
                          "description": "排序字段，默认使用 @timestamp排序，排序方向为 desc",
                          "items": {
                            "$ref": "#/components/schemas/Sort"
                          },
                          "type": "array"
                        },
                        "condition": {
                          "$ref": "#/components/schemas/Condition"
                        },
                        "limit": {
                          "type": "integer",
                          "description": "返回的数量，默认值 10。范围 1-100"
                        },
                        "need_total": {
                          "type": "boolean",
                          "description": "是否需要总数，默认false"
                        }
                      }
                    },
                    "DataProperty": {
                      "required": [
                        "name",
                        "display_name",
                        "type",
                        "comment",
                        "mapped_field",
                        "index",
                        "fulltext_config",
                        "vector_config"
                      ],
                      "properties": {
                        "comment": {
                          "description": "属性描述",
                          "type": "string"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名"
                        },
                        "fulltext_config": {
                          "$ref": "#/components/schemas/FulltextConfig"
                        },
                        "index": {
                          "type": "boolean",
                          "description": "是否开启索引，默认是true"
                        },
                        "mapped_field": {
                          "$ref": "#/components/schemas/ViewField"
                        },
                        "name": {
                          "description": "属性名称。只能包含小写英文字母、数字、下划线（_）、连字符（-），且不能以下划线和连字符开头",
                          "type": "string"
                        },
                        "type": {
                          "description": "属性数据类型。除了视图的字段类型之外，还有 metric、objective、event、trace、log、operator",
                          "type": "string"
                        },
                        "vector_config": {
                          "$ref": "#/components/schemas/VectorConfig"
                        }
                      },
                      "type": "object",
                      "description": "数据属性"
                    },
                    "ObjectDataResponse": {
                      "type": "object",
                      "description": "节点（对象类）信息",
                      "required": [
                        "groups",
                        "type",
                        "datas",
                        "search_after"
                      ],
                      "properties": {
                        "datas": {
                          "items": {
                            "type": "object"
                          },
                          "type": "array",
                          "description": "对象实例数据。动态数据字段，其值可以是基本类型、MetricProperty或OperatorProperty"
                        },
                        "object_type": {
                          "$ref": "#/components/schemas/ObjectTypeDetail"
                        },
                        "search_after": {
                          "description": "表示返回的最后一个文档的排序值，获取这个用于下一次 search_after 分页",
                          "items": {},
                          "type": "array"
                        },
                        "total_count": {
                          "type": "integer",
                          "description": "总条数"
                        }
                      }
                    },
                    "ViewField": {
                      "type": "object",
                      "description": "视图字段信息",
                      "required": [
                        "name"
                      ],
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "字段名称"
                        },
                        "type": {
                          "type": "string",
                          "description": "视图字段类型，查看时有此字段"
                        },
                        "display_name": {
                          "type": "string",
                          "description": "字段显示名.查看时有此字段"
                        }
                      }
                    },
                    "FulltextConfig": {
                      "description": "全文索引的配置",
                      "required": [
                        "analyzer",
                        "field_keyword"
                      ],
                      "properties": {
                        "analyzer": {
                          "description": "分词器",
                          "enum": [
                            "standard",
                            "ik_max_word"
                          ],
                          "type": "string"
                        },
                        "field_keyword": {
                          "type": "boolean",
                          "description": "是否保留原始字符串，保留原始字符串可用于精确匹配。默认是false"
                        }
                      },
                      "type": "object"
                    },
                    "LogicProperty": {
                      "type": "object",
                      "description": "逻辑属性",
                      "required": [
                        "name",
                        "data_source",
                        "parameters"
                      ],
                      "properties": {
                        "display_name": {
                          "type": "string",
                          "description": "属性显示名"
                        },
                        "index": {
                          "type": "boolean",
                          "description": "是否开启索引，默认是true"
                        },
                        "name": {
                          "type": "string",
                          "description": "属性名称。只能包含小写英文字母、数字、下划线（_）、连字符（-），且不能以下划线和连字符开头"
                        },
                        "parameters": {
                          "description": "逻辑所需的参数",
                          "items": {
                            "$ref": "#/components/schemas/Parameter"
                          },
                          "type": "array"
                        },
                        "type": {
                          "description": "属性数据类型。除了视图的字段类型之外，还有 metric、objective、event、trace、log、operator",
                          "type": "string"
                        },
                        "comment": {
                          "description": "属性描述",
                          "type": "string"
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/LogicSource"
                        }
                      }
                    },
                    "ObjectTypeDetail": {
                      "description": "对象类信息",
                      "properties": {
                        "tags": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "标签。 （可以为空）"
                        },
                        "comment": {
                          "description": "备注（可以为空）",
                          "type": "string"
                        },
                        "create_time": {
                          "format": "int64",
                          "description": "创建时间",
                          "type": "integer"
                        },
                        "primary_keys": {
                          "items": {
                            "type": "string"
                          },
                          "type": "array",
                          "description": "主键"
                        },
                        "creator": {
                          "description": "创建人ID",
                          "type": "string"
                        },
                        "icon": {
                          "description": "图标",
                          "type": "string"
                        },
                        "detail": {
                          "type": "string",
                          "description": "说明书。按需返回，若指定了include_detail=true，则返回，否则不返回。列表查询时不返回此字段"
                        },
                        "module_type": {
                          "enum": [
                            "object_type"
                          ],
                          "type": "string",
                          "description": "模块类型"
                        },
                        "logic_properties": {
                          "type": "array",
                          "description": "逻辑属性",
                          "items": {
                            "$ref": "#/components/schemas/LogicProperty"
                          }
                        },
                        "data_source": {
                          "$ref": "#/components/schemas/DataSource"
                        },
                        "updater": {
                          "description": "最近一次修改人",
                          "type": "string"
                        },
                        "update_time": {
                          "format": "int64",
                          "description": "最近一次更新时间",
                          "type": "integer"
                        },
                        "concept_groups": {
                          "type": "array",
                          "description": "概念分组id",
                          "items": {
                            "$ref": "#/components/schemas/ConceptGroup"
                          }
                        },
                        "name": {
                          "type": "string",
                          "description": "对象类名称"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络id"
                        },
                        "data_properties": {
                          "type": "array",
                          "description": "数据属性",
                          "items": {
                            "$ref": "#/components/schemas/DataProperty"
                          }
                        },
                        "color": {
                          "type": "string",
                          "description": "颜色"
                        },
                        "branch": {
                          "description": "分支ID",
                          "type": "string"
                        },
                        "id": {
                          "description": "对象类ID",
                          "type": "string"
                        },
                        "display_key": {
                          "type": "string",
                          "description": "对象实例的显示属性"
                        }
                      },
                      "type": "object"
                    },
                    "Parameter4Operator": {
                      "type": "object",
                      "description": "逻辑参数",
                      "required": [
                        "name",
                        "value_from"
                      ],
                      "properties": {
                        "value": {
                          "type": "string",
                          "description": "参数值。value_from=property时，填入的是对象类的数据属性名称；value_from=input时，不设置此字段"
                        },
                        "value_from": {
                          "type": "string",
                          "description": "值来源",
                          "enum": [
                            "property",
                            "input"
                          ]
                        },
                        "name": {
                          "type": "string",
                          "description": "参数名称"
                        },
                        "source": {
                          "type": "string",
                          "description": "参数来源"
                        },
                        "type": {
                          "type": "string",
                          "description": "参数类型"
                        }
                      }
                    },
                    "Condition": {
                      "type": "object",
                      "description": "过滤条件结构，用于构建对象实例的查询筛选条件。\n\n**重要规则：**\n- `value_from` 和 `value` 必须同时使用，不能单独使用\n- `value_from` 当前仅支持 \"const\"（常量值）\n- 当使用 `value_from: \"const\"` 时，必须同时提供 `value` 字段\n",
                      "required": [
                        "operation"
                      ],
                      "properties": {
                        "value_from": {
                          "description": "字段值来源。\n\n**重要：** 当前仅支持 \"const\"（常量值），且必须与 `value` 字段同时使用\n",
                          "enum": [
                            "const"
                          ],
                          "type": "string"
                        },
                        "field": {
                          "type": "string",
                          "description": "字段名称，也即对象类的属性名称"
                        },
                        "operation": {
                          "type": "string",
                          "description": "查询条件操作符。\n**注意：** 虽然这里列出了所有可能的操作符，但每个对象类实际支持的操作符列表以对象类定义中的 `condition_operations` 字段为准。\n",
                          "enum": [
                            "and",
                            "or",
                            "==",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<=",
                            "in",
                            "not_in",
                            "like",
                            "not_like",
                            "exist",
                            "not_exist",
                            "match"
                          ]
                        },
                        "sub_conditions": {
                          "type": "array",
                          "description": "子过滤条件数组，用于逻辑操作符(and/or)的组合查询",
                          "items": {
                            "$ref": "#/components/schemas/Condition"
                          }
                        },
                        "value": {
                          "description": "字段值，格式根据操作符类型而定：\n- 比较操作符: 单个值\n- 范围查询: [min, max]数组\n- 集合操作: 值数组\n- 向量搜索: 特定格式数组\n\n**必须与 `value_from: \"const\"` 同时使用**\n"
                        }
                      }
                    },
                    "Parameter": {
                      "type": "object",
                      "description": "逻辑/指标参数",
                      "oneOf": [
                        {
                          "$ref": "#/components/schemas/Parameter4Operator"
                        },
                        {
                          "$ref": "#/components/schemas/Parameter4Metric"
                        }
                      ]
                    },
                    "Parameter4Metric": {
                      "description": "逻辑参数",
                      "required": [
                        "name",
                        "value_from",
                        "operation"
                      ],
                      "properties": {
                        "name": {
                          "type": "string",
                          "description": "参数名称"
                        },
                        "operation": {
                          "description": "操作符。映射指标模型的属性时，此字段必须",
                          "enum": [
                            "in",
                            "=",
                            "!=",
                            ">",
                            ">=",
                            "<",
                            "<="
                          ],
                          "type": "string"
                        },
                        "value": {
                          "type": "string",
                          "description": "参数值。value_from=property时，填入的是对象类的数据属性名称；value_from=input时，不设置此字段"
                        },
                        "value_from": {
                          "type": "string",
                          "description": "值来源",
                          "enum": [
                            "property",
                            "input"
                          ]
                        }
                      },
                      "type": "object"
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": null,
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770801820452636200,
            "update_time": 1770801820452636200,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "e378f603-3325-4752-b852-0c409b7fcbaf",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "66e07ec0-9212-4bd7-bafb-de185803ff7c",
            "name": "create_kn_index_build_job",
            "description": "创建一个全量构建业务知识网络的任务",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "d25512c9-33a6-43b2-b666-7f1c1e5c3c04",
              "summary": "创建业务知识网络构建任务",
              "description": "创建一个全量构建业务知识网络的任务",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_build_ontology",
              "method": "POST",
              "create_time": 1770868224368851700,
              "update_time": 1770868224368851700,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {
                    "application/json": {
                      "example": {
                        "kn_id": "kn_1234567890",
                        "name": "全量构建任务"
                      },
                      "schema": {
                        "$ref": "#/components/schemas/CreateJobRequest"
                      }
                    }
                  },
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "401",
                    "description": "未授权",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "201",
                    "description": "创建成功",
                    "content": {
                      "application/json": {
                        "example": {
                          "id": "job_1234567890"
                        },
                        "schema": {
                          "$ref": "#/components/schemas/CreateJobResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "CreateJobRequest": {
                      "type": "object",
                      "required": [
                        "kn_id",
                        "name"
                      ],
                      "properties": {
                        "kn_id": {
                          "description": "业务知识网络ID",
                          "type": "string"
                        },
                        "name": {
                          "description": "任务名称",
                          "type": "string"
                        }
                      }
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "link": {
                          "description": "错误链接",
                          "type": "string"
                        },
                        "solution": {
                          "description": "解决方案",
                          "type": "string"
                        },
                        "code": {
                          "type": "string",
                          "description": "错误码"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        },
                        "detail": {
                          "description": "错误详情",
                          "type": "object"
                        }
                      }
                    },
                    "CreateJobResponse": {
                      "type": "object",
                      "properties": {
                        "id": {
                          "type": "string",
                          "description": "任务ID"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "OntologyJob"
                ],
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770868224369224000,
            "update_time": 1770874561648186000,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "d25512c9-33a6-43b2-b666-7f1c1e5c3c04",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          },
          {
            "tool_id": "a5bdf17d-f6be-4e68-a005-1d4b20ed3398",
            "name": "get_kn_index_build_status",
            "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
            "status": "enabled",
            "metadata_type": "openapi",
            "metadata": {
              "version": "28bd5bd8-48a2-4290-a3a0-d5741010208e",
              "summary": "获取业务知识网络构建状态",
              "description": "查询最新50个构建任务的整体状态（按创建时间倒排）。如果所有任务都已完成则返回completed，如果有任务正在运行则返回running",
              "server_url": "http://agent-retrieval:30779",
              "path": "/api/agent-retrieval/in/v1/kn/full_ontology_building_status",
              "method": "GET",
              "create_time": 1770868224371221500,
              "update_time": 1770868224371221500,
              "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
              "api_spec": {
                "parameters": [
                  {
                    "name": "x-account-id",
                    "in": "header",
                    "description": "账户ID，用于内部服务调用时传递账户信息",
                    "required": false,
                    "schema": {
                      "type": "string"
                    }
                  },
                  {
                    "name": "x-account-type",
                    "in": "header",
                    "description": "账户类型：user(用户), app(应用), anonymous(匿名)",
                    "required": false,
                    "schema": {
                      "default": "user",
                      "enum": [
                        "user",
                        "app",
                        "anonymous"
                      ],
                      "type": "string"
                    }
                  },
                  {
                    "name": "kn_id",
                    "in": "query",
                    "description": "业务知识网络ID",
                    "required": true,
                    "schema": {
                      "type": "string"
                    }
                  }
                ],
                "request_body": {
                  "description": "",
                  "content": {},
                  "required": false
                },
                "responses": [
                  {
                    "status_code": "400",
                    "description": "参数错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "401",
                    "description": "未授权",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "500",
                    "description": "服务器内部错误",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/ErrorResponse"
                        }
                      }
                    }
                  },
                  {
                    "status_code": "200",
                    "description": "成功返回构建状态",
                    "content": {
                      "application/json": {
                        "example": {
                          "kn_id": "d5levlh818p1vl2slp60",
                          "state": "completed",
                          "state_detail": "All latest 50 jobs are completed"
                        },
                        "schema": {
                          "$ref": "#/components/schemas/BuildStatusSimpleResponse"
                        }
                      }
                    }
                  }
                ],
                "components": {
                  "schemas": {
                    "BuildStatusSimpleResponse": {
                      "type": "object",
                      "description": "构建状态响应",
                      "required": [
                        "kn_id",
                        "state",
                        "state_detail"
                      ],
                      "properties": {
                        "state": {
                          "description": "构建状态（running表示有任务正在运行，completed表示所有任务都已完成）",
                          "enum": [
                            "running",
                            "completed"
                          ],
                          "type": "string"
                        },
                        "state_detail": {
                          "type": "string",
                          "description": "状态详情"
                        },
                        "kn_id": {
                          "type": "string",
                          "description": "业务知识网络ID"
                        }
                      }
                    },
                    "ErrorResponse": {
                      "type": "object",
                      "properties": {
                        "code": {
                          "description": "错误码",
                          "type": "string"
                        },
                        "description": {
                          "type": "string",
                          "description": "错误描述"
                        },
                        "detail": {
                          "type": "object",
                          "description": "错误详情"
                        },
                        "link": {
                          "description": "错误链接",
                          "type": "string"
                        },
                        "solution": {
                          "type": "string",
                          "description": "解决方案"
                        }
                      }
                    }
                  }
                },
                "callbacks": null,
                "security": null,
                "tags": [
                  "OntologyJob"
                ],
                "external_docs": null
              }
            },
            "use_rule": "",
            "global_parameters": {
              "name": "",
              "description": "",
              "required": false,
              "in": "",
              "type": "",
              "value": null
            },
            "create_time": 1770868224371568600,
            "update_time": 1770874505570531800,
            "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
            "extend_info": null,
            "resource_object": "tool",
            "source_id": "28bd5bd8-48a2-4290-a3a0-d5741010208e",
            "source_type": "openapi",
            "script_type": "",
            "code": ""
          }
        ],
        "create_time": 1770778843187375600,
        "update_time": 1770874613717222400,
        "create_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "update_user": "ede150ba-06f4-11f1-85aa-3a34099a4c4b",
        "metadata_type": "openapi"
      }
    ]
  }
}