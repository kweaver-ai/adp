package mock_interfaces

// Local mocks for kweaver-go-lib interfaces.
// Needed because kweaver-go-lib has not yet migrated to go.uber.org/mock.
// TODO: Remove these local mocks once kweaver-go-lib completes its migration.

//go:generate mockgen -destination=mock_http_client.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/rest HTTPClient
//go:generate mockgen -destination=mock_hydra.go -package=mock_interfaces github.com/kweaver-ai/kweaver-go-lib/rest Hydra
