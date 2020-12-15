package main

import "github.com/slack-go/slack"

func listConversationsForUser(client *slack.Client, params slack.GetConversationsForUserParameters) ([]slack.Channel, error) {
	var chans []slack.Channel

LOOP:
	for {
		list, nextCursor, err := client.GetConversationsForUser(&params)
		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			break LOOP
		}

		chans = append(chans, list...)

		if nextCursor == "" {
			break LOOP
		}

		params.Cursor = nextCursor
	}

	return chans, nil
}
