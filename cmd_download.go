package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

type downloadCmd struct {
	_ struct{} `help:"download a file" usage:""`

	Output *string `cli:"output,o=FILE_NAME"`

	Target gli.StrList   `default:"Name,Title,ID"`
	Older  time.Duration `cli:"older-than,older" help:"Timestamp (e.g. '24h' for 1-day)"`

	Chan      string `help:"a channel name"`
	innerChan string

	Sort gli.StrList `default:"Name,-Timestamp,ID" help:"sort fields"`

	Format string `default:"{{.ID}}\t{{.Timestamp.Time}}\t{{.Name}}"`
}

func init() {
	gApp.AddExtraCommand(&downloadCmd{}, "download,down", "")
}

func (c *downloadCmd) Before(global globalCmd) error {
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

func (c downloadCmd) Run(global globalCmd, args []string) error {
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

		client := http.Client{}
		req, err := http.NewRequest("GET", f.URLPrivateDownload, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+config.Slack.AccessToken)

		//resp, err := http.Get(f.URLPrivateDownload)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("download %v: %v", f.URLPrivateDownload, err)
		}
		defer resp.Body.Close()

		// store
		if c.Output != nil {
			file, err := os.Create(*c.Output)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, resp.Body)
			if err != nil {
				return err
			}
		} else {
			io.Copy(os.Stdout, resp.Body)
		}

		break
	}

	return nil
}
