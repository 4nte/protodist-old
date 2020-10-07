package provider

import (
	"context"
	"net/http"
	"time"
)
import "github.com/google/go-github/v32/github"
import "golang.org/x/oauth2"

type GitProvider struct {
	Name string
	Host string
}

var GithubProvider = GitProvider{
	Name: "github",
	Host: "github.com",
}

func NewHttpClient(accessToken string) *http.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	return oauth2.NewClient(ctx, ts)
}

func NewGithubClient(accessToken string) *github.Client {
	httpClient := NewHttpClient(accessToken)
	return github.NewClient(httpClient)
}
