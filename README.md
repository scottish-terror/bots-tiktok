## WALL*E Trello Scrum Master Extraordinaire

[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/forgecloud/ForgeCloud%2Fbots-wall-e%2Fbots-wall-e?branch=master&key=eyJhbGciOiJIUzI1NiJ9.NWE3MGI5MDM1OTJmYWQwMDAxNTE4YTY4.rnvqniYdWuESr3Iwxf_fjUZS4ZgdqUxZN8-SeKCL_H4&type=cf-1)]( https://g.codefresh.io/repositories/ForgeCloud/bots-wall-e/builds?filter=trigger:build;branch:master;service:5af21119b7f5c600015156d1~bots-wall-e)

WALL-E v3 is an always on services type bot and no longer executes on the CLI from Cron.   

WALL-E *required* CLI parameters
```
Usage of wall-e.go:
  -tkey  Trello API Key
  -ttoken  Trello API Token
  -slackhook  Slack API Webhook URL (required)  
  -slacktoken   Slack Bot Token
  -slackoauth  Slack App User OAuth Token (required to manage slack channels)
  -git  Github API Token
  -dbuser  Google Cloud SQL User
  -dbpass  Google Cloud SQL Password
```

WALL-E *optional* CLI parameters
```
  -nocron  Do not load built-in cronjobs on start.  crons.toml
```

Wall-E has a wall-e.toml file that has some base configuration parameters you need to set for him to run

Each team/board that needs Wall-E to operate on it must be configured a specific way and then contain a .TOML file with the settings and configuration for that board.   Below is how to configure your specific board .TOML file.   WALL-E does not need to be restarted if a new .TOML file is built and added to his working directory.  He will find it on command.  The name of the TOML file (minus the extension) is how you will refer to that when talking to Wall-E.  For example a file called  SAAS.toml will be referred to when talking to WALL-E as [SAAS].  

###  Wall-E Test Config
There's a test slack server called `ForgeBots` and we have Slack API keys for it so Wall-E can be run separately from ForgeRock slack for testing purposes. There is also a test configured Trello board called [SaaS-test] and it has an appropriately configured TOML file. That trello board is https://trello.com/b/34fsfToC/saas-eng-automation-test

###  Configuring your Trello Board
* Board must be created ahead of time you will need the boardID.  BoardID is in the URL when viewing the board. `https://trello.com/b/<BOARDID>/`
* Board must initially have colum/list names Matching the list below. *NOTE*: Columns/Lists can be renamed once the ID's have been acquired, but can not be deleted and re-created without getting the new ID
* On new board you must create the following labels worded like this to enable specific features.  Just like lists you can rename or re-color labels at any time after initial configuration, but if you delete them and re-create them you will need to get the new label ID and update wall-e's config.   Initially they must be named this way and can be any color.
  * ROLL-OVER 
  * TEMPLATE CARD DO NOT MOVE
  * DEMO
  * Training
  * Wall-E Hush

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

#### Have Wall-E start your config for you
To find all the unique Trello UID's for the TOML config file, you can ask wall-e to find them for you.  This will help you build your config file.
`@WALL-E build a configuration file for [&lt;trello board id&gt;]`.  He will then DM you the results in slack.

### CRON JOBS
* All cron jobs for Wall-E are contained in the toml file called `crons.toml` stored in Wall-E's working directory.  This file can be edited at any time and you can issue a `reload cronjobs` command to Wall-E and he will re-read the file and load the new changes.   He will log errors around this in whichever slack channel you've specified logging to go to.

#### Available Cron Functions
* Cron functions are now listed in the Wall-E Help Wiki here: https://github.com/ForgeCloud/bots-wall-e/wiki/Wall*E-Help

### PERMISSIONS
For specific tasks (such as shutdown) WALL-E will require you to have permissions.  Currently some tasks are for `admin` only.  This is a list of slack users contained in the `actions.go` file in the Permissions Function. Eventually this will be moved to a datastore.   Other permissions (such as launching a new sprint) will require the user to a member of a specific private slack channel.


