package main

import (
	"os"
	"time"

	"github.com/shu-go/gli"
)

// Version is app version
var Version string

func init() {
	if Version == "" {
		Version = "dev-" + time.Now().Format("20060102")
	}
}

var gApp gli.App = gli.NewWith(&globalCmd{})

type globalCmd struct {
	Config string `default:"./slack-file.conf"`
}

func main() {
	gApp.Name = "slack-file"
	gApp.Desc = "manipulate files in Slack"
	gApp.Version = Version
	gApp.Usage = `-------------------
how to start (uniq)
-------------------

1. follow 'slack-file help auth'
     'slack-file auth'
     'slack-file auth {CLIENT_ID} {CLIENT_SECRET}'
2. 'slack-file uniq --dry-run'
3. 'slack-file uniq'

---------------------------
 --key, --sort, --exclude
---------------------------

id
created
timestamp
name
title
mimetype
imageexifrotation
filetype
prettytype
user
mode
editable
isexternal
externaltype
size
urlprivate
urlprivatedownload
originalh
originalw
permalink
permalinkpublic
editlink
preview
previewhighlight
lines
linesmore
ispublic
publicurlshared
channels
groups
ims
initialcomment
commentscount
numstars
isstarred
`
	gApp.Copyright = "(C) 2020 Shuhei Kubota"
	gApp.Run(os.Args)
}
