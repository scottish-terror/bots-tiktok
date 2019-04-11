package tiktokmod

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// gitmyhub - creates connection with github to prepare API request
func gitmyhub(tiktok *TikTokConf) (ctx context.Context, client *github.Client) {
	gitKey := tiktok.Config.GitToken
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitKey},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)

	return ctx, client
}

// RetrieveOrgRepo - retrieve list of repositories for an Organization
func RetrieveOrgRepo(tiktok *TikTokConf, orgName string) (repos []*github.Repository) {

	ctx, client := gitmyhub(tiktok)

	opt := &github.RepositoryListByOrgOptions{
		Type: "all",
	}

	repos, _, err := client.Repositories.ListByOrg(ctx, orgName, opt)
	if err != nil {
		errTrap(tiktok, "Error in `retrieveRepo` function in `git.go`", err)
		return
	}

	return repos

}

// RetrieveRepo - retrieve list of repositories for the API user
func RetrieveRepo(tiktok *TikTokConf) (repos []*github.Repository) {

	ctx, client := gitmyhub(tiktok)

	opt := &github.RepositoryListOptions{
		Sort:        "updated",
		Type:        "all",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repos, _, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		errTrap(tiktok, "Error in `retrieveRepo` function in `git.go`", err)
		return
	}

	return repos

}

// RetrieveUsers - retrieve list of Users
func RetrieveUsers(tiktok *TikTokConf, org string) (users []*github.User) {

	ctx, client := gitmyhub(tiktok)

	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	users, _, err := client.Organizations.ListMembers(ctx, org, opt)
	if err != nil {
		errTrap(tiktok, "Error in `RetrieveUsers` function in `git.go`", err)
		return
	}

	return users

}

// GitPRList - retrieve all PRs in a repo
func GitPRList(tiktok *TikTokConf, repoName string, org string) (pulls []*github.PullRequest, err error) {

	ctx, client := gitmyhub(tiktok)

	opt := &github.PullRequestListOptions{
		Sort:      "updated",
		Direction: "desc",
	}

	pulls, _, err = client.PullRequests.List(ctx, org, repoName, opt)
	if err != nil {
		errTrap(tiktok, "Error in `GitPRList` function in `git.go`", err)
		return pulls, err
	}

	return pulls, nil
}

// GitPR - retrieve a single PR in a repo
func GitPR(tiktok *TikTokConf, repoName string, PRID int, org string) (pull *github.PullRequest, err error) {

	ctx, client := gitmyhub(tiktok)

	pull, _, err = client.PullRequests.Get(ctx, org, repoName, PRID)
	if err != nil {
		errTrap(tiktok, "Error in `GitPR` function in `git.go`", err)
		return pull, err
	}

	return pull, nil
}
