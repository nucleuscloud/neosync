package pyroscope_env

import (
	"errors"
	"log/slog"
	"runtime"

	"github.com/grafana/pyroscope-go"
	pyroscope_logger "github.com/nucleuscloud/neosync/internal/pyroscope/logger"
	"github.com/spf13/viper"
)

func NewFromEnv(applicationName string, logger *slog.Logger) (*pyroscope.Config, bool, error) {
	isPyroscopeEnabled := viper.GetBool("PYROSCOPE_ENABLED")
	if !isPyroscopeEnabled {
		return nil, false, nil
	}

	logger.Debug("pyroscope is enabled")
	serverAddress := viper.GetString("PYROSCOPE_SERVER_ADDRESS")
	if serverAddress == "" {
		return nil, false, errors.New("PYROSCOPE_SERVER_ADDRESS is required")
	}
	basicAuthUser := viper.GetString("PYROSCOPE_BASIC_AUTH_USER")
	basicAuthPassword := viper.GetString("PYROSCOPE_BASIC_AUTH_PASSWORD")

	pyroscopeTags := map[string]string{}

	isLoggerEnabled := viper.GetBool("PYROSCOPE_LOGGER_ENABLED")
	var pyroscopeLogger pyroscope.Logger
	if isLoggerEnabled {
		pyroscopeLogger = pyroscope_logger.New(logger)
	} else {
		pyroscopeLogger = pyroscope_logger.NewNoop()
	}

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	pyroscopeConfig := &pyroscope.Config{
		ApplicationName:   applicationName,
		ServerAddress:     serverAddress,
		Logger:            pyroscopeLogger,
		BasicAuthUser:     basicAuthUser,
		BasicAuthPassword: basicAuthPassword,
		Tags:              pyroscopeTags,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
		DisableGCRuns: true,
	}

	return pyroscopeConfig, true, nil
}
