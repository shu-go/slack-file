package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/slack-go/slack"
)

type uploadCmd struct {
	_ struct{} `help:"upload a file" usage:"slack-file upload --chan mychannel /path/to/myfile.log"`

	Title string `help:"the title of the file"`

	Chan string `default:"general"  help:"channel or group name (sub-match, posting to all matching channels and groups, no #)"`
}

func init() {
	gApp.AddExtraCommand(&uploadCmd{}, "upload,up", "")
}

func (c uploadCmd) Run(global globalCmd, args []string) error {
	config, _ := loadConfig(global.Config)

	if config.Slack.AccessToken == "" {
		return errors.New("auth first")
	}

	if len(args) != 1 {
		return errors.New("one file is required")
	}

	filename := filepath.Base(args[0])
	if c.Title == "" {
		c.Title = filename
	}

	sl := slack.New(config.Slack.AccessToken)

	upparams := slack.FileUploadParameters{
		File:     args[0],
		Channels: []string{c.Chan},
		Title:    c.Title,
		Filename: filename,
	}
	_, err := sl.UploadFile(upparams)
	if err != nil {
		return fmt.Errorf("failed to upload file %v: %v", filename, err)
	}

	return nil
}
