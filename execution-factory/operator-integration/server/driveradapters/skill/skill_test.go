package skill

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestSkillHandler(t *testing.T) {
	Convey("SkillHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("RegisterSkill binds multipart form and calls registry", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Reader:   mockReader,
			}
			mockRegistry.EXPECT().RegisterSkill(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.RegisterSkillReq) (*interfaces.RegisterSkillResp, error) {
					So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.UserID, ShouldEqual, "user-1")
					So(req.FileType, ShouldEqual, "content")
					return &interfaces.RegisterSkillResp{SkillID: "skill-1", Status: "active"}, nil
				},
			)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			So(writer.WriteField("file_type", "content"), ShouldBeNil)
			filePart, err := writer.CreateFormFile("file", "SKILL.md")
			So(err, ShouldBeNil)
			_, err = filePart.Write([]byte("---\nname: demo\ndescription: desc\n---\nbody"))
			So(err, ShouldBeNil)
			So(writer.Close(), ShouldBeNil)

			recorder := performSkillRequest(http.MethodPost, "/skills", writer.FormDataContentType(), body.String(), map[string]string{
				"x-business-domain": "bd-test",
				"user_id":           "user-1",
			}, handler.RegisterSkill)

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"skill_id":"skill-1"`)
		})

		Convey("RegisterSkill rejects unsupported content type", func() {
			handler := &skillHandler{}
			recorder := performSkillRequest(http.MethodPost, "/skills", "text/plain", "raw", map[string]string{
				"x-business-domain": "bd-test",
				"user_id":           "user-1",
			}, handler.RegisterSkill)

			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("GetSkillGuide binds uri and calls reader", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Reader:   mockReader,
			}
			mockReader.EXPECT().GetSkillGuide(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.GetSkillGuideReq) (*interfaces.GetSkillGuideResp, error) {
					So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-2")
					return &interfaces.GetSkillGuideResp{SkillID: "skill-2", Content: "guide"}, nil
				},
			)

			recorder := performSkillRequest(http.MethodGet, "/skills/:skill_id/guide", "", "", map[string]string{
				"x-business-domain": "bd-test",
			}, handler.GetSkillGuide, "skill-2")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"content":"guide"`)
		})

		Convey("ReadSkillFile binds body and calls reader", func() {
			mockRegistry := mocks.NewMockSkillRegistry(ctrl)
			mockReader := mocks.NewMockSkillReader(ctrl)
			handler := &skillHandler{
				Registry: mockRegistry,
				Reader:   mockReader,
			}
			mockReader.EXPECT().ReadSkillFile(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ any, req *interfaces.ReadSkillFileReq) (*interfaces.ReadSkillFileResp, error) {
					So(req.BusinessDomainID, ShouldEqual, "bd-test")
					So(req.SkillID, ShouldEqual, "skill-3")
					So(req.RelPath, ShouldEqual, "refs/guide.md")
					return &interfaces.ReadSkillFileResp{SkillID: "skill-3", RelPath: "refs/guide.md", Content: "body"}, nil
				},
			)

			recorder := performSkillRequest(http.MethodPost, "/skills/:skill_id/files/read", "application/json", `{"rel_path":"refs/guide.md"}`, map[string]string{
				"x-business-domain": "bd-test",
			}, handler.ReadSkillFile, "skill-3")

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.String(), ShouldContainSubstring, `"content":"body"`)
		})
	})
}

func performSkillRequest(method, path, contentType, body string, headers map[string]string, handler func(c *gin.Context), pathParams ...string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Handle(method, path, func(c *gin.Context) {
		for i, param := range pathParams {
			paramName := strings.Split(path, "/")[i+1][1:]
			c.Params = append(c.Params, gin.Param{Key: paramName, Value: param})
		}
		handler(c)
	})

	formattedPath := path
	for _, param := range pathParams {
		start := strings.Index(formattedPath, ":")
		if start == -1 {
			break
		}
		end := strings.Index(formattedPath[start:], "/")
		if end == -1 {
			end = len(formattedPath)
		} else {
			end += start
		}
		formattedPath = formattedPath[:start] + param + formattedPath[end:]
	}

	req := httptest.NewRequest(method, formattedPath, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
