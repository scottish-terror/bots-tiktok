package tiktokmod

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// gitmyhub - creates connection with github to prepare API request
func gitmyhub(baloo *BalooConf) (ctx context.Context, client *github.Client) {
	gitKey := baloo.Config.GitToken
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitKey},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)

	return ctx, client
}

// RetrieveRepo - retrieve list of repositories
func RetrieveRepo(baloo *BalooConf) (repos []*github.Repository) {

	ctx, client := gitmyhub(baloo)

	opt := &github.RepositoryListOptions{
		Sort:        "updated",
		Type:        "all",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repos, _, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		errTrap(baloo, "Error in `retrieveRepo` function in `git.go`", err)
		return
	}

	return repos

}

// RetrieveUsers - retrieve list of Users
func RetrieveUsers(baloo *BalooConf) (users []*github.User) {

	ctx, client := gitmyhub(baloo)

	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	users, _, err := client.Organizations.ListMembers(ctx, "ForgeCloud", opt)
	if err != nil {
		errTrap(baloo, "Error in `RetrieveUsers` function in `git.go`", err)
		return
	}

	return users

}

// GitPRList - retrieve all PRs in a repo
func GitPRList(baloo *BalooConf, repoName string) (pulls []*github.PullRequest, err error) {

	ctx, client := gitmyhub(baloo)

	opt := &github.PullRequestListOptions{
		Sort:      "updated",
		Direction: "desc",
	}

	pulls, _, err = client.PullRequests.List(ctx, "ForgeCloud", repoName, opt)
	if err != nil {
		errTrap(baloo, "Error in `GitPRList` function in `git.go`", err)
		return pulls, err
	}

	return pulls, nil
}

// GitPR - retrieve a single PR in a repo
func GitPR(baloo *BalooConf, repoName string, PRID int) (pull *github.PullRequest, err error) {

	ctx, client := gitmyhub(baloo)

	pull, _, err = client.PullRequests.Get(ctx, "ForgeCloud", repoName, PRID)
	if err != nil {
		errTrap(baloo, "Error in `GitPR` function in `git.go`", err)
		return pull, err
	}

	return pull, nil
}
