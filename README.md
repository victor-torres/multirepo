# multirepo

An alternative to git submodules and monorepos written in Go

## Install

You have two options to acquire the binary executable for this project:
- download pre-compiled releases from Github for Linux, macOS, or Windows ([click here](https://github.com/victor-torres/multirepo/releases/latest))
- clone this repository and compile the project locally using go

Once you have the executable binary on your machine,
copy or link it to somewhere like `/usr/local/bin`
or add its containing directory to your PATH environment variable. 

Example:

```shell
cd /tmp
wget https://github.com/victor-torres/multirepo/releases/download/v3.1.0/multirepo-v3.1.0-darwin-arm64.tar.gz
tar -xvf multirepo-v3.1.0-darwin-arm64.tar.gz
sudo mv multirepo /usr/local/bin/multirepo
```

Or compiling locally (depends on go-lang):

```shell
git clone https://github.com/victor-torres/multirepo.git
go get
go build
sudo ln -sf multirepo /usr/local/bin/multirepo
```

**Tip:** you can also create an alias in your `.bashrc` or `.zshrc` file
to make it easier to call our executable from your terminal. For example:

```shell
echo "alias mr=multirepo" >> ~/.$(lsof -p $$ | cut -d " " -f 1 | tail -n 1)rc
source ~/.$(lsof -p $$ | cut -d " " -f 1 | tail -n 1)rc
```

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

You can also optionally specify these flags:

- `--force` or `-f` to discard uncommited changes permanently
- `--recurse` or `-r` to recursively checkout submodules

### Shallow clones

For large dependencies you can limit how much history is cloned with the
optional `depth` field (it maps to `git clone --depth`):

```yaml
repositories:
  fastapi:
    path: /tmp/multirepo/fastapi
    url: https://github.com/tiangolo/fastapi.git
    tag: 0.111.0
    depth: 1
```

Omit `depth` (or set it to 0) for a full clone. If the configured
reference is not included in the shallow history, sync automatically
falls back to fetching it.

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

### Machine-readable status

For scripting, CI gates, or feeding an AI agent, status can emit JSON or
a markdown table instead of the colored human output:

```shell
multirepo status --json
multirepo status --md
```

The JSON output is an array with one object per repository, including
`name`, `path`, `found`, `commit`, `branch`, `tags`, `dirty`,
`target_type`, `target_name`, `ref_matches`, and `in_sync` (true when
the repository is on the configured reference with a clean tree).

If you're just on the wrong reference, you can simply run the sync command again.
Sync automatically fetches from the remote when the configured reference
is not available locally, and fast-forwards branch targets to their
remote counterpart. But you'll need to perform additional operations if:
- you have uncommited changes
  - you need to stash or revert your changes (or sync with `--force`)

## Lock

Branch (and tag) targets are moving references. To make a checkout
reproducible — for teammates or CI — pin every repository to its exact
current commit:

```shell
multirepo lock
```

This writes a `repositories.lock` file next to `repositories.yaml`:

```yaml
repositories:
  fastapi:
    commit: 1c3e6918750ccb3f20ea260e9a4238ce2c0e5f63
  pytest:
    commit: 93dd34e76d9c687d1c249fe8cf94bdf46813f783
```

Commit the lock file to your main repository. Anyone can then reproduce
that exact state, regardless of where branches have moved since:

```shell
multirepo sync --locked
```

Run `multirepo lock` again whenever you want to roll the pins forward.

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
