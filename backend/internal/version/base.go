package version

var (
	// These variables are replaced by ldflags at build time
	gitVersion = "v0.0.0-main"
	gitCommit  = ""
	buildDate  = "1970-01-01T00:00:00Z" // build date in ISO8601 format
)
