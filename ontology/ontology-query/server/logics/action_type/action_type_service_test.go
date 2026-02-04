package action_type

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	dmock "ontology-query/interfaces/mock"
	"ontology-query/logics"
)

func Test_NewActionTypeService(t *testing.T) {
	Convey("Test NewActionTypeService", t, func() {
		appSetting := &common.AppSetting{}

		Convey("成功 - 创建服务实例", func() {
			service := NewActionTypeService(appSetting)
			So(service, ShouldNotBeNil)
		})

		Convey("成功 - 单例模式", func() {
			service1 := NewActionTypeService(appSetting)
			service2 := NewActionTypeService(appSetting)
			So(service1, ShouldEqual, service2)
		})
	})
}

func Test_actionTypeService_GetActionsByActionTypeID(t *testing.T) {
	Convey("Test actionTypeService GetActionsByActionTypeID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := dmock.NewMockOntologyManagerAccess(mockCtrl)
		ots := dmock.NewMockObjectTypeService(mockCtrl)
		uAccess := dmock.NewMockUniqueryAccess(mockCtrl)

		// 设置全局变量
		logics.OMA = omAccess
		logics.UA = uAccess

		service := &actionTypeService{
			appSetting: appSetting,
			omAccess:   omAccess,
			ots:        ots,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"
		actionTypeID := "at1"
		objectTypeID := "ot1"

		Convey("成功 - 获取行动数据", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "param1",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "prop1",
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{
						"id":    "123",
						"prop1": "value1",
						"prop2": "value2",
					},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(len(result.Actions), ShouldEqual, 1)
			So(result.Actions[0].Parameters["param1"], ShouldEqual, "value1")
		})

		Convey("失败 - 行动类不存在", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ActionType{}, false, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("失败 - 获取行动类错误", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ActionType{}, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("成功 - 带行动条件", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				// 验证条件是否正确合并
				So(query.ActualCondition, ShouldNotBeNil)
				So(query.ActualCondition.Operation, ShouldEqual, "and")
				return objects, nil
			})

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
		})

		Convey("成功 - 参数来源为常量", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "const_param",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     "constant_value",
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.Actions[0].Parameters["const_param"], ShouldEqual, "constant_value")
		})

		Convey("成功 - 参数来源为输入", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "input_param",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.Actions[0].DynamicParams["input_param"], ShouldBeNil)
		})

		Convey("成功 - 包含类型信息", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeTypeInfo: true,
				},
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.ActionType, ShouldNotBeNil)
			So(result.ActionType.ATID, ShouldEqual, actionTypeID)
		})

		Convey("成功 - 多个对象", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
					{"id": "456"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
					{"id": "456"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(len(result.Actions), ShouldEqual, 2)
		})

		Convey("失败 - 获取对象数据错误", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("成功 - 空对象列表", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(len(result.Actions), ShouldEqual, 0)
		})
	})
}
