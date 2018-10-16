package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"

	jira "github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func perror(err error) {
	if err == nil {
		return
	}

	println(err.Error())
	os.Exit(1)
}

func perrmsg(msg string) {
	perror(fmt.Errorf("%s", msg))
}

var (
	token           string
	configFile      string
	globalCtx       context.Context
	config          *Config
	githubClient    *github.Client
	jiraClient      *jira.Client
	conflunceClient *jira.Client
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "work-reporter",
		Short: "Work Reporter",
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "C", "", "Config File, default ~/.work-reporter/config.toml")

	rootCmd.AddCommand(
		newDailyCommand(),
		newWeeklyCommand(),
	)

	cobra.OnInitialize(initGlobal)
	cobra.EnablePrefixMatching = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(rootCmd.UsageString())
	}
}

func initGlobal() {
	usr, err := user.Current()
	perror(err)

	if len(configFile) == 0 {
		configFile = path.Join(usr.HomeDir, ".work-reporter/config.toml")
	}
	cfg, err := NewConfigFromFile(configFile)
	perror(err)

	globalCtx = context.Background()
	config = cfg

	initRepoQuery()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.Token},
	)

	tc := oauth2.NewClient(globalCtx, ts)
	githubClient = github.NewClient(tc)

	initTeamMembers()

	jiraTransport := jira.BasicAuthTransport{
		Username: config.Jira.User,
		Password: config.Jira.Password,
	}

	jiraClient, err = jira.NewClient(jiraTransport.Client(), config.Jira.Endpoint)
	perror(err)

	// In our company, we use same user and password for Jira and Confluence.
	if len(config.Confluence.User) == 0 {
		config.Confluence.User = config.Jira.User
	}

	if len(config.Confluence.Password) == 0 {
		config.Confluence.Password = config.Jira.Password
	}

	// A little tricky here, both JIRA and Confluence use the same REST style.
	confluenceTransport := jira.BasicAuthTransport{
		Username: config.Confluence.User,
		Password: config.Confluence.Password,
	}
	conflunceClient, err = jira.NewClient(confluenceTransport.Client(), config.Confluence.Endpoint)
	perror(err)
}
