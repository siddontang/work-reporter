# Work Reporter

We have been heavily using Github, JIRA and Confluence in our work, but we meet some problems now, for example:

1. We use Sprint to manage our work, the Sprint starts at 0:00 on Friday and ends at 24:00 on Thursday. So every week, we need to close the current active Sprint and create a new Sprint manually. 
2. We hold a weekly meeting on Friday, everyone needs to write what he/she does in this Sprint, he/she needs to search through JIRA, Github, then adds the associated links to the weekly report page manually.
3. Each Sprint there are two duty rosters, and the rosters need to check new issues, PRs and assign them to the specified people every day. Sometimes, they may forget to do it.

As you can see, here we do many work manually, we need a powerful tool to help us, to do things automatically. 

The tool needs to support:

## Weekly

+ Grabs new OnCall issues from the OnCall board, adds to weekly report
+ Grabs new Github issues, adds to weekly report
+ For each team member, grabs his/her current Sprint / next Sprint work from JIRA, reviewed pull requests from Github, adds to weekly report
+ Closes the current Sprint, creates a new next Sprint, sends messages to slack channel

## Daily

+ Grabs new issues, pull requests during last 24 hours, adds to weekly duty report
+ sends messages to slack channel