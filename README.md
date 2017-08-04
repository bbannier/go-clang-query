Tis tool runs clang-query in parallel over a set of files and deduplicates the
result set.

This is implemented in Go for the lulz.

Install
-------

    % glide install
    % go build

Usage
-----

This works almost like `clang-query` and passed arguments through to it.
Queries are read from stdin.

The call syntax is

    % go-clang-query <file>... [--- <clang-query options>]
