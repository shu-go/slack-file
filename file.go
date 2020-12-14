package main

import (
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

func listFiles(client *slack.Client, params slack.ListFilesParameters) ([]slack.File, error) {
	var files []slack.File

LOOP:
	for {
		list, rparams, err := client.ListFiles(params)
		if err != nil {
			return nil, err
		}

		params = *rparams

		if len(list) == 0 {
			break LOOP
		}
		if rparams.Cursor == "" {
			break LOOP
		}

		files = append(files, list...)
	}

	return files, nil
}

func testProp(f slack.File, prop string) bool {
	p := fileProp(f, prop)
	return p != "" && p != "0"
}

func fileProp(f slack.File, prop string) string {
	p := strings.ToLower(prop)
	if strings.HasPrefix(p, "-") {
		p = p[1:]
	}

	switch p {
	case "id":
		return f.ID
	case "created":
		return strconv.FormatInt(int64(f.Created), 10)
	case "timestamp":
		return strconv.FormatInt(int64(f.Timestamp), 10)
	case "name":
		return f.Name
	case "title":
		return f.Title
	case "mimetype":
		return f.Mimetype
	case "imageexifrotation":
		return strconv.Itoa(f.ImageExifRotation)
	case "filetype":
		return f.Filetype
	case "prettytype":
		return f.PrettyType
	case "user":
		return f.User
	case "mode":
		return f.Mode
	case "editable":
		return bool2NumStr(f.Editable)
	case "isexternal":
		return bool2NumStr(f.IsExternal)
	case "externaltype":
		return f.ExternalType
	case "size":
		return strconv.Itoa(f.Size)
	case "urlprivate":
		return f.URLPrivate
	case "urlprivatedownload":
		return f.URLPrivateDownload
	case "originalh":
		return strconv.Itoa(f.OriginalH)
	case "originalw":
		return strconv.Itoa(f.OriginalW)
	case "permalink":
		return f.Permalink
	case "permalinkpublic":
		return f.PermalinkPublic
	case "editlink":
		return f.EditLink
	case "preview":
		return f.Preview
	case "previewhighlight":
		return f.PreviewHighlight
	case "lines":
		return strconv.Itoa(f.Lines)
	case "linesmore":
		return strconv.Itoa(f.LinesMore)
	case "ispublic":
		return bool2NumStr(f.IsPublic)
	case "publicurlshared":
		return bool2NumStr(f.PublicURLShared)
	case "channels":
		return strings.Join(f.Channels, ",")
	case "groups":
		return strings.Join(f.Groups, ",")
	case "ims":
		return strings.Join(f.IMs, ",")
	case "initialcomment":
		return f.InitialComment.Comment
	case "commentscount":
		return strconv.Itoa(f.CommentsCount)
	case "numstars":
		return strconv.Itoa(f.NumStars)
	case "isstarred":
		return bool2NumStr(f.IsStarred)
	default:
		return ""
	}
	return ""
}

func bool2NumStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func filePropsCompare(f1, f2 slack.File, props []string) int {
	for _, p := range props {
		c := strings.Compare(fileProp(f1, p), fileProp(f2, p))
		if c == 0 {
			continue
		}

		if strings.HasPrefix(p, "-") {
			return -c
		}
		return c
	}
	return 0
}
