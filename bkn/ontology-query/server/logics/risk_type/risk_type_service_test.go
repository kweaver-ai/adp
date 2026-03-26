// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package risk_type

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	"ontology-query/interfaces"
	dtype "ontology-query/interfaces/data_type"
	dmock "ontology-query/interfaces/mock"
)

func Test_riskTypeService_Evaluate(t *testing.T) {
	Convey("Test riskTypeService Evaluate - 通过内置工具执行风险评估", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		omAccess := dmock.NewMockOntologyManagerAccess(mockCtrl)
		aoAccess := dmock.NewMockAgentOperatorAccess(mockCtrl)
		service := &riskTypeService{
			appSetting: &common.AppSetting{},
			omAccess:   omAccess,
			aoAccess:   aoAccess,
		}
		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		riskTypeID := "rt1"

		Convey("无 risk_type_configs 时直接 Allow", func() {
			actionType := &interfaces.ActionType{
				ATID:            "at1",
				RiskTypeConfigs: nil,
			}
			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue)
		})

		Convey("risk_type_configs 为空列表时 Allow", func() {
			actionType := &interfaces.ActionType{
				ATID:            "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{},
			}
			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue)
		})

		Convey("无 when 时取 rule.Decision 更新 maxLevel", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{ID: "r1", When: nil, Decision: RiskLevelHigh},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelHigh, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue)
		})

		Convey("when.Type==condition 且条件满足时命中规则", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER, Value: 100}}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type: "condition",
							Condition: &cond.CondCfg{
								Name:        "amount",
								Operation:   cond.OperationGte,
								ValueOptCfg: cond.ValueOptCfg{Value: 50},
							},
						},
						Decision: RiskLevelMedium,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelMedium, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue)
		})

		Convey("when.Type==condition 且条件不满足时不更新 maxLevel", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER, Value: 10}}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type: "condition",
							Condition: &cond.CondCfg{
								Name:        "amount",
								Operation:   cond.OperationGte,
								ValueOptCfg: cond.ValueOptCfg{Value: 50},
							},
						},
						Decision: RiskLevelHigh,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelSafe, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue) // amount=10 不满足 >=50，maxLevel 保持 safe
		})

		Convey("when.Type==natural_language 时不命中", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type:            "natural_language",
							NaturalLanguage: "金额大于100",
						},
						Decision: RiskLevelHigh,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelSafe, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue) // natural_language 不评估，maxLevel 保持 safe
		})

		Convey("when.Condition==nil 时视为满足", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelMedium,
				Parameters:         []interfaces.Parameter{},
				RiskRules: []interfaces.RiskRule{
					{
						ID:       "r1",
						When:     &interfaces.RiskRuleWhen{Type: "condition", Condition: nil},
						Decision: RiskLevelMedium,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelMedium, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue)
		})

		Convey("cfg.Parameters==nil 时使用空 map 评估", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: nil},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type: "condition",
							Condition: &cond.CondCfg{
								Name:        "amount",
								Operation:   cond.OperationEq,
								ValueOptCfg: cond.ValueOptCfg{Value: 100},
							},
						},
						Decision: RiskLevelLow,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelSafe, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue) // data 为空，amount 不存在，条件不满足，maxLevel 保持 safe
		})

		Convey("Params 优先于 Parameters 用于评估", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{
						RiskTypeID: riskTypeID,
						Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER, Value: 10}},
						Params:     map[string]any{"amount": 100},
					},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelCritical,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type: "condition",
							Condition: &cond.CondCfg{
								Name:        "amount",
								Operation:   cond.OperationGte,
								ValueOptCfg: cond.ValueOptCfg{Value: 50},
							},
						},
						Decision: RiskLevelMedium,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelMedium, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldBeNil)
			So(result.Allow, ShouldBeTrue) // Params.amount=100 满足 >=50，使用 Params 而非 Parameters.amount=10
		})

		Convey("风险等级超过 max_acceptable_level 时返回 Allow=false 且 err 不为 nil", func() {
			actionType := &interfaces.ActionType{
				ATID: "at1",
				RiskTypeConfigs: []interfaces.RiskTypeConfig{
					{RiskTypeID: riskTypeID, Parameters: []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER, Value: 1000}}},
				},
			}
			riskType := interfaces.RiskType{
				RTID:               riskTypeID,
				MaxAcceptableLevel: RiskLevelLow,
				Parameters:         []interfaces.Parameter{{Name: "amount", Type: dtype.DATATYPE_INTEGER}},
				RiskRules: []interfaces.RiskRule{
					{
						ID: "r1",
						When: &interfaces.RiskRuleWhen{
							Type: "condition",
							Condition: &cond.CondCfg{
								Name:        "amount",
								Operation:   cond.OperationGte,
								ValueOptCfg: cond.ValueOptCfg{Value: 100},
							},
						},
						Decision: RiskLevelHigh,
					},
				},
			}
			omAccess.EXPECT().GetRiskTypesByIDs(gomock.Any(), knID, branch, []string{riskTypeID}).Return([]interfaces.RiskType{riskType}, nil)
			aoAccess.EXPECT().ExecuteTool(gomock.Any(), interfaces.BuiltinToolBoxID, interfaces.BuiltinToolToolID, gomock.Any()).
				Return(map[string]any{"result": map[string]any{"risk_level": RiskLevelHigh, "success": true, "message": ""}}, nil)

			result, err := service.Evaluate(ctx, actionType, knID, branch)
			So(err, ShouldNotBeNil)
			So(result, ShouldNotBeNil)
			So(result.Allow, ShouldBeFalse)
			So(result.Message, ShouldContainSubstring, "exceeds max_acceptable_level")
		})
	})
}
