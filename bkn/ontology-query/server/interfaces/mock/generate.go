package mock_interfaces

// Local mocks for kweaver-go-lib interfaces.
// Needed because kweaver-go-lib still uses deprecated github.com/golang/mock.
// TODO: Remove when kweaver-go-lib migrates to go.uber.org/mock.

//go:generate mockgen -destination=mock_http_client.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/rest HTTPClient
//go:generate mockgen -destination=mock_hydra.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/rest Hydra
