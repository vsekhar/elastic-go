# Elastic Go

Elastic Go is a fork of the Go programming language that supports building
binaries that utilize a remote API for runtime services (memory, scheduling,
etc.). This is a first step to running binaries on cloud infrastructure in a way
that scales automatically as required.

The project is active but not yet functional.

## Hacking

Install Go 1.5 or later to your system (for bootstrapping). You can download the latest at https://golang.org/dl/.

Set `GOBOOTSTRAP` to the location of your go installation (e.g. `/usr/local/go`).

Clone this repo.

Add the [gofmt pre-commit hook](https://golang.org/misc/git/pre-commit) to your repo.

## Testing

To run the full compiler test suite, run `all.bash` from within the `src` directory.

The full compiler test suite should pass before committing code.

For a faster code-test loop, you can run tests specific to remote compilation using:

    $ GOTESTONLY=testremote ./all.bash

These tests are found in `/misc/remote/test.bash` but cannot be invoked from that script directly (doing so would incorrectly use your system's Go installation rather than a freshly compiled one from this repo).

## Merging upstream
To keep things clean, all development of Elastic Go happens on the `dev.remote` branch. All other branches follow the corresponding branch in the upstream repo.

Add the original repo as upstream and fetch:

    $ git remote add -f upstream https://go.googlesource.com/go

Fetch and merge to master, then merge to dev.remote:

    $ git checkout master
    $ git merge upstream/master
    $ git checkout dev.remote
    $ git merge master

The upstream README follows.

---

# The Go Programming Language

Go is an open source programming language that makes it easy to build simple,
reliable, and efficient software.

![Gopher image](doc/gopher/fiveyears.jpg)
*Gopher image by [Renee French][rf], licensed under [Creative Commons 3.0 Attributions license][cc3-by].*

Our canonical Git repository is located at https://go.googlesource.com/go.
There is a mirror of the repository at https://github.com/golang/go.

Unless otherwise noted, the Go source files are distributed under the
BSD-style license found in the LICENSE file.

### Download and Install

#### Binary Distributions

Official binary distributions are available at https://golang.org/dl/.

After downloading a binary release, visit https://golang.org/doc/install
or load doc/install.html in your web browser for installation
instructions.

#### Install From Source

If a binary distribution is not available for your combination of
operating system and architecture, visit
https://golang.org/doc/install/source or load doc/install-source.html
in your web browser for source installation instructions.

### Contributing

Go is the work of hundreds of contributors. We appreciate your help!

To contribute, please read the contribution guidelines:
	https://golang.org/doc/contribute.html

Note that the Go project does not use GitHub pull requests, and that
we use the issue tracker for bug reports and proposals only. See
https://golang.org/wiki/Questions for a list of places to ask
questions about the Go language.

[rf]: https://reneefrench.blogspot.com/
[cc3-by]: https://creativecommons.org/licenses/by/3.0/
