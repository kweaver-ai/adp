package actions

import (
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/ide-go-lib/telemetry/trace"
)

const (
	OpReturn            = "@internal/return"
	ReturnResultSuccess = "success"
	ReturnResultFailed  = "failed"
)

type Return struct {
}

type ReturnParam struct {
	Result string `json:"result"`
}

func (a *Return) Name() string {
	return OpReturn
}

func (r *Return) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*ReturnParam)

	return input.Result, nil
}

func (r *Return) ParameterNew() interface{} {
	return &ReturnParam{}
}
