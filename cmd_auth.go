package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/browser"
	"github.com/shu-go/minredir"
)

var (
	slackOAuth2ClientID     string = ""
	slackOAuth2ClientSecret string = ""
)

func init() {
	gApp.AddExtraCommand(&authCmd{}, "auth", "")
}

type authCmd struct {
	_ struct{} `help:"authenticate"   usage:"1. go to https://api.slack.com/apps\n2. make a new app\n\tRedirect URLs: https://localhost:7878\n\tScopes: files:read, files:write\n3. slack-file-uniq slack auth CLIENT_ID CLIENT_SECRET"`

	Port    int `cli:"port=PORT" default:"7878" help:"a temporal PORT for OAuth authentication."`
	Timeout int `cli:"timeout=TIMEOUT" default:"60" help:"set TIMEOUT (in seconds) on authentication transaction. < 0 is infinite."`
}

func (c authCmd) Run(global globalCmd, args []string) error {
	config, _ := loadConfig(global.Config)

	var argClientID, argCLientSecret string
	if len(args) >= 2 {
		argClientID = args[0]
		argCLientSecret = args[1]
	}

	//
	// prepare
	//
	slackOAuth2ClientID = firstNonEmpty(
		argClientID,
		config.Slack.ClientID,
		os.Getenv("SLACK_OAUTH2_CLIENT_ID"),
		slackOAuth2ClientID)
	slackOAuth2ClientSecret = firstNonEmpty(
		argCLientSecret,
		config.Slack.ClientSecret,
		os.Getenv("SLACK_OAUTH2_CLIENT_SECRET"),
		slackOAuth2ClientSecret)

	if slackOAuth2ClientID == "" || slackOAuth2ClientSecret == "" {
		fmt.Fprintf(os.Stderr, "both SLACK_OAUTH2_CLIENT_ID and SLACK_OAUTH2_CLIENT_SECRET must be given.\n")
		fmt.Fprintf(os.Stderr, "access to https://api.slack.com/apps\n")
		browser.OpenURL("https://api.slack.com/apps")
		return nil
	}

	redirectURI := fmt.Sprintf("https://localhost:%d/", c.Port)

	//
	// fetch the authentication code
	//
	authURI := slackAuthURI(slackOAuth2ClientID, redirectURI)
	if err := browser.OpenURL(authURI); err != nil {
		return fmt.Errorf("failed to open the authURI(%s): %v", authURI, err)
	}

	resultChan := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	err, errChan := minredir.ServeTLS(ctx, fmt.Sprintf(":%v", c.Port), resultChan)

	authCode := waitForStringChan(resultChan, time.Duration(c.Timeout)*time.Second)
	cancel()

	if authCode == "" {
		select {
		case err = <-errChan:
		default:
			err = nil
		}
		return fmt.Errorf("failed or timed out fetching an authentication code: %w", err)
	}

	//
	// fetch the access token
	//
	accessToken, err := slackFetchAccessToken(slackOAuth2ClientID, slackOAuth2ClientSecret, authCode, redirectURI)
	if err != nil {
		return fmt.Errorf("failed or timed out fetching the refresh token: %v", err)
	}

	//
	// store the token to the config file.
	//
	config.Slack.AccessToken = accessToken
	saveConfig(config, global.Config)

	return nil
}

func slackAuthURI(clientID, redirectURI string, optTeamAndState ...string) string {
	const (
		oauth2Scope       = "chat:write:bot channels:read"
		oauth2AuthBaseURL = "https://slack.com/oauth/authorize"
	)

	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("scope", oauth2Scope)
	form.Add("redirect_uri", redirectURI)
	if len(optTeamAndState) >= 1 {
		form.Add("team", optTeamAndState[0])
	}
	if len(optTeamAndState) >= 2 {
		form.Add("state", optTeamAndState[1])
	}
	return fmt.Sprintf("%s?%s", oauth2AuthBaseURL, form.Encode())
}

func slackFetchAccessToken(clientID, clientSecret, authCode, redirectURI string) (string, error) {
	const (
		oauth2TokenBaseURL = "https://slack.com/api/oauth.access"
	)

	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("code", authCode)
	form.Add("redirect_uri", redirectURI)

	resp, err := http.PostForm(oauth2TokenBaseURL, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	t := slackOAuth2AuthedTokens{}
	err = dec.Decode(&t)
	if errors.Is(err, io.EOF) {
		return "", fmt.Errorf("auth response from the server is empty")
	} else if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

type slackOAuth2AuthedTokens struct {
	AccessToken string `json:"access_token"`
}
