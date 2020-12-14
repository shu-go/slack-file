package main

import (
	"errors"
	"sort"

	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

type listCmd struct {
	_ struct{} `help:"list files"`

	Order gli.StrList `default:"Name,-Created,-Timestamp,ID" help:"order of the files"`
	Group gli.StrList `default:"" help:"e.g. Channels,Groups,IMs"`
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
	sortProps = append(sortProps, c.Order...)
	sort.Slice(files, func(i, j int) bool {
		c := filePropsCompare(files[i], files[j], sortProps)
		return c < 0
	})

	var prev *slack.File
	for _, f := range files {
		if prev == nil || filePropsCompare(*prev, f, c.Group) != 0 {
			temp := f
			prev = &temp

			for _, g := range c.Group {
				println(fileProp(f, g))
			}
		}
		println(f.Name)
	}

	return nil
}
