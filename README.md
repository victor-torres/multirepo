# multirepo

An alternative to git submodules and monorepos written in Go

## Install

TODO

## Sync

Create a `repositories.yaml` file inside your main repository,
and define your git dependencies like this:

```yaml
repositories:
  fastapi:
    path: /tmp/multirepo/fastapi
    url: https://github.com/tiangolo/fastapi.git
    tag: 0.111.0
  pytest:
    path: /tmp/multirepo/pytest
    url: https://github.com/pytest-dev/pytest.git
    branch: main
  pydantic:
    path: /tmp/multirepo/pydantic
    url: https://github.com/pydantic/pydantic.git
    commit: 7061f36
```

Sync the config file with your filesystem with the following command:

```shell
multirepo sync
```

This command should output something like this:

```shell
3 repositories detected

➜ /tmp/multirepo/fastapi$ git clone https://github.com/tiangolo/fastapi.git
Cloning into '/tmp/multirepo/fastapi'...
➜ /tmp/multirepo/fastapi$ git checkout 0.111.0
Note: switching to '0.111.0'.
[...]
HEAD is now at 1c3e6918 📝 Update release notes

➜ /tmp/multirepo/pydantic$ git clone https://github.com/pydantic/pydantic.git
Cloning into '/tmp/multirepo/pydantic'...
➜ /tmp/multirepo/pydantic$ git checkout 7061f36
Note: switching to '7061f36'.
[...]
HEAD is now at 7061f36b fix json schema doc link (#9405)

➜ /tmp/multirepo/pytest$ git clone https://github.com/pytest-dev/pytest.git
Cloning into '/tmp/multirepo/pytest'...
➜ /tmp/multirepo/pytest$ git checkout main
Already on 'main'
Your branch is up to date with 'origin/main'.

3 repositories detected

fastapi     ✔ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) 
pydantic    ✔ 7061f36bc721ef4f173ef8f2e098f25e1eaea705  
pytest      ✔ 93dd34e76d9c687d1c249fe8cf94bdf46813f783 (branch: main) 
```

You can sync repositories how many times you want:

- If the repository does not exist, it will be cloned.
- If the repository is dirty, we'll abort the operation.

## Status

For your convenience, we run a status command everytime you sync a `repositories.yaml` file.

But you can also query the status of all git repositories on demand:

```shell
multirepo status
```

This will produce a similar output:

```shell
3 repositories detected

fastapi     ✔ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) 
pydantic    ✔ 7061f36bc721ef4f173ef8f2e098f25e1eaea705  
pytest      ✔ 93dd34e76d9c687d1c249fe8cf94bdf46813f783 (branch: main) 
```

If you make changes to the repository:

```shell
touch /tmp/multirepo/fastapi/new-file.txt
echo "new content" > /tmp/multirepo/pydantic/README.md
git -C /tmp/multirepo/pytest/ checkout 940b78232e48c34501cfe6e0bfd0ea6d64f4521b
```

The status output should now reflect those changes:

```shell
3 repositories detected

fastapi     ✗ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) (uncommitted changes)
pydantic    ✗ 7061f36bc721ef4f173ef8f2e098f25e1eaea705 (uncommitted changes)
pytest      ✗ 940b78232e48c34501cfe6e0bfd0ea6d64f4521b (branch: main ➜ 940b782) 
```

If you're just on the wrong reference, you can simply run the sync command again.
But it might be possible that you'll need to perform additional operations if:
- you have uncommited changes
  - you need to stash or revert your changes
- you don't have all the references locally
  - you need to fetch references from remote

## Run

Guess what: there's a convenient method to run a command on all of them at once with:

```shell
multirepo run --all git stash
```

This should loop through all the repos, execute the specified command, and print the output:
```shell
3 repositories detected

➜ /tmp/multirepo/fastapi$ git stash
No local changes to save

➜ /tmp/multirepo/pydantic$ git stash
Saved working directory and index state WIP on (no branch): 7061f36b fix json schema doc link (#9405)

➜ /tmp/multirepo/pytest$ git stash
No local changes to save
```

After stashing all repositories, and running the sync command again, you might notice that
one of our repositories still have uncommited changes, and another one still has a wrong reference:

```shell
3 repositories detected

fastapi     ✗ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) (uncommitted changes)
pydantic    ✔ 7061f36bc721ef4f173ef8f2e098f25e1eaea705  
pytest      ✗ 940b78232e48c34501cfe6e0bfd0ea6d64f4521b (branch: main ➜ 940b782) 
```

Let's try to sync the repositories again:

```shell
multirepo sync
```

Let's check the output this time:

```shell
fastapi     ✗ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) (uncommitted changes)
pydantic    ✔ 7061f36bc721ef4f173ef8f2e098f25e1eaea705  
pytest      ✔ 93dd34e76d9c687d1c249fe8cf94bdf46813f783 (branch: main) 
```

The wrong branch has been fixed, but we still have an uncommited change being reported.
It's also possible to run commands in individual repositories. Let's try to find out what's happening:

```shell
multirepo run fastapi git status
```

Let's check the output:

```shell
3 repositories detected

➜ /tmp/multirepo/fastapi$ git status
HEAD detached at 0.111.0
Untracked files:
  (use "git add <file>..." to include in what will be committed)
        new-file.txt

nothing added to commit but untracked files present (use "git add" to track)
```

Ok, that's an untracked file that we've just added some steps ago.
Let's remove it and check the status again:

```shell
multirepo run fastapi rm new-file.txt
```

Now it should be looking good again :)

```shell
3 repositories detected

fastapi     ✔ 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63 (tag: 0.111.0) 
pydantic    ✔ 7061f36bc721ef4f173ef8f2e098f25e1eaea705  
pytest      ✔ 93dd34e76d9c687d1c249fe8cf94bdf46813f783 (branch: main) 
```

## Environment variables

It's possible to use environment variables to define repository paths. For example:

```yaml
repositories:
  fastapi:
    path: $MY_BASE_DIRECTORY/fastapi
    url: https://github.com/tiangolo/fastapi.git
    tag: 0.111.0
```

If there's a `.env` file in the same location as the `repositories.yaml` file,
we'll also try to load your environment variables automatically,
so you don't need to manually export them.
