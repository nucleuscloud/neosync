package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	whitelist = map[string]bool{
		"buf":  true,
		"mgmt": true,
	}
	prefix = "neosync."
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a directory path")
		return
	}

	dir := os.Args[1]
	err := filepath.Walk(dir, processFile)
	if err != nil {
		fmt.Printf("Error walking through directory: %v\n", err)
	}
}

func processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if !strings.HasSuffix(path, ".py") {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	modified := false
	var newLines []string

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			newLine, changed := processImportLine(line)
			if changed {
				modified = true
				line = newLine
			}
		}
		newLines = append(newLines, line)
	}

	if modified {
		newContent := strings.Join(newLines, "\n")
		err = os.WriteFile(path, []byte(newContent), info.Mode())
		if err != nil {
			return err
		}
		fmt.Printf("Modified imports in: %s\n", path)
	}

	return nil
}

func processImportLine(line string) (string, bool) {
	for module := range whitelist {
		// Pattern for "import module" or "from module import ..."
		importPattern := regexp.MustCompile(`^(from\s+|import\s+)(` + module + `)(\s|\.|\.)`)
		if importPattern.MatchString(line) {
			return importPattern.ReplaceAllString(line, "${1}"+prefix+"${2}${3}"), true
		}
	}
	return line, false
}
