package testdoubles

// ServerDeps aggregates all outbound-port fakes for unit tests.
// Ports will be defined as server/ and client/ are migrated to
// internal/{ports,adapters,app,domain}.
type ServerDeps struct{}

// NewServerDeps returns a ServerDeps with all fakes initialised to safe zero-value defaults.
func NewServerDeps() *ServerDeps {
	return &ServerDeps{}
}
