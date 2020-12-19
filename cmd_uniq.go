package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/shu-go/gli"
	"github.com/slack-go/slack"
)

func init() {
	gApp.AddExtraCommand(&uniqCmd{}, "uniq", "")
}

func fileString(f slack.File) string {
	var title string
	if f.Title != f.Name {
		title = "(" + f.Title + ")"
	}
	return fmt.Sprintf("%v%v - %v", f.Name, title, f.Created.Time())
}

type uniqCmd struct {
	_ struct{} `help:"delete duplicated files" usage:"# SIMULATE delete duplicated files by Name, keep newest Timestamp\nslack-file uniq --key Name --sort -Timestamp --dry-run\n# DELETE\nslack-file uniq --key Name --sort -Timestamp"`

	Key     gli.StrList `default:"Name,Title" help:"a unique key set of files"`
	Sort    gli.StrList `default:"-Created,-Timestamp,ID" help:"sort fields of each --key group"`
	Exclude gli.StrList `default:"IsStarred,IsExternal" help:"do not delete if any properties not empty"`

	DryRun bool `cli:"dry-run" help:"do not delete files actually"`

	Format string `default:"{{.Name}}({{.ID}})\t{{.Timestamp.Time}}"`
}

func (c uniqCmd) Run(global globalCmd) error {
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
	sortProps = append(sortProps, c.Key...)
	sortProps = append(sortProps, c.Sort...)
	sort.Slice(files, func(i, j int) bool {
		c := filePropsCompare(files[i], files[j], sortProps)
		return c < 0
	})

	var head *slack.File
	for _, f := range files {
		filestr, err := fileToString(c.Format, f)
		if err != nil {
			return err
		}

		excluded := false
		for _, e := range c.Exclude {
			if testProp(f, e) {
				excluded = true
			}
		}
		if excluded {
			fmt.Printf("[EXCLUDED] %v\n", filestr)
			continue
		}

		if head == nil || filePropsCompare(*head, f, c.Key) != 0 {
			fmt.Printf("%v\n", filestr)
			curr := f
			head = &curr
		} else {
			fmt.Printf("  [DEL] %v\n", filestr)
			if !c.DryRun && !excluded {
				err := sl.DeleteFile(f.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
