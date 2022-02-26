package build

import "runtime"

// build parameters
var (
	Branch    string
	GoVersion = runtime.Version()
	Revision  string
	Version   string
)
