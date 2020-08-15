# Overview

Metabox is a backup and restore tool that aims for:

1. Recording backup information in a VCS-friendly txt file.
2. Save backups in multiple silos: remote, local, s3, etc

# Usage

> NOTE: This might change since Go's "flag" library isn't up to par.

## Backup

```sh
$ metabox -config ./examples/ouroboros/ouroboros.metabox.yml backup
```

## Restore

```sh
$ metabox -config ./examples/ouroboros/ouroboros.metabox.yml restore
```

# Roadmap

None, it's too early and still shitty. Maybe a checklist if things to do first:

-   [ ] Config option to store backups to another local path
-   [ ] Config option to store backups to a remote computer
-   [ ] Config option to store backups to Amazon S3
-   [ ] Use cache. No longer compress / download if it's already in the cache
-   [ ] Multiple values for backup config option
-   [ ] Merge / restore strategies: merge, nuke, existing_only, nonexisting_only
-   [ ] Fix cli to use spf13/cobra for sane invocations
