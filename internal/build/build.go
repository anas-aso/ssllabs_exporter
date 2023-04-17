package build

import "runtime"

// build parameters
var (
	Branch    = "dev"
	GoVersion = runtime.Version()
	Revision  = "n/a"
	Version   = "n/a"
)
