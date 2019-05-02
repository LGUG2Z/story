# story
[![Go Report Card](https://goreportcard.com/badge/github.com/lgug2z/story)](https://goreportcard.com/report/github.com/lgug2z/story)
[![Maintainability](https://api.codeclimate.com/v1/badges/ed8cb042219f695c8436/maintainability)](https://codeclimate.com/github/LGUG2Z/story/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/ed8cb042219f695c8436/test_coverage)](https://codeclimate.com/github/LGUG2Z/story/test_coverage)
[![Build Status](https://travis-ci.org/LGUG2Z/story.svg?branch=master)](https://travis-ci.org/LGUG2Z/story)

`story` works as a layer on top of [meta](https://github.com/mateodelnorte/meta) to aid development, continuous integration,
testing, container building and deployments when working with meta-repos containing a large number of inter-dependent
`node` projects.

- [Installation](#installation)
  * [Go Get](#go-get)
  * [Homebrew](#homebrew)
  * [Bash Completion](#bash-completion)
- [The .meta File](#the-meta-file)
  * [The trunk `.meta` file](#the-trunk--meta--file)
  * [The `story` `.meta` file](#the--story---meta--file)
- [Commands](#commands)
- [Workflow Examples](#workflow-examples)
  * [Starting a New Story](#starting-a-new-story)
  * [Updating From Trunk Branches](#updating-from-trunk-branches)
  * [Migrating Existing Branches to a New Story](#migrating-existing-branches-to-a-new-story)
  * [Switching Stories](#switching-stories)
  * [Merging Completed Stories](#merging-completed-stories)

# Installation
## Go Get
```bash
go get -u github.com/LGUG2Z/story
cd $GOPATH/src/github.com/LGUG2Z/story
make install
```

## Homebrew
```bash
brew tap LGUG2Z/tap
brew install LGUG2Z/tap/story
```

## Bash Completion
Add the following to your `.bashrc` or `.zshrc`
```bash
PROG=story source /usr/local/etc/bash_completion.d/story
```

# The .meta File
## The trunk `.meta` file
There are two types of `.meta` files used by `story`, which are both supersets of the `.meta` file used by
 [meta](https://github.com/mateodelnorte/meta):

The `.meta` file for the overall meta-repo includes two extra keys, `artifacts` and `organisation`:

```json
{
  "artifacts": {
    "api": false,
    "app": false
  },
  "organisation": "SecretOrg",
  "projects": {
    "api": "git@github.com:SecretOrg/api.git",
    "app": "git@github.com:SecretOrg/app.git",
    "lib-1": "git@github.com:SecretOrg/lib-1.git",
    "lib-2": "git@github.com:SecretOrg/lib-2.git"
  }
}
```

`artifacts` refers to projects that can be built and deployed, and should be set to `false` in the `.meta` file for a meta-repo.

`organisation` refers to the name of the organisation on GitHub where private repositories are hosted.

A JSONSchema for the trunk `.meta` file is available [here](meta.json).

## The `story` `.meta` file
The `.meta` file for stories includes a number of extra keys on top of those introduced above:
```json
{
  "story": "story/auth-endpoint",
  "organisation": "SecretOrg",
  "projects": {
    "api": "git@github.com:SecretOrg/api.git",
    "lib-2": "git@github.com:SecretOrg/lib-2.git"
  },
  "hashes": {
    "api": "c917d416366a04f2ec62c2e8aaee5bc740d8c6aa",
    "lib-2": "6bbe39ebe169c46eee7b2a6bc392e0b37e397a0e"
  },
  "blastRadius": {
    "api": null,
    "lib-2": ["api", "app"]
  },
  "artifacts": {
    "api": true,
    "app": true
  },
  "allProjects": {
    "api": "git@github.com:SecretOrg/api.git",
    "app": "git@github.com:SecretOrg/app.git",
    "lib-1": "git@github.com:SecretOrg/lib-1.git",
    "lib-2": "git@github.com:SecretOrg/lib-2.git"
  }
}
```
`story` refers to the name of the branch that will be checked out on any projects added to a story.

`projects` refers to the subset of projects that the story requires work to be done on.

`hashes` refers to the current commit hashes of each project at the time of a commit to the meta-repo.

`blastRadius` refers to projects in the meta-repo that can be impacted by made changes in the scope of the current story.

`allProjects` refers to the complete list of projects in the meta-repo.

This latter file is automatically generated and maintained by `story` commands. For example, adding or removing a project
to a story will update the `projects`, `hashes`, `blastRadius` and `artifacts` keys accordingly, and making a commit
across the meta-repo will update the `hashes` key before making a final commit to the meta-repo.

A JSONSchema for the `story` `.meta` file is available [here](story.json).

# Commands
```
NAME:
   story - A workflow tool for implementing stories across a node meta-repo

USAGE:
   story [global options] command [command options] [arguments...]

VERSION:
   0.3.4

AUTHOR:
   J. Iqbal <jade@beamery.com>

COMMANDS:
     create       Creates a new story
     load         Loads an existing story
     reset        Resets all story branches to trunk branches
     add          Adds a project to the current story
     remove       Removes a project from the current story
     list         Shows a list of projects added to the current story
     blastradius  Shows a list of current story's blast radius
     artifacts    Shows a list of artifacts to be built and deployed for the current story
     commit       Commits code across the current story
     push         Pushes commits across the current story
     unpin        Unpins code in the current story
     pin          Pins code in the current story
     prepare      Prepares a story for merges to trunk
     update       Updates code from the upstream master branch across the current story
     merge        Merges prepared code to master branches across the current story
     pr           Opens pull requests for the current story
     help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --trunk value  (default: "master") [$STORY_TRUNK]
   --help, -h     show help
   --version, -v  print the version
```

# Workflow Examples
## Starting a New Story
```bash
# navigate to your metarepo
cd ~/metarepo

# create the story
story create story/sso-login

# add the repoos you will be working on
story add login-service marketing-app

# stage the changes made by creating and adding to the story
meta git add -p

# make an initial commit across all the repos
story commit -m "add login-service and marketing-app, update package.json files"

# push branches in all story repos
story push

# open PRs linked to a central issue
story pr --issue https://github.com/SecretOrg/tracking-board/issues/9
```

## Updating From Trunk Branches
```bash
# load the story
story load story/email-login

# merge in changes from trunk on every repo in the story
story update
```

## Migrating Existing Branches to a New Story
```bash
# create a new story
story new story/otp-login

# merge in changes from other existing on every repo in the story
story update --from-branch feature/otp-login
```

## Switching Stories
```bash
# reset to the trunk branches
story reset

# load another story
story load story/sso-acl
```

## Merging Completed Stories
```bash
# load the story
story load story/sso-login

# make sure we have the latest from master
story update

##########
# EITHER #
##########
# reset the package.json dependencies to point to master
story unpin

##########
# OR     #
##########
# pin the package.json dependencies to a specific commit hash
story pin

# archive the story manifest and reset the .meta file for merge
story prepare

# push final changes to the story branches
# still on branch story/sso-login at this point
story push --from-manifest story/sso-login

# merge the story to the trunk branch across all story repos
story merge

# push just the repos that were changed in the story post-merge
# on master branch at this point
story push --from-manifest story/sso-login
```