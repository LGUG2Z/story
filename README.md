# story
`story` is workflow tool for implementing stories across multiple `node` projects in a meta-repo
powered by [meta](https://github.com/mateodelnorte/meta).

## Overview
[meta](https://github.com/mateodelnorte/meta) is a great tool, but it gets messy to handle and slow
when dealing with a large number of projects in a meta-repo. Enter `story`, which allows you to
generate and manage subsets of your meta-repo on a per-story basis to allow for a faster and more
atomic development workflow that plays nicely with containerised builds and deployments.

## Walkthrough

Given a meta-repo with the following projects:
```
├── service-1
├── service-2
├── service-3
├── library-1
├── library-2
└── library-3
```

And some example `package.json` files within them:
```json
// service-1/package.json
{ 
  "dependencies": {
    "library-1": "git+ssh://git@github.com:SomeOrg/library-1.git",
    "library-2": "git+ssh://git@github.com:SomeOrg/library-2.git"
  }
}

```

```json
// service-2/package.json
{ 
  "dependencies": {
    "library-1": "git+ssh://git@github.com:SomeOrg/library-1.git",
    "library-2": "git+ssh://git@github.com:SomeOrg/library-2.git"
  }
}
```

```json
// library-3/package.json
{ 
  "dependencies": {
    "library-1": "git+ssh://git@github.com:SomeOrg/library-1.git"
  }
}
```


When we set up a new story and add projects to it:
```bash
# go to the meta-repo
cd meta-repo

# set a story
story set new-navigation
story add service-1
```

Then:
* a new story is created with a new `.meta`
```json
{
  "story": "new-navigation",
  "primaries": {
    "service-1": true
  },
  "projects": {
    "service-1": "git@github.com:SomeOrg/service-1.git",
    "library-1": "git@github.com:SomeOrg/library-1.git",
    "library-2": "git@github.com:SomeOrg/library-2.git"
  }
}
```

* service-1, library-1, library-2 are added to the story
* branch `new-navigation` is checked out on all three and the meta-repo
* `service-1/package.json` is updated

```json
// service-1/package.json
{ 
  "dependencies": {
    "library-1": "git+ssh://git@github.com:SomeOrg/library-1.git#new-navigation",
    "library-2": "git+ssh://git@github.com:SomeOrg/library-2.git#new-navigation"
  }
}
```

We can use the usual `meta` commands for local development

```bash
meta git add .
meta git commit -m "new nav is done!"
```

Assuming that we have only modified `service-1` and `library-1`, we can prune unmodified projects:

```
story prune
```

* the story is updated with the unmodified `library-2` removed and `service-1/package.json` modified
```json
{
  "story": "new-navigation",
  "primaries": {
    "service-1": true
  },
  "projects": {
    "service-1": "git@github.com:SomeOrg/service-1.git",
    "library-1": "git@github.com:SomeOrg/library-1.git"
  }
}
```

```json
// service-1/package.json
{ 
  "dependencies": {
    "library-1": "git+ssh://git@github.com:SomeOrg/library-1.git#new-navigation",
    "library-2": "git+ssh://git@github.com:SomeOrg/library-2.git"
  }
}
```

* the unmodified `new-navigation` branch on `library-2` is removed and master is checked out

Finally, to make sure that everything within the blast radius of our story is covered (ie. `library-3`,
which has `library-1` as a dependency):

```bash
beam blast
```

* the story is updated again with projects within the blast radius
```json
{
  "story": "new-navigation",
  "blast-radius": {
    "library-3": "git@github.com:SomeOrg/library-1.git"
  },
  "primaries": {
    "service-1": true
  },
  "projects": {
    "service-1": "git@github.com:SomeOrg/service-1.git",
    "library-1": "git@github.com:SomeOrg/library-1.git"
  }
}
```
* a `new-navigation` branch is checked out for `library-3`
* any project in the story that uses `library-3` has its' `package.json` updated
```json
// */package.json
{ 
  "dependencies": {
    "library-3": "git+ssh://git@github.com:SomeOrg/library-3.git#new-navigation",
  }
}
```

