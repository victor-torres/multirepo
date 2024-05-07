# git-subrepos

An alternative to git submodules written in Go

## How it works

You create a `subrepos.yaml` file inside your main repository,
and define your git dependencies like this:

```yaml
repos:
  alice-repo:
    path: alice-repo
    url: git@bitbucket.org:company/alice-repo.git
    branch: alice-branch
  bob-repo:
    path: bob-repo
    url: git@bitbucket.org:company/bob-repo.git
    tag: bob-tag
  chad-repo:
    path: chad-repo
    url: git@github.com:company/chad-repo.git
    commit: 770a815107480f2f39b2359f381b3e20cc3c0af0
  dani-repo:
    path: dani-repo
    url: git@github.com:company/dani-repo.git
```

If cloning inside the main repository,
it's advised to add the directory names to the `.gitignore` file:

```gitignore
alice-repo/
bob-repo/
chad-repo/
dani-repo/
```

Now we need to sync the config file with your filesystem.
In other words,
we need to clone the repositories,
and checkout the reference
specified in the `subrepos.yaml` file.

You can sync repositories how many times you want:

- If the repository does not exist, it will be cloned.
- If the repository is dirty, we'll abort the operation.

## Motivation

I was working on a project with lots of git submodules,
and maintaining those submodules up-to-date was an awful task :)

I've decided to create this project to experiment with the idea
of treating our dependencies like a regular git repository,
and not like a git submodule anymore.

This repository is on its very early stages,
and I'm not very experienced with go,
it's one of my first projects using the language.

In other words, this is an experiment and an opportunity to learn.
