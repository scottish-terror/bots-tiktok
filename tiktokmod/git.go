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

// RetrieveRepo - retrieve list of repositories. org is optional to filter out specific organization repos from API user repos
func RetrieveRepo(tiktok *TikTokConf, org string) (repos []*github.Repository) {
	var newRepo []*github.Repository

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
	if org != "" {
		for _, r := range repos {
			if *r.Organization.Name == org {
				newRepo = append(newRepo, r)
			}
		}

		return newRepo
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
