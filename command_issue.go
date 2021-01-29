package github

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/seabird-chat/seabird-go/pb"
)

func parseTag(line string) (string, string, bool) {
	line = strings.TrimPrefix(line, "#")
	split := strings.SplitN(line, " ", 2)
	if len(split) != 2 {
		// TODO: useful error here
		return "", "", false
	}

	tag := strings.TrimSpace(split[0])
	line = strings.TrimSpace(split[1])

	return tag, line, true
}

func parseUser(line string) (string, string, bool) {
	line = strings.TrimPrefix(line, "@")
	split := strings.SplitN(line, " ", 2)
	if len(split) != 2 {
		// TODO: useful error here
		return "", "", false
	}

	user := strings.TrimSpace(split[0])
	line = strings.TrimSpace(split[1])

	return user, line, true
}

func (c *Client) issueCallback(source *pb.ChannelSource, cmd *pb.CommandEvent) {
	arg := strings.TrimSpace(cmd.Arg)

	var ok bool

	tag := "default"
	user := ""

	if strings.HasPrefix(arg, "#") {
		if tag, arg, ok = parseTag(arg); !ok {
			_ = c.replyf(source, "Failed to parse repo tag")
			return
		}
	}

	if strings.HasPrefix(arg, "@") {
		if user, arg, ok = parseUser(arg); !ok {
			_ = c.replyf(source, "Failed to parse user")
			return
		}
	}

	target, ok := c.repos[tag]
	if !ok {
		_ = c.replyf(source, "Unknown repo for tag %s", tag)
		return
	}

	req := &github.IssueRequest{
		Title: &arg,
	}
	if user != "" {
		req.Assignee = &user
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	issue, _, err := c.api.Issues.Create(ctx, target.Owner, target.Name, req)
	if err != nil {
		_ = c.replyf(source, "Got error while creating issue: %s", err)
		return
	}

	if issue.Number == nil || issue.URL == nil {
		_ = c.replyf(source, "Got invalid issue response")
		return
	}

	_ = c.replyf(source, "Created issue #%d: %s", *issue.Number, *issue.HTMLURL)
}
