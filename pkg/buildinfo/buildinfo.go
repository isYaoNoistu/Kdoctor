package buildinfo

import "fmt"

var (
	Version = "v2.0.0"
	Commit  = "dev"
)

func ToolVersion() string {
	return Version
}

func FullVersion() string {
	if Commit == "" || Commit == "dev" {
		return Version
	}
	return fmt.Sprintf("%s (%s)", Version, Commit)
}
