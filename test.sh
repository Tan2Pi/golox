#!/bin/bash

set -eux

rm -f test/comments/unicode.lox
dart tool/bin/test.dart chap13_inheritance -i glox || exit 1