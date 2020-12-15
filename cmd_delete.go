package main

import (
	"errors"
	"sort"
	"time"

	"github.com/gobwas/glob"
	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

type deleteCmd struct {
	_ struct{} `help:"delete files" usage:"# delete my*.txt\nslack-file delete my*.txt\n# delete 2days-older files\nslack-file delete --older 48h *"`

	Target gli.StrList   `default:"Name,Title,ID"`
	Older  time.Duration `cli:"older-than,older" help:"Timestamp (e.g. '24h' for 1-day)"`
	DryRun bool          `cli:"dry-run" help:"do not delete files actually"`

	Format string `default:"{{.ID}}\t{{.Timestamp.Time}}\t{{.Name}}"`
}

func init() {
	gApp.AddExtraCommand(&deleteCmd{}, "delete,del,remove,rm", "")
}

func (c deleteCmd) Run(global globalCmd, args []string) error {
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
	sortProps = append(sortProps, c.Target...)
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
		if !matched {
			continue
		}

		s, err := fileToString(c.Format, f)
		if err != nil {
			return err
		}
		println(s)

		if !c.DryRun {
			err := sl.DeleteFile(f.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
