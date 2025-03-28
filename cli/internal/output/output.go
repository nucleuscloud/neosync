package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	autoOutput  OutputType = "auto"
	PlainOutput OutputType = "plain"
	TtyOutput   OutputType = "tty"
)

var (
	outputMap = map[string]OutputType{
		string(autoOutput):  autoOutput,
		string(PlainOutput): PlainOutput,
		string(TtyOutput):   TtyOutput,
	}
)

type OutputType string

func AttachOutputFlag(cmd *cobra.Command) {
	outputVals := []string{}
	for outputType := range outputMap {
		outputVals = append(outputVals, outputType)
	}

	cmd.Flags().
		StringP("output", "o", string(autoOutput), fmt.Sprintf("Set type of output (%s).", strings.Join(outputVals, ", ")))
}

func ValidateAndRetrieveOutputFlag(cmd *cobra.Command) (OutputType, error) {
	if cmd == nil {
		return "", fmt.Errorf("must provide non-nil cmd")
	}
	outputFlag, err := cmd.Flags().GetString("output")
	if err != nil {
		return "", err
	}

	p, ok := parseOutputString(outputFlag)
	if !ok {
		return "", fmt.Errorf("must provide valid progress type")
	}
	if p != autoOutput {
		return p, nil
	}
	if isGithubAction() || !IsTerminal() {
		return PlainOutput, nil
	}
	return TtyOutput, nil
}

func parseOutputString(str string) (OutputType, bool) {
	p, ok := outputMap[strings.ToLower(str)]
	return p, ok
}

func isGithubAction() bool {
	val := os.Getenv("GITHUB_ACTIONS")
	return val == "true"
}

func GetStdoutFd() int {
	return int(os.Stdout.Fd())
}

func IsTerminal() bool {
	return term.IsTerminal(GetStdoutFd())
}
