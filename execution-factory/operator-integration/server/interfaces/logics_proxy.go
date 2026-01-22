package interfaces

import (
	"context"
	"io"
)

// ProxyHandler 代理处理器
//
//go:generate mockgen -source=logics_proxy.go -destination=../mocks/logics_proxy.go -package=mocks
type ProxyHandler interface {
	HandlerRequest(ctx context.Context, req *HTTPRequest) (resp *HTTPResponse, err error)
}

// IOutboxMessageEvent 消息事件管理
type IOutboxMessageEvent interface {
	Publish(ctx context.Context, req *OutboxMessageReq) (err error)
}

// Forwarder 转发器接口
type Forwarder interface {
	Forward(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
	ForwardStream(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
}

// StreamProcessor 流式处理器接口
type StreamProcessor interface {
	ProcessSSE(ctx context.Context, reader io.Reader, writer io.Writer) error
	ProcessHTTPStream(ctx context.Context, reader io.Reader, writer io.Writer) error
}
