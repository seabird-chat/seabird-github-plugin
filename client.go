package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/seabird-chat/seabird-go"
	"github.com/seabird-chat/seabird-go/pb"
)

// This is really just a holding type for some configuration.
type Repo struct {
	Owner string
	Name  string
}

type Client struct {
	*seabird.Client

	api   *github.Client
	repos map[string]Repo
}

func NewClient(seabirdCoreUrl, seabirdCoreToken, githubToken string, repos map[string]Repo) (*Client, error) {
	client, err := seabird.NewClient(seabirdCoreUrl, seabirdCoreToken)
	if err != nil {
		return nil, err
	}

	// Create an oauth2 client for use with the GitHub API.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.TODO(), ts)

	return &Client{
		Client: client,
		api:    github.NewClient(tc),
		repos:  repos,
	}, nil
}

func (c *Client) replyf(source *pb.ChannelSource, format string, args ...interface{}) error {
	// This is a bit ugly, but it's the simplest way to do this.
	msg := fmt.Sprintf(
		"%s: %s",
		source.GetUser().GetDisplayName(),
		fmt.Sprintf(format, args...))

	_, err := c.Client.Inner.SendMessage(context.TODO(), &pb.SendMessageRequest{
		ChannelId: source.ChannelId,
		Text:      msg,
		Tags: map[string]string{
			"url/skip": "1",
		},
	})
	return err
}

func (c *Client) Run() error {
	events, err := c.StreamEvents(map[string]*pb.CommandMetadata{
		"issue": {
			Name:      "issue",
			ShortHelp: "[#tag ][@asignee ]<issue title>",
			FullHelp:  "Creates a new seabird issue on GitHub",
		},
	})
	if err != nil {
		return err
	}
	defer events.Close()

	for event := range events.C {
		switch v := event.GetInner().(type) {
		case *pb.Event_Command:
			if v.Command.Command == "issue" {
				c.issueCallback(v.Command.Source, v.Command)
			}
		}
	}

	return errors.New("event stream closed")
}
