package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v70/github"
	"github.com/joho/godotenv"
)

var githubClient *github.Client

func initClient() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: GITHUB_TOKEN is required")
		os.Exit(1)
	}
	githubClient = github.NewClient(nil).WithAuthToken(token)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: github-cli <command> [flags]

Commands:
  list-repos       List GitHub repositories for a user or organization
  get-repo         Get GitHub repository details
  list-prs         List pull requests
  get-pr           Get pull request details
  create-pr-comment Create a comment on a pull request
  get-file         Get file content from a GitHub repository
  create-pr        Create a new pull request
  list-issues      List GitHub issues for a repository
  get-issue        Get GitHub issue details
  comment-issue    Comment on a GitHub issue
  issue-action     Close or reopen a GitHub issue
  approve-pr       Approve a pull request

Global flags (available on every command):
  --env string     Path to .env file to load
  --output string  Output format: text or json (default "text")

Run 'github-cli <command> --help' for command-specific flags.
`)
}

func outputText(text string) {
	fmt.Print(text)
}

func outputJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

// parseGlobalFlags strips --env and --output from args before the subcommand
// flag set parses. Returns remaining args and the two global values.
func parseGlobalFlags(args []string) (envFile, output string, remaining []string) {
	output = "text"
	remaining = make([]string, 0, len(args))
	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--env" || arg == "-env":
			if i+1 < len(args) {
				envFile = args[i+1]
				i += 2
			} else {
				fmt.Fprintln(os.Stderr, "Error: --env requires a value")
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--env="):
			envFile = strings.TrimPrefix(arg, "--env=")
			i++
		case strings.HasPrefix(arg, "-env="):
			envFile = strings.TrimPrefix(arg, "-env=")
			i++
		case arg == "--output" || arg == "-output":
			if i+1 < len(args) {
				output = args[i+1]
				i += 2
			} else {
				fmt.Fprintln(os.Stderr, "Error: --output requires a value")
				os.Exit(1)
			}
		case strings.HasPrefix(arg, "--output="):
			output = strings.TrimPrefix(arg, "--output=")
			i++
		case strings.HasPrefix(arg, "-output="):
			output = strings.TrimPrefix(arg, "-output=")
			i++
		default:
			remaining = append(remaining, arg)
			i++
		}
	}
	return
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		printUsage()
		os.Exit(0)
	}

	envFile, output, remaining := parseGlobalFlags(os.Args[2:])

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading env file %s: %v\n", envFile, err)
			os.Exit(1)
		}
	}

	initClient()

	switch cmd {
	case "list-repos":
		runListRepos(remaining, output)
	case "get-repo":
		runGetRepo(remaining, output)
	case "list-prs":
		runListPRs(remaining, output)
	case "get-pr":
		runGetPR(remaining, output)
	case "create-pr-comment":
		runCreatePRComment(remaining, output)
	case "get-file":
		runGetFile(remaining, output)
	case "create-pr":
		runCreatePR(remaining, output)
	case "list-issues":
		runListIssues(remaining, output)
	case "get-issue":
		runGetIssue(remaining, output)
	case "comment-issue":
		runCommentIssue(remaining, output)
	case "issue-action":
		runIssueAction(remaining, output)
	case "approve-pr":
		runApprovePR(remaining, output)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

// list-repos --owner <owner> [--type all]
func runListRepos(args []string, output string) {
	fs := flag.NewFlagSet("list-repos", flag.ExitOnError)
	owner := fs.String("owner", "", "GitHub username or organization name (required)")
	repoType := fs.String("type", "all", "Type of repositories to list (all/owner/public/private/member)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}

	opt := &github.RepositoryListOptions{
		Type: *repoType,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	repos, _, err := githubClient.Repositories.List(context.Background(), *owner, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list repositories: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(repos)
		return
	}

	var sb strings.Builder
	for _, repo := range repos {
		sb.WriteString(fmt.Sprintf("Name: %s\n", repo.GetFullName()))
		sb.WriteString(fmt.Sprintf("Description: %s\n", repo.GetDescription()))
		sb.WriteString(fmt.Sprintf("URL: %s\n", repo.GetHTMLURL()))
		sb.WriteString(fmt.Sprintf("Language: %s\n", repo.GetLanguage()))
		sb.WriteString(fmt.Sprintf("Stars: %d\n", repo.GetStargazersCount()))
		sb.WriteString(fmt.Sprintf("Forks: %d\n", repo.GetForksCount()))
		sb.WriteString(fmt.Sprintf("Created: %s\n", repo.GetCreatedAt().Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("Last Updated: %s\n\n", repo.GetUpdatedAt().Format("2006-01-02 15:04:05")))
	}
	outputText(sb.String())
}

// get-repo --owner <owner> --repo <repo>
func runGetRepo(args []string, output string) {
	fs := flag.NewFlagSet("get-repo", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}

	repository, _, err := githubClient.Repositories.Get(context.Background(), *owner, *repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get repository: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(repository)
		return
	}

	var sb strings.Builder
	sb.WriteString("Repository Details:\n")
	sb.WriteString(fmt.Sprintf("Full Name: %s\n", repository.GetFullName()))
	sb.WriteString(fmt.Sprintf("Description: %s\n", repository.GetDescription()))
	sb.WriteString(fmt.Sprintf("URL: %s\n", repository.GetHTMLURL()))
	sb.WriteString(fmt.Sprintf("Clone URL: %s\n", repository.GetCloneURL()))
	sb.WriteString(fmt.Sprintf("Default Branch: %s\n", repository.GetDefaultBranch()))
	sb.WriteString(fmt.Sprintf("Language: %s\n", repository.GetLanguage()))
	sb.WriteString(fmt.Sprintf("Stars: %d\n", repository.GetStargazersCount()))
	sb.WriteString(fmt.Sprintf("Forks: %d\n", repository.GetForksCount()))
	sb.WriteString(fmt.Sprintf("Open Issues: %d\n", repository.GetOpenIssuesCount()))
	sb.WriteString(fmt.Sprintf("Created: %s\n", repository.GetCreatedAt().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Last Updated: %s\n", repository.GetUpdatedAt().Format("2006-01-02 15:04:05")))
	outputText(sb.String())
}

// list-prs --owner <owner> --repo <repo> [--state open]
func runListPRs(args []string, output string) {
	fs := flag.NewFlagSet("list-prs", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	state := fs.String("state", "open", "PR state (open/closed/all)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}

	opt := &github.PullRequestListOptions{
		State: *state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	prs, _, err := githubClient.PullRequests.List(context.Background(), *owner, *repo, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to list pull requests: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(prs)
		return
	}

	var sb strings.Builder
	for _, pr := range prs {
		sb.WriteString(fmt.Sprintf("PR #%d: %s\n", pr.GetNumber(), pr.GetTitle()))
		sb.WriteString(fmt.Sprintf("State: %s\n", pr.GetState()))
		sb.WriteString(fmt.Sprintf("Author: %s\n", pr.GetUser().GetLogin()))
		sb.WriteString(fmt.Sprintf("URL: %s\n", pr.GetHTMLURL()))
		sb.WriteString(fmt.Sprintf("Created: %s\n", pr.GetCreatedAt().Format("2006-01-02 15:04:05")))
		if !pr.GetMergedAt().IsZero() {
			sb.WriteString(fmt.Sprintf("Merged: %s\n", pr.GetMergedAt().Format("2006-01-02 15:04:05")))
		}
		if !pr.GetClosedAt().IsZero() {
			sb.WriteString(fmt.Sprintf("Closed: %s\n", pr.GetClosedAt().Format("2006-01-02 15:04:05")))
		}
		sb.WriteString(fmt.Sprintf("Base: %s\n", pr.GetBase().GetRef()))
		sb.WriteString(fmt.Sprintf("Head: %s\n", pr.GetHead().GetRef()))
		sb.WriteString("\n")
	}
	outputText(sb.String())
}

// get-pr --owner <owner> --repo <repo> --number <n>
func runGetPR(args []string, output string) {
	fs := flag.NewFlagSet("get-pr", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Pull request number (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}

	prNumber := 0
	fmt.Sscanf(*number, "%d", &prNumber)

	pr, _, err := githubClient.PullRequests.Get(context.Background(), *owner, *repo, prNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get pull request: %v\n", err)
		os.Exit(1)
	}

	comments, _, err := githubClient.Issues.ListComments(context.Background(), *owner, *repo, prNumber, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get pull request comments: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(map[string]interface{}{"pr": pr, "comments": comments})
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("PR #%d: %s\n", pr.GetNumber(), pr.GetTitle()))
	sb.WriteString(fmt.Sprintf("State: %s\n", pr.GetState()))
	sb.WriteString(fmt.Sprintf("Author: %s\n", pr.GetUser().GetLogin()))
	sb.WriteString(fmt.Sprintf("URL: %s\n", pr.GetHTMLURL()))
	sb.WriteString(fmt.Sprintf("Created: %s\n", pr.GetCreatedAt().Format("2006-01-02 15:04:05")))
	if !pr.GetMergedAt().IsZero() {
		sb.WriteString(fmt.Sprintf("Merged: %s\n", pr.GetMergedAt().Format("2006-01-02 15:04:05")))
	}
	if !pr.GetClosedAt().IsZero() {
		sb.WriteString(fmt.Sprintf("Closed: %s\n", pr.GetClosedAt().Format("2006-01-02 15:04:05")))
	}
	sb.WriteString(fmt.Sprintf("Base: %s\n", pr.GetBase().GetRef()))
	sb.WriteString(fmt.Sprintf("Head: %s\n", pr.GetHead().GetRef()))
	sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", pr.GetBody()))
	if len(comments) > 0 {
		sb.WriteString("\nComments:\n")
		for _, comment := range comments {
			sb.WriteString(fmt.Sprintf("\nFrom @%s at %s:\n%s\n",
				comment.GetUser().GetLogin(),
				comment.GetCreatedAt().Format("2006-01-02 15:04:05"),
				comment.GetBody()))
		}
	}
	outputText(sb.String())
}

// create-pr-comment --owner <owner> --repo <repo> --number <n> --comment <text>
func runCreatePRComment(args []string, output string) {
	fs := flag.NewFlagSet("create-pr-comment", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Pull request number (required)")
	comment := fs.String("comment", "", "Comment text (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}
	if *comment == "" {
		fmt.Fprintln(os.Stderr, "Error: --comment is required")
		os.Exit(1)
	}

	prNumber := 0
	fmt.Sscanf(*number, "%d", &prNumber)

	issueComment := &github.IssueComment{
		Body: github.String(*comment),
	}

	result, _, err := githubClient.Issues.CreateComment(context.Background(), *owner, *repo, prNumber, issueComment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create comment: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(result)
		return
	}
	outputText("Comment created successfully\n")
}

// get-file --owner <owner> --repo <repo> --path <path> [--ref <ref>]
func runGetFile(args []string, output string) {
	fs := flag.NewFlagSet("get-file", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	path := fs.String("path", "", "Path to the file in the repository (required)")
	ref := fs.String("ref", "", "Branch name, tag, or commit SHA")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: --path is required")
		os.Exit(1)
	}

	opts := &github.RepositoryContentGetOptions{}
	if *ref != "" {
		opts.Ref = *ref
	}

	content, _, _, err := githubClient.Repositories.GetContents(context.Background(), *owner, *repo, *path, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get file content: %v\n", err)
		os.Exit(1)
	}

	decodedContent, err := content.GetContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to decode content: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(map[string]interface{}{"content": decodedContent, "meta": content})
		return
	}
	outputText(decodedContent)
}

// create-pr --owner <owner> --repo <repo> --title <title> --head <head> --base <base> [--body <body>]
func runCreatePR(args []string, output string) {
	fs := flag.NewFlagSet("create-pr", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	title := fs.String("title", "", "Pull request title (required)")
	head := fs.String("head", "", "Name of the branch where your changes are implemented (required)")
	base := fs.String("base", "", "Name of the branch you want your changes pulled into (required)")
	body := fs.String("body", "", "Pull request description")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *title == "" {
		fmt.Fprintln(os.Stderr, "Error: --title is required")
		os.Exit(1)
	}
	if *head == "" {
		fmt.Fprintln(os.Stderr, "Error: --head is required")
		os.Exit(1)
	}
	if *base == "" {
		fmt.Fprintln(os.Stderr, "Error: --base is required")
		os.Exit(1)
	}

	newPR := &github.NewPullRequest{
		Title: github.String(*title),
		Head:  github.String(*head),
		Base:  github.String(*base),
		Body:  github.String(*body),
	}

	pr, _, err := githubClient.PullRequests.Create(context.Background(), *owner, *repo, newPR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create pull request: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(pr)
		return
	}
	outputText(fmt.Sprintf("Pull request created successfully: %s\n", pr.GetHTMLURL()))
}

// list-issues --owner <owner> --repo <repo> [--state open]
func runListIssues(args []string, output string) {
	fs := flag.NewFlagSet("list-issues", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	state := fs.String("state", "open", "Issue state (open/closed/all)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}

	opt := &github.IssueListByRepoOptions{
		State: *state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	issues, resp, err := githubClient.Issues.ListByRepo(context.Background(), *owner, *repo, opt)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			fmt.Fprintf(os.Stdout, "No issues found for repository %s/%s\n", *owner, *repo)
			return
		}
		fmt.Fprintf(os.Stderr, "Error: failed to list issues: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(issues)
		return
	}

	if len(issues) == 0 {
		outputText(fmt.Sprintf("No %s issues found for repository %s/%s\n", *state, *owner, *repo))
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues for %s/%s:\n\n", len(issues), *owner, *repo))
	for _, issue := range issues {
		if issue == nil {
			continue
		}
		if issue.IsPullRequest() {
			continue
		}
		sb.WriteString(fmt.Sprintf("Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
		sb.WriteString(fmt.Sprintf("State: %s\n", issue.GetState()))
		if user := issue.GetUser(); user != nil {
			sb.WriteString(fmt.Sprintf("Author: %s\n", user.GetLogin()))
		}
		sb.WriteString(fmt.Sprintf("URL: %s\n", issue.GetHTMLURL()))
		if !issue.GetCreatedAt().IsZero() {
			sb.WriteString(fmt.Sprintf("Created: %s\n", issue.GetCreatedAt().Format("2006-01-02 15:04:05")))
		}
		if !issue.GetClosedAt().IsZero() {
			sb.WriteString(fmt.Sprintf("Closed: %s\n", issue.GetClosedAt().Format("2006-01-02 15:04:05")))
		}
		if len(issue.Labels) > 0 {
			labels := make([]string, 0, len(issue.Labels))
			for _, label := range issue.Labels {
				if label != nil {
					labels = append(labels, label.GetName())
				}
			}
			if len(labels) > 0 {
				sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(labels, ", ")))
			}
		}
		sb.WriteString("\n")
	}
	outputText(sb.String())
}

// get-issue --owner <owner> --repo <repo> --number <n>
func runGetIssue(args []string, output string) {
	fs := flag.NewFlagSet("get-issue", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Issue number (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}

	issueNumber := 0
	fmt.Sscanf(*number, "%d", &issueNumber)

	issue, _, err := githubClient.Issues.Get(context.Background(), *owner, *repo, issueNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get issue: %v\n", err)
		os.Exit(1)
	}

	comments, _, err := githubClient.Issues.ListComments(context.Background(), *owner, *repo, issueNumber, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get issue comments: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(map[string]interface{}{"issue": issue, "comments": comments})
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
	sb.WriteString(fmt.Sprintf("State: %s\n", issue.GetState()))
	sb.WriteString(fmt.Sprintf("Author: %s\n", issue.GetUser().GetLogin()))
	sb.WriteString(fmt.Sprintf("URL: %s\n", issue.GetHTMLURL()))
	sb.WriteString(fmt.Sprintf("Created: %s\n", issue.GetCreatedAt().Format("2006-01-02 15:04:05")))
	if !issue.GetClosedAt().IsZero() {
		sb.WriteString(fmt.Sprintf("Closed: %s\n", issue.GetClosedAt().Format("2006-01-02 15:04:05")))
	}
	if len(issue.Labels) > 0 {
		labels := make([]string, 0, len(issue.Labels))
		for _, label := range issue.Labels {
			labels = append(labels, label.GetName())
		}
		sb.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(labels, ", ")))
	}
	sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", issue.GetBody()))
	if len(comments) > 0 {
		sb.WriteString("\nComments:\n")
		for _, comment := range comments {
			sb.WriteString(fmt.Sprintf("\nFrom @%s at %s:\n%s\n",
				comment.GetUser().GetLogin(),
				comment.GetCreatedAt().Format("2006-01-02 15:04:05"),
				comment.GetBody()))
		}
	}
	outputText(sb.String())
}

// comment-issue --owner <owner> --repo <repo> --number <n> --comment <text>
func runCommentIssue(args []string, output string) {
	fs := flag.NewFlagSet("comment-issue", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Issue number (required)")
	comment := fs.String("comment", "", "Comment text (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}
	if *comment == "" {
		fmt.Fprintln(os.Stderr, "Error: --comment is required")
		os.Exit(1)
	}

	issueNumber := 0
	fmt.Sscanf(*number, "%d", &issueNumber)

	issueComment := &github.IssueComment{
		Body: github.String(*comment),
	}

	result, _, err := githubClient.Issues.CreateComment(context.Background(), *owner, *repo, issueNumber, issueComment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create comment: %v\n", err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(result)
		return
	}
	outputText("Comment created successfully\n")
}

// issue-action --owner <owner> --repo <repo> --number <n> --action <close|reopen>
func runIssueAction(args []string, output string) {
	fs := flag.NewFlagSet("issue-action", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Issue number (required)")
	action := fs.String("action", "", "Action to take: close or reopen (required)")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}
	if *action == "" {
		fmt.Fprintln(os.Stderr, "Error: --action is required")
		os.Exit(1)
	}

	issueNumber := 0
	fmt.Sscanf(*number, "%d", &issueNumber)

	var state string
	switch *action {
	case "close":
		state = "closed"
	case "reopen":
		state = "open"
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid action: %s. Must be either 'close' or 'reopen'\n", *action)
		os.Exit(1)
	}

	issueReq := &github.IssueRequest{
		State: &state,
	}

	result, _, err := githubClient.Issues.Edit(context.Background(), *owner, *repo, issueNumber, issueReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to %s issue: %v\n", *action, err)
		os.Exit(1)
	}

	if output == "json" {
		outputJSON(result)
		return
	}
	outputText(fmt.Sprintf("Issue %sd successfully\n", *action))
}

// approve-pr --owner <owner> --repo <repo> --number <n> [--merge-method <merge|squash|rebase>]
func runApprovePR(args []string, output string) {
	fs := flag.NewFlagSet("approve-pr", flag.ExitOnError)
	owner := fs.String("owner", "", "Repository owner (required)")
	repo := fs.String("repo", "", "Repository name (required)")
	number := fs.String("number", "", "Pull request number (required)")
	mergeMethod := fs.String("merge-method", "", "Merge method (merge/squash/rebase). Leave empty to not merge.")
	fs.Parse(args)

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "Error: --owner is required")
		os.Exit(1)
	}
	if *repo == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo is required")
		os.Exit(1)
	}
	if *number == "" {
		fmt.Fprintln(os.Stderr, "Error: --number is required")
		os.Exit(1)
	}

	prNumber := 0
	fmt.Sscanf(*number, "%d", &prNumber)

	review := &github.PullRequestReviewRequest{
		Event: github.String("APPROVE"),
	}
	_, _, err := githubClient.PullRequests.CreateReview(context.Background(), *owner, *repo, prNumber, review)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to approve pull request: %v\n", err)
		os.Exit(1)
	}

	if *mergeMethod != "" {
		_, _, err = githubClient.PullRequests.Merge(context.Background(), *owner, *repo, prNumber, "Auto-merge after approval", &github.PullRequestOptions{
			MergeMethod: *mergeMethod,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to merge pull request: %v\n", err)
			os.Exit(1)
		}
	}

	if output == "json" {
		outputJSON(map[string]interface{}{"status": "approved"})
		return
	}
	outputText("Pull request approved and merge initiated successfully\n")
}
