#!/bin/sh

# $ monitfiles -f go  -p . -s ./monitfiles.sh -v -b -max 10
#
# Using https://github.com/pedromg/monitfiles
# 2021/04/23 01:07:20 * entering directory: mailsender
# 2021/04/23 01:07:20    > checking .gitignore
# 2021/04/23 01:07:20    > checking .travis.yml
# 2021/04/23 01:07:20    > checking LICENSE
# 2021/04/23 01:07:20    > checking Makefile
# 2021/04/23 01:07:20    > checking README.md
# 2021/04/23 01:07:20    > checking go.mod
# 2021/04/23 01:07:20    > checking go.sum
# 2021/04/23 01:07:20    > checking mailsender.go
# 2021/04/23 01:07:20    + adding mailsender.go (2021-04-23 00:59:33.676433605 +0100 WEST)
# 2021/04/23 01:07:20    > checking mailsender_test.go
# 2021/04/23 01:07:20    + adding mailsender_test.go (2021-04-23 00:49:32.832705641 +0100 WEST)
# 2021/04/23 01:07:20    > checking monitfiles.sh
# 2021/04/23 01:07:20 * entering directory: test
# 2021/04/23 01:07:20    > checking mailsender.json
# 2021/04/23 01:07:20    > checking mailsender_renamed.json
# 2021/04/23 01:07:20    > checking mailsender_too_big.json
# 2021/04/23 01:07:20 ************************************************
# 2021/04/23 01:07:20 Root path: .../mailsender
# 2021/04/23 01:07:20 File types: [go]
# 2021/04/23 01:07:20 File types with no extension ? false
# 2021/04/23 01:07:20 Exclude dot dirs ? true
# 2021/04/23 01:07:20 Max number of files: 10
# 2021/04/23 01:07:20 Blocking ? true
# 2021/04/23 01:07:20 Verbose ? true
# 2021/04/23 01:07:20 Interval: 2 seconds
# 2021/04/23 01:07:20 Script: ./monitfiles.sh
# 2021/04/23 01:07:20 Web: false
# 2021/04/23 01:07:20 Number of directories scanned: 2
# 2021/04/23 01:07:20 Number of files added and being monitored: 2
# 2021/04/23 01:07:20 ************************************************

make testall 2>&1
