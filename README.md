# write_mailmap

Use this program and a .mailmap file to apply fixes to your AUTHORS list.

For more, see
https://stacktoheap.com/blog/2013/01/06/using-mailmap-to-fix-authors-list-in-git/.

## Install

go get github.com/kevinburke/write_mailmap

## Usage

write_mailmap > AUTHORS.txt

You can put it in a Makefile like this:

```
WRITE_MAILMAP := $(shell command -v write_mailmap)

authors:
ifndef WRITE_MAILMAP
	go get github.com/kevinburke/write_mailmap
endif
	write_mailmap > AUTHORS.txt
```

Then run `make authors` and you'll get an AUTHORS.txt and your contributors
don't have to worry about how to install it.
