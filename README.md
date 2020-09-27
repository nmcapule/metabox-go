# Overview

Metabox is a backup and restore tool that can:

1. Save backups in multiple silos: remote, local, s3, etc
2. Record every unique backups in a VCS-friendly txt file.

Metabox is a re-implementation of a shell script of the same name archived in
[nmcapule/metabox](https://github.com/nmcapule/metabox/blob/master/metabox) and
is also inspired by [huacnlee/gobackup](https://github.com/huacnlee/gobackup).

## Installation

You need `go` toolchain installed.

```sh
$ go install github.com/nmcapule/metabox-go
```

If $GOBIN is in your $PATH, you can test it with:

```sh
$ metabox-go --help
```

## How it works

At its core, metabox is just a backup/restore tool. To set it up, you will need a
`yaml` file describing which files to backup and how will it be stored.

For example:

```yml
target:
    prefix_path: ./target
    includes:
        - "**/*"
    excludes:
        - ".git/"
        - "examples/"
```

> See the `examples` folder for more usage examples.

When executed with `metabox backup`, it will:

1. Archive all contents of the `target` folder into a `<hash>.tar.gz`, where _hash_
   is the `sha256` hash of all the included files in the folder.
2. Put the generated `.tar.gz` in workspace folder called `./cache`.
3. Create an entry for this particular backup in a `backups.txt` file. One line in
   this file corresponds to exactly one backup.

The first step is important since it determines if this particular set of target
files has been backed up before. If yes, then `metabox` will skip step #2 and #3.

## The `backups.txt` file

When backing up/restoring files, `metabox` will look first in the `backups.txt` file
usually in the same folder as the metabox yaml file.

This file tracks all the made backups to check if a backup for this set of files
already exists. The `backups.txt` file is also important for quickly searching backups
via tags -- which can be attached while doing a `metabox backup` command.

Each line in the `backups.txt` file corresponds to exactly one backup. This convention
makes it very easy to embed to distributed VCS with all the merging and parallel
workflows going on.

Here is an example of a backup file:

```txt
824f4cb43a55bef5611b555dd305126f 1598781213 anonymous branch:development,database:default
76fcbf4753af490f4ba8f52758ddb107 1597515810 anonymous -
98838a715cc75d24359a7632b26627dc 1598786787 anonymous hello,world
```

Each column corresponds to:

-   **Files/Folders hash**
-   **Created timestamp**
-   **Creator**
-   **Tags**

## `*.metabox.yml` config flags

| Flag                            | Values    | Description                                                |
| :------------------------------ | :-------- | :--------------------------------------------------------- |
| version                         | 0.1       | Placeholder for future proofing                            |
| workspace                       | Object    | Specifier for the current workspace                        |
| workspace.root_path             | directory | Working directory. Default: directory of yml file          |
| workspace.cache_path            | directory | Folder name of cache relative to working directory         |
| workspace.versions_path         | file      | Filename of version tracker. Default: `backups.txt`        |
| workspace.hooks.pre_backup      | commands  | List of commands to execute before backup process          |
| workspace.hooks.post_backup     | commands  | List of commands to execute after backup process           |
| workspace.hooks.pre_restore     | commands  | List of commands to execute before restore process         |
| workspace.hooks.post_restore    | commands  | List of commands to execute after restore process          |
| workspace.options               | Object    | Configuration on how to archive                            |
| workspace.options.compress      | tgz       | Compression algorithm                                      |
| workspace.options.hash          | md5       | Hashing algorithm to use when hashing target files/folders |
| target                          | Object    | Specifier for target folder to backup                      |
| target.prefix_path              | directory | Target folder relative to root                             |
| target.includes                 | matchers  | File matchers similar to `.gitignore`. Defaults to all     |
| target.excludes                 | matchers  | File exclusions similar to `.gitignore`. Defaults to none  |
| backups                         | Array     | Specifier for how to store backups.                        |
| backups.\*.driver               | driver    | Can be `s3` or `local`                                     |
| backups.\*.s3                   | Object    | Specifier for how to store backups in s3 if `driver: s3`   |
| backups.\*.s3.prefix_path       | directory | Prefix path when storing to s3 bucket                      |
| backups.\*.s3.access_key_id     | string    | AWS access key ID                                          |
| backups.\*.s3.secret_access_key | string    | AWS secret access key                                      |
| backups.\*.s3.region            | string    | AWS region specifier                                       |
| backups.\*.s3.bucket            | string    | Name of S3 bucket to store the backups                     |
| backups.\*.s3.endpoint          | string    | Assign value to specify custom S3 endpoint (e.g. linode)   |
| backups.\*.local                | Object    | Specifier for backups in local if `driver: local`          |
| backups.\*.local.path           | Object    | Prefix path when storing to local                          |

> You can checkout `config/config.go` for a possibly full list.

# Usage

Make sure `metabox-go` is reachable in your \$PATH env.

## Backup

### Basic backup

```sh
$ metabox-go backup ./examples/ouroboros/ouroboros.metabox.yml
```

### Backup with tags

```sh
$ metabox-go backup ./examples/ouroboros/ouroboros.metabox.yml -t hello -t branch:development
```

## Restore

### Basic restore

```sh
$ metabox-go restore ./examples/ouroboros/ouroboros.metabox.yml
```

### Restore latest backup matching the specified tags

```sh
$ metabox-go restore ./examples/ouroboros/ouroboros.metabox.yml -t hello -t branch:development
```

# Roadmap

None, it's too early and still shitty. Maybe a checklist if things to do first:

-   [x] Config option to store backups to another local path
-   [ ] Config option to store backups to a remote computer
-   [x] Config option to store backups to Amazon S3
-   [x] Use cache. No longer compress / download if it's already in the cache
-   [x] Multiple values for backup config option
-   [ ] Merge / restore strategies: merge, nuke, existing_only, nonexisting_only
-   [x] Fix cli to use spf13/cobra for sane invocations
-   [ ] Allow restore command by specifying hash
-   [ ] Automated unit tests

# FAQs

## Why are you using bazel? You don't even need it??

I wanted to mess around with bazel.
