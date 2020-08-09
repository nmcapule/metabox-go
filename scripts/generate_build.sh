#!/bin/bash

# Generates BUILD and WORKSPACE and repositories.bzl.

## Generate go.mod.
go mod tidy

## Add dependency repositories to repositories.bzl
bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=repositories.bzl%go_repositories

## Update Go BUILD files.
bazel run //:gazelle .
