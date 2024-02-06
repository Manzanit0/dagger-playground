package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/Masterminds/vcs"
)

var (
	gitRepository = flag.String("git-repository", "", "git repository to build from")
)

func main() {
	flag.Parse()

	local := os.TempDir()
	repo, err := vcs.NewRepo(*gitRepository, local)
	if err != nil {
		panic(err)
	}

	err = repo.Get()
	if err != nil {
		panic(err)
	}

	tags, err := repo.Tags()
	if err != nil {
		panic(err)
	}

	valid := []*semver.Version{}
	for _, tag := range tags {
		tag = strings.TrimPrefix(tag, "v")

		v, err := semver.NewVersion(tag)
		if err != nil {
			fmt.Println("invalid version", err.Error())
			continue
		}

		valid = append(valid, v)
	}
	if len(valid) == 0 {
		fmt.Println("No valid tags")
		return
	}

	sort.Sort(semver.Collection(valid))

	fmt.Println("next version should be", valid[len(valid)-1].IncMinor().String())

	// TODO: tag HEAD and push.
}
