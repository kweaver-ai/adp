package mock_interfaces

// Local mocks for kweaver-go-lib interfaces (not shipped with generated mocks).

//go:generate mockgen -destination=mock_http_client.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/rest HTTPClient
//go:generate mockgen -destination=mock_hydra.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/hydra Hydra
