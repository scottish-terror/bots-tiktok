package wallemod

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// gitmyhub - creates connection with github to prepare API request
func gitmyhub(wOpts *WallConf) (ctx context.Context, client *github.Client) {
	gitKey := wOpts.Walle.GitToken
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitKey},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)

	return ctx, client
}

// RetrieveRepo - retrieve list of repositories
func RetrieveRepo(wOpts *WallConf) (repos []*github.Repository) {

	ctx, client := gitmyhub(wOpts)

	opt := &github.RepositoryListOptions{
		Sort:        "updated",
		Type:        "all",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repos, _, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		errTrap(wOpts, "Error in `retrieveRepo` function in `git.go`", err)
		return
	}

	return repos

}

// RetrieveUsers - retrieve list of Users
func RetrieveUsers(wOpts *WallConf) (users []*github.User) {

	ctx, client := gitmyhub(wOpts)

	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	users, _, err := client.Organizations.ListMembers(ctx, "ForgeCloud", opt)
	if err != nil {
		errTrap(wOpts, "Error in `RetrieveUsers` function in `git.go`", err)
		return
	}

	return users

}

// GitPRList - retrieve all PRs in a repo
func GitPRList(wOpts *WallConf, repoName string) (pulls []*github.PullRequest, err error) {

	ctx, client := gitmyhub(wOpts)

	opt := &github.PullRequestListOptions{
		Sort:      "updated",
		Direction: "desc",
	}

	pulls, _, err = client.PullRequests.List(ctx, "ForgeCloud", repoName, opt)
	if err != nil {
		errTrap(wOpts, "Error in `GitPRList` function in `git.go`", err)
		return pulls, err
	}

	return pulls, nil
}

// GitPR - retrieve a single PR in a repo
func GitPR(wOpts *WallConf, repoName string, PRID int) (pull *github.PullRequest, err error) {

	ctx, client := gitmyhub(wOpts)

	pull, _, err = client.PullRequests.Get(ctx, "ForgeCloud", repoName, PRID)
	if err != nil {
		errTrap(wOpts, "Error in `GitPR` function in `git.go`", err)
		return pull, err
	}

	return pull, nil
}
