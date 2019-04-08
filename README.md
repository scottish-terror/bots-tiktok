## Tik-Tok Bot - Distinguished Scrum Master 

Tik-Tok v3 is an always on services type bot.  

Tik-Tok Modules
```
go get github.com/robfig/cron
go get github.com/nlopes/slack
go get github.com/google/go-github/github
go get golang.org/x/oauth2
go get github.com/parnurzeal/gorequest
go get github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql
go get github.com/BurntSushi/toml
go get github.com/jinzhu/copier
```

Tik-Tok *required* CLI parameters
```
Usage of Tik-Tok.go:
  -tkey  Trello API Key
  -ttoken  Trello API Token
  -slackhook  Slack API Webhook URL (required)  
  -slacktoken   Slack Bot Token
  -slackoauth  Slack App User OAuth Token (required to manage slack channels)
  -git  Github API Token
  -dbuser  Google Cloud SQL User
  -dbpass  Google Cloud SQL Password
```

Tik-Tok *optional* CLI parameters
```
  -nocron  Do not load built-in cronjobs on start.  crons.toml
```

Tik-Tok has a Tik-Tok.toml file that has some base configuration parameters you need to set for him to run

Each team/board that needs Tik-Tok to operate on it must be configured a specific way and then contain a .TOML file with the settings and configuration for that board.   Below is how to configure your specific board .TOML file.   Tik-Tok does not need to be restarted if a new .TOML file is built and added to his working directory.  He will find it on command.  The name of the TOML file (minus the extension) is how you will refer to that when talking to Tik-Tok.  For example a file called  SAAS.toml will be referred to when talking to Tik-Tok as [SAAS].  

###  Configuring your Trello Board
* Board must be created ahead of time you will need the boardID.  BoardID is in the URL when viewing the board. `https://trello.com/b/<BOARDID>/`
* Board must initially have colum/list names Matching the list below. *NOTE*: Columns/Lists can be renamed once the ID's have been acquired, but can not be deleted and re-created without getting the new ID
* On new board you must create the following labels worded like this to enable specific features.  Just like lists you can rename or re-color labels at any time after initial configuration, but if you delete them and re-create them you will need to get the new label ID and update Tik-Tok's config.   Initially they must be named this way and can be any color.
  * ROLL-OVER 
  * TEMPLATE CARD DO NOT MOVE
  * DEMO
  * Training
  * Tik-Tok Hush

* Board must have custom fields power-up enabled and the following fields created.  These can not be renamed. (Trello limitation)
  * Text Field called:  Burndown
  * Text Field called:  Sprint
  
#### Column/List Name initial requirements for auto-config
* Backlog
* Upcoming
* Scoped
* Next Sprint
* Ready for Work
* Working
* Ready for Review (PR)
* Done

#### Have Tik-Tok start your config for you
To find all the unique Trello UID's for the TOML config file, you can ask Tik-Tok to find them for you.  This will help you build your config file.
`@Tik-Tok build a configuration file for [&lt;trello board id&gt;]`.  He will then DM you the results in slack.

### CRON JOBS
* All cron jobs for Tik-Tok are contained in the toml file called `crons.toml` stored in Tik-Tok's working directory.  This file can be edited at any time and you can issue a `reload cronjobs` command to Tik-Tok and he will re-read the file and load the new changes.   He will log errors around this in whichever slack channel you've specified logging to go to.

#### Available Cron Functions
* Cron functions are now listed in the Tik-Tok Help Wiki here: https://github.com/srv1054/bots-Tik-Tok/wiki/Tik-Tok-Help

### PERMISSIONS
For specific tasks (such as shutdown) Tik-Tok will require you to have permissions.  Currently some tasks are for `admin` only.  This is a list of slack users contained in the `actions.go` file in the Permissions Function. Eventually this will be moved to a datastore.   Other permissions (such as launching a new sprint) will require the user to a member of a specific private slack channel.


