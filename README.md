# write_mailmap

Use this program and a .mailmap file to apply fixes to your AUTHORS list.

Use a `.mailmap` file to merge entries in a commit log, for example:

```
Kevin Burke <kev@inburke.com> Kevin Burke <burke@shyp.com>
```

If I have commits under both of those email addresses, they'll get merged into
one, for the purposes of this program's output.

For more, see
https://stacktoheap.com/blog/2013/01/06/using-mailmap-to-fix-authors-list-in-git/.

## Install

If you have a Go development environment, you can get the binary by running the
following:

```bash
go get -u github.com/kevinburke/write_mailmap
```

## Usage

```
write_mailmap > AUTHORS.txt
```

You can put it in a Makefile like this:

```
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

$(WRITE_MAILMAP):
	go get -u github.com/kevinburke/write_mailmap

AUTHORS.txt: $(WRITE_MAILMAP)
	write_mailmap > AUTHORS.txt

authors: AUTHORS.txt
```

Then run `make authors` and you'll get an AUTHORS.txt and your contributors
don't have to worry about how to install it.
