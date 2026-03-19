package skill

import (
	"context"
	"testing"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillGovernanceContract(t *testing.T) {
	Convey("Skill governance contract", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("Skill registry should expose formal auth resource type", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}

			So(registry.resourceType(), ShouldEqual, interfaces.AuthResourceTypeSkill)
			So(registry.resourceType().String(), ShouldEqual, "skill")
		})

		Convey("Skill registry should use skill resource type when binding business domain", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}

			mockBusinessDomainService.EXPECT().
				AssociateResource(gomock.Any(), "bd-1", "skill-1", interfaces.AuthResourceTypeSkill).
				Return(nil).
				Times(1)

			err := registry.associateBusinessDomain(context.Background(), "bd-1", "skill-1")
			So(err, ShouldBeNil)
		})

		Convey("Skill registry should use skill resource type when filtering market visible ids", func() {
			mockAuthService := mocks.NewMockIAuthorizationService(ctrl)
			mockBusinessDomainService := mocks.NewMockIBusinessDomainService(ctrl)
			registry := &skillRegistry{
				AuthService:           mockAuthService,
				BusinessDomainService: mockBusinessDomainService,
			}
			accessor := &interfaces.AuthAccessor{ID: "user-1"}

			mockAuthService.EXPECT().
				ResourceListIDs(gomock.Any(), accessor, interfaces.AuthResourceTypeSkill, interfaces.AuthOperationTypePublicAccess).
				Return([]string{"skill-1", "skill-2"}, nil).
				Times(1)

			ids, err := registry.listMarketVisibleIDs(context.Background(), accessor)
			So(err, ShouldBeNil)
			So(ids, ShouldResemble, []string{"skill-1", "skill-2"})
		})
	})
}
