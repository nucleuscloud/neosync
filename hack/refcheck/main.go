package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/modfile"
)

const (
	gitPrefix  = "github.com/nucleuscloud/neosync"
	goModName  = "go.mod"
	goWorkName = "go.work"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	gitRepo, err := git.PlainOpen(".")
	if err != nil {
		panic(fmt.Errorf("unable to open git repo: %w", err))
	}

	parentRef := "refs/heads/main"
	if len(os.Args) == 2 {
		parentRef = os.Args[1]
	}
	logger.Info(fmt.Sprintf("parent ref %s", parentRef))
	branchRefName := plumbing.ReferenceName(parentRef)
	branchRef, err := gitRepo.Reference(branchRefName, true)
	if err != nil {
		panic(fmt.Errorf("unable to find reference for %s: %w", branchRefName.String(), err))
	}
	branchRefObject, err := gitRepo.CommitObject(branchRef.Hash())
	if err != nil {
		panic(fmt.Errorf("unable to find commit object for branch ref %s: %w", branchRef.Hash(), err))
	}

	invalidVersions := map[string][]string{}
	for gomodpath, gitVersions := range getGitShasByGoMod() {
		for _, gitVersion := range gitVersions {
			ref, err := gitRepo.ResolveRevision(plumbing.Revision(gitVersion))
			if err != nil {
				panic(fmt.Errorf("unable to resolve revision %s: %w", gitVersion, err))
			}
			commitObject, err := gitRepo.CommitObject(plumbing.NewHash(ref.String()))
			if err != nil {
				panic(fmt.Errorf("unable to find commit object %s: %w", ref.String(), err))
			}

			ok, err := commitObject.IsAncestor(branchRefObject)
			if err != nil {
				panic(fmt.Errorf("unable to check if commit object is ancestor of branch: %w", err))
			}
			if !ok {
				invalidVersions[gomodpath] = append(invalidVersions[gomodpath], gitVersion)
			}
		}
	}

	ok := len(invalidVersions) == 0
	for gomodpath, badVersions := range invalidVersions {
		logger.Info(fmt.Sprintf("bad versions (%s) in module: %s", strings.Join(badVersions, ","), gomodpath))
	}
	if !ok {
		panic("found invalid versions")
	}
}

func getGitShasByGoMod() map[string][]string {
	// example: v0.0.0-20231024190303-a76ecf39fdcd
	regex, err := regexp.Compile(`v\d+\.\d+\.\d+\-\d+\-([a-zA-Z0-9]+)`) //nolint
	if err != nil {
		panic(err)
	}

	gitVersions := map[string][]string{}
	for gomodpath, modVersions := range getGoModVersions() {
		for _, version := range modVersions {
			output := regex.FindStringSubmatch(version)
			if len(output) != 2 {
				continue
			}
			gitVersions[gomodpath] = append(gitVersions[gomodpath], output[1])
		}
	}
	return gitVersions
}

func getGoModVersions() map[string][]string {
	workbits, err := os.ReadFile(goWorkName)
	if err != nil {
		panic(err)
	}
	wf, err := modfile.ParseWork(goWorkName, workbits, nil)
	if err != nil {
		panic(err)
	}

	versions := map[string][]string{}

	for _, use := range wf.Use {
		fullGoModPath := fmt.Sprintf("%s/%s", use.Path, goModName)
		gomodbits, err := os.ReadFile(fullGoModPath)
		if err != nil {
			panic(err)
		}
		gm, err := modfile.Parse(goModName, gomodbits, nil)
		if err != nil {
			panic(err)
		}
		for _, req := range gm.Require {
			if strings.HasPrefix(req.Mod.Path, gitPrefix) {
				versions[fullGoModPath] = append(versions[fullGoModPath], req.Mod.Version)
			}
		}
	}

	return versions
}
