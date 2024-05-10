# multirepo

An alternative to git submodules and monorepos written in Go

## How it works

You create a `repositories.yaml` file inside your main repository,
and define your git dependencies like this:

```yaml
repositories:
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
    branch: main
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

```shell
$ multirepo sync
4 repositories detected

➜ alice-repo$ git clone git@bitbucket.org:company/alice-repo.git
➜ alice-repo$ git clone git@bitbucket.org:company/alice-repo.git
Cloning into 'alice-repo'...
➜ alice-repo$ git checkout main
branch 'alice-branch' set up to track 'origin/alice-branch'.
Switched to a new branch 'alice-branch'

➜ bob-repo$ git clone git@bitbucket.org:company/bob-repo.git
➜ bob-repo$ git clone git@bitbucket.org:company/bob-repo.git
Cloning into 'bob-repo'...
➜ bob-repo$ git checkout main
Note: switching to 'bob-tag'.
HEAD is now at 84227b7 Update .gitignore

➜ chad-repo$ git clone git@github.com:company/chad-repo.git
➜ chad-repo$ git clone git@github.com:company/chad-repo.git
Cloning into 'chad-repo'...
➜ chad-repo$ git checkout main
Note: switching to '770a815107480f2f39b2359f381b3e20cc3c0af0'.
HEAD is now at 770a815 Update Dockerfile

➜ dani-repo$ git clone git@github.com:company/dani-repo.git
➜ dani-repo$ git clone git@github.com:company/dani-repo.git
Cloning into 'dani-repo'...
➜ dani-repo$ git checkout main
Already on 'main'
Your branch is up to date with 'origin/main'.

4 repositories detected

alice-repo                    ✔ commit 8c32bbbf474beb4fd95f6a57c6726adad5c946c7 (HEAD, -> alice-branch, origin/alice-branch)
bob-repo                      ✔ commit 84227b7a73b212dfc8fe129475a82098a393842c (HEAD, tag: bob-tag)
chad-repo                     ✔ commit 770a815107480f2f39b2359f381b3e20cc3c0af0
dani-repo                     ✔ commit 93b5706409a74b1da30623f1036a2da87e856c79 (HEAD -> main, origin/main)
```

For your convenience, we run a status command everytime you sync a project.

You can sync repositories how many times you want:

- If the repository does not exist, it will be cloned.
- If the repository is dirty, we'll abort the operation.

You can also query the status of all git repositories:

```shell
$ multirepo status
4 repositories detected

alice-repo                    ✗ commit 8c32bbbf474beb4fd95f6a57c6726adad5c946c7 (HEAD, -> alice-branch, origin/alice-branch) uncommited changes
bob-repo                      ✔ commit 84227b7a73b212dfc8fe129475a82098a393842c (HEAD, tag: bob-tag)
chad-repo                     ✗ commit cbc18a7a2ae9df9ea11b94cb391c33fa6e464e53
dani-repo                     ✔ commit 93b5706409a74b1da30623f1036a2da87e856c79 (HEAD -> main, origin/main)
```

And there's a convenient method to run a command on all of them at once with:

```shell
$ multirepo run git stash
➜ alice-repo$ git stash
Saved working directory and index state WIP on main: 8c32bbb Huge refactor

➜ bob-repo$ git stash
No local changes to save

➜ chad-repo$ git stash
No local changes to save

➜ dani-repo$ git stash
No local changes to save
```

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
