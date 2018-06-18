# story
[![Go Report Card](https://goreportcard.com/badge/github.com/lgug2z/story)](https://goreportcard.com/report/github.com/lgug2z/story)
[![Maintainability](https://api.codeclimate.com/v1/badges/ed8cb042219f695c8436/maintainability)](https://codeclimate.com/github/LGUG2Z/story/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/ed8cb042219f695c8436/test_coverage)](https://codeclimate.com/github/LGUG2Z/story/test_coverage)
[![Build Status](https://travis-ci.org/LGUG2Z/story.svg?branch=master)](https://travis-ci.org/LGUG2Z/story)

`story` works as a layer on top of [meta](https://github.com/mateodelnorte/meta) to aid development, continuous integration,
testing, container building and deployments when working with meta-repos containing a large number of inter-dependent
`node` projects.

- [The .meta File](#the-meta-file)
- [Commands](#commands)
  * [Create](#create)
  * [Load](#load)
  * [Reset](#reset)
  * [List](#list)
  * [Artifacts](#artifacts)
  * [Add](#add)
  * [Remove](#remove)
  * [Commit](#commit)

# The .meta File
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
  "blast-radius": {
    "api": null,
    "lib-2": ["api", "app"]
  },
  "artifacts": {
    "api": true,
    "app": true
  },
  "all-projects": {
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

`blast-radius` refers to projects in the meta-repo that can be impacted by made changes in the scope of the current story.

`all-projects` refers to the complete list of projects in the meta-repo.

This latter file is automatically generated and maintained by `story` commands. For example, adding or removing a project
to a story will update the `projects`, `hashes`, `blast-radius` and `artifacts` keys accordingly, and making a commit
across the meta-repo will update the `hashes` key before making a final commit to the meta-repo.

# Commands
## Create
`story create [story-name]` will:
* Checkout a new branch on the meta-repo repository with the story name
* Create a new `.meta` file for the story

## Load
`story load [story-name]` will:
* Checkout the desired branch on the meta-repo if it exists
* Checkout the desired branch on all projects in the story if they exist

## Reset
`story reset` will:
* Checkout the trunk branch (default `master`) on all projects of a story

## List
`story list` will:
* Print a list of all projects in the current story

## Artifacts
`story artifacts` will:
* Print a list of all projects that should be built and deployed in the current story

## Add
`story add [projects]` will:
* Add the projects to the `.meta` file
* Checkout a story branch for the projects
* Update the blast radius of the story
* Update the commit hashes of projects in the story
* Update `package.json` files to pin references to projects in the story to the story branch

## Remove
`story remove [projects]` will:
* Remove the projects from the `.meta` file
* Delete story branches for the projects
* Update the blast radius of the story
* Update the commit hashes of projects in the story
* Update `package.json` files to unpin references to projects in the story from the story branch

## Commit
`story commit [-m "commit msg"]` will:
* Commit staged files in all story projects with the given commit message
* Update the commit hashes of projects in the story
* Commit the updated `.meta` file to the meta-repo, with GitHub links to each project commit in the extended commit message
    * ```
      commit 8f59319fc2d0403199fb2c6148b5ab67919424f3 (HEAD -> story/commit-example)
      Author: jiqb <gthbji@ml1.net>
      Date:   Mon Jun 18 08:23:30 2018 +0100

      This is the message passed via the -m flag of the commit command 

      https://github.com/SecretOrg/api/commit/ad4f419b7d65292ef28ab8d1d3ef4346a6bdebe4
      https://github.com/SecretOrg/lib-2/commit/e1f99366bcc71df8bccf6f3df66271a319c33240
    ```
* Update `package.json` files to unpin references to projects in the story from the story branch
