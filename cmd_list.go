package main

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

type listCmd struct {
	_ struct{} `help:"list files" usage:"# list all\nslack-file list\n# find by pattern\nslack-file list my*.txt\n# files older than 1day\nslack-file list --older 24h\n# files in a general channel\nslack-file list --chan general"`

	Target gli.StrList   `default:"Name,Title,ID"`
	Older  time.Duration `cli:"older-than,older" help:"Timestamp (e.g. '24h' for 1-day)"`

	Chan      string `help:"a channel name"`
	innerChan string

	Sort  gli.StrList `default:"Name,-Timestamp,ID" help:"sort fields"`
	Group gli.StrList `default:"" help:"e.g. Channels,Groups,IMs"`

	Format string `default:"{{.ID}}\t{{.Timestamp.Time}}\t{{.Name}}"`
}

func init() {
	gApp.AddExtraCommand(&listCmd{}, "list,ls", "")
}

func (c *listCmd) Before(global globalCmd) error {
	if c.Chan != "" {
		config, _ := loadConfig(global.Config)

		if config.Slack.AccessToken == "" {
			return errors.New("auth first")
		}

		sl := slack.New(config.Slack.AccessToken)

		params := slack.GetConversationsForUserParameters{
			Types: []string{"public_channel", "private_channel"},
		}
		chans, err := listConversationsForUser(sl, params)
		if err != nil {
			return err
		}

		for _, ch := range chans {
			if strings.ToLower(c.Chan) == strings.ToLower(ch.Name) {
				c.innerChan = ch.ID
			}
		}

		if c.innerChan == "" {
			return errors.New("no channel " + c.Chan + " found")
		}
	}
	return nil
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

	useOlder := false
	oldTimestamp := time.Now()
	if c.Older != time.Duration(0) {
		useOlder = true
		oldTimestamp = time.Now().Add(-c.Older)
	}

	var prev *slack.File
	for _, f := range files {
		if useOlder && !f.Timestamp.Time().Before(oldTimestamp) {
			continue
		}

		matched := false
		for _, p := range patterns {
			for _, tgt := range c.Target {
				if p.Match(fileProp(f, tgt)) {
					matched = true
					break
				}
			}
		}
		if len(patterns) != 0 && !matched {
			continue
		}

		if c.innerChan != "" {
			found := false
			for _, fc := range f.Channels {
				if c.innerChan == fc {
					found = true
					break
				}
			}
			for _, fc := range f.Groups {
				if c.innerChan == fc {
					found = true
					break
				}
			}

			if !found {
				continue
			}
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
