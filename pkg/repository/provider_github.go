package repository

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func init() {
	RegisterDriver("github", NewGithubRepository)
}

type GithubRepository struct {
	Config  map[string]string
	Context context.Context
	Client  *github.Client
}

func NewGithubRepository(config map[string]string) (Repository, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config["token"]},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return &GithubRepository{Config: config, Context: ctx, Client: client}, nil
}

func (c *GithubRepository) GetURL(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, name)
}

func (c *GithubRepository) Exists(owner, name string) (bool, error) {
	_, response, err := c.Client.Repositories.Get(c.Context, owner, name)

	if response.Response.StatusCode == 200 {
		return true, nil
	}

	if response.Response.StatusCode == 404 {
		return false, nil
	}

	return false, err
}

func (c *GithubRepository) Create(org, name string, private bool) error {
	repo := github.Repository{
		Name: &name,
		Private: &private,
	}

	_, _, err := c.Client.Repositories.Create(c.Context, org, &repo)
	if err != nil {
		return err
	}

	return nil
}
