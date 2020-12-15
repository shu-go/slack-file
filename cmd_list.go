package main

import (
	"errors"
	"sort"

	"github.com/gobwas/glob"
	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

type listCmd struct {
	_ struct{} `help:"list files"`

	Sort  gli.StrList `default:"Name,-Created,-Timestamp,ID" help:"sort fields"`
	Group gli.StrList `default:"" help:"e.g. Channels,Groups,IMs"`

	Format string `default:"{{.ID}}\t{{.Timestamp.Time}}\t{{.Name}}"`
}

func init() {
	gApp.AddExtraCommand(&listCmd{}, "list,ls", "")
}

func (c listCmd) Run(global globalCmd, args []string) error {
	config, _ := loadConfig(global.Config)

	if config.Slack.AccessToken == "" {
		return errors.New("auth first")
	}

	sl := slack.New(config.Slack.AccessToken)

	files, err := listFiles(sl, slack.ListFilesParameters{
		Limit: 10,
	})
	if err != nil {
		return err
	}

	var sortProps []string
	sortProps = append(sortProps, c.Group...)
	sortProps = append(sortProps, c.Sort...)
	sort.Slice(files, func(i, j int) bool {
		c := filePropsCompare(files[i], files[j], sortProps)
		return c < 0
	})

	var patterns []glob.Glob
	for _, a := range args {
		patterns = append(patterns, glob.MustCompile(a))
	}

	var prev *slack.File
	for _, f := range files {
		matched := false
		for _, p := range patterns {
			if p.Match(f.Name) {
				matched = true
				break
			}
		}
		if len(patterns) != 0 && !matched {
			continue
		}

		if prev == nil || filePropsCompare(*prev, f, c.Group) != 0 {
			if prev != nil {
				println("")
			}
			for _, g := range c.Group {
				println(fileProp(f, g))
			}

			temp := f
			prev = &temp
		}
		s, err := fileToString(c.Format, f)
		if err != nil {
			return err
		}
		println(s)
	}

	return nil
}
