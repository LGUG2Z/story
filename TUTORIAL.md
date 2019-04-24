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


## Switching Stories
```bash
# reset to the trunk branches
story reset

# load another story
story load story/sso-acl
```

# Merging Completed Stories
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

# merge the story to the trunk branch across all story repos
story merge

# push just the repos that were changed in the story post-merge
story merge --from-manifest story/sso-login
```
