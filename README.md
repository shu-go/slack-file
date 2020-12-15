manipulate files in Slack

[![Go Report Card](https://goreportcard.com/badge/github.com/shu-go/slack-file)](https://goreportcard.com/report/github.com/shu-go/slack-file)
![MIT License](https://img.shields.io/badge/License-MIT-blue)

* List
* Delete
* Uniq (delete duplicated files)
* Upload

# Usage

```
slack-file

Sub commands:
  auth                     authenticate
  delete, remove, del, rm  delete files
  list, ls                 list files
  uniq                     delete duplicated files

Options:
  --config   (default: ./slack-file.conf)

Usage:
  -------------------
  how to start (uniq)
  -------------------

  1. follow 'slack-file help auth'
  2. 'slack-file auth {CLIENT_ID} {CLIENT_SECRET}'
  3. 'slack-file uniq --dry-run'
  4. 'slack-file help uniq'
  5. 'slack-file uniq'

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

Help sub commands:
  help     slack-file help subcommnad subsubcommand
  version  show version

(C) 2020 Shuhei Kubota
```

## Auth (first step)

```
command auth - authenticate

Options:
  --port PORT        a temporal PORT for OAuth authentication. (default: 7878)
  --timeout TIMEOUT  set TIMEOUT (in seconds) on authentication transaction. < 0 is infinite. (default: 60)

Global Options:
  --config   (default: ./slack-file.conf)

Usage:
  1. go to https://api.slack.com/apps
  2. make a new app (files:read, files:write)
  3. slack-file slack auth CLIENT_ID CLIENT_SECRET
```

## List

```
command list - list files

Options:
  --target                (default: Name,Title,ID)
  --older, --older-than  Timestamp (e.g. '24h' for 1-day)
  --chan                 a channel name
  --sort                 sort fields (default: Name,-Timestamp,ID)
  --group                e.g. Channels,Groups,IMs
  --format                (default: {{.ID}}     {{.Timestamp.Time}}     {{.Name}})

Global Options:
  --config   (default: ./slack-file.conf)

Usage:
  # list all
  slack-file list
  # find by pattern
  slack-file list my*.txt
  # files older than 1day
  slack-file list --older 24h
  # files in a general channel
  slack-file list --chan general
```


## Delete

```
command delete - delete files

Options:
  --target                (default: Name,Title,ID)
  --older, --older-than  Timestamp (e.g. '24h' for 1-day)
  --chan                 a channel name
  --dry-run              do not delete files actually
  --format                (default: {{.ID}}     {{.Timestamp.Time}}     {{.Name}})

Global Options:
  --config   (default: ./slack-file.conf)

Usage:
  # delete by pattern
  slack-file delete my*.txt
  # files older than 1day
  slack-file delete --older 24h *
  # files in a general channel
  slack-file delete --chan general *
```

## Uniq

```
command uniq - delete duplicated files

Options:
  --key      a unique key set of files (default: Name,Title)
  --sort     sort fields of each --key group (default: -Created,-Timestamp,ID)
  --exclude  do not delete if any properties not empty (default: IsStarred,IsExternal)
  --dry-run  do not delete files actually

Global Options:
  --config   (default: ./slack-file.conf)

Usage:
  # SIMULATE delete duplicated files by Name, keep newest Timestamp
  slack-file uniq --key Name --sort -Timestamp --dry-run
  # DELETE
  slack-file uniq --key Name --sort -Timestamp
```

## Upload

```
command upload - upload a file

Options:
  --title  the title of the file
  --chan   channel or group name (sub-match, posting to all matching channels and groups, no #) (default: general)

Global Options:
  --config   (default: ./slack-file.conf)

Usage:
  slack-file upload --chan mychannel /path/to/myfile.log
```
