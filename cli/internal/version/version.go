package version

import (
	"fmt"
	"runtime"

	"github.com/nucleuscloud/neosync/backend/pkg/utils"
	"google.golang.org/grpc/metadata"
)

type VersionInfo struct {
	GitVersion string `json:"gitVersion" yaml:"gitVersion"`
	GitCommit  string `json:"gitCommit"  yaml:"gitCommit"`
	BuildDate  string `json:"buildDate"  yaml:"buildDate"`
	GoVersion  string `json:"goVersion"  yaml:"goVersion"`
	Compiler   string `json:"compiler"   yaml:"compiler"`
	Platform   string `json:"platform"   yaml:"platform"`
}

func (info *VersionInfo) String() string {
	return info.GitVersion
}

func (info *VersionInfo) Headers() map[string]string {
	return map[string]string{
		utils.CliVersionKey:  info.GitVersion,
		utils.CliPlatformKey: info.Platform,
		utils.CliCommitKey:   info.GitCommit,
		"User-Agent":         constructUserAgent(info),
	}
}

func constructUserAgent(info *VersionInfo) string {
	return fmt.Sprintf(
		"neosync/%s (commit: %s; build: %s; go: %s; compiler: %s; platform: %s)",
		info.GitVersion,
		info.GitCommit,
		info.BuildDate,
		info.GoVersion,
		info.Compiler,
		info.Platform,
	)
}

func (info *VersionInfo) GrpcMetadata() metadata.MD {
	return metadata.New(info.Headers())
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() *VersionInfo {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in ./base.go
	return &VersionInfo{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
