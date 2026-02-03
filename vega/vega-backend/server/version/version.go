package version

import (
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/audit"
)

var (
	ServerName    string = "vega-backend"
	ServerVersion string = "1.0.0"
	LanguageGo    string = "go"
	GoVersion     string = runtime.Version()
	GoArch        string = runtime.GOARCH
)

func init() {
	audit.DEFAULT_AUDIT_LOG_FROM = audit.AuditLogFrom{
		Package: "Vega",
		Service: audit.AuditLogFromService{
			Name: "vega-backend",
		},
	}
}
