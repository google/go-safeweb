package safehttp

var (
	isLocalDev bool
	// freezeLocalDev is set on Mux construction.
	freezeLocalDev bool
)

// UseLocalDev instructs the framework to disable some security mechanisms that
// would make local development hard or impossible. This cannot be undone without
// restarting the program and should only be done before any other function or type
// of the framework is used.
// This function should ideally be called by the main package immediately after
// flag parsing.
// This configuration is not valid for production use.
func UseLocalDev() {
	if freezeLocalDev {
		panic("UseLocalDev should be called before any other part of the framework")
	}
	isLocalDev = true
}

// IsLocalDev returns whether the framework is set up to use local development
// rules. Please see the doc on UseLocalDev.
func IsLocalDev() bool {
	return isLocalDev
}
