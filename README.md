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

Find your target operating system (darwin, windows, linux) and desired bin
directory, and modify the command below as appropriate:

    curl --silent --location --output=/usr/local/bin/write_mailmap https://github.com/kevinburke/write_mailmap/releases/download/0.2/write_mailmap-linux-amd64 && chmod 755 /usr/local/bin/write_mailmap

On Travis, you may want to create `$HOME/bin` and write to that, since
/usr/local/bin isn't writable with their container-based infrastructure.

The latest version is 0.2.

If you have a Go development environment, you can get the binary by running the
following:

```bash
go get -u github.com/kevinburke/write_mailmap
```

## Usage

```
write_mailmap > AUTHORS.txt
```

You can put it in a Makefile like this. Run `make authors` and the dependencies
will automatically download.

```
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

$(WRITE_MAILMAP):
	curl --silent --location --output=$(WRITE_MAILMAP) https://github.com/kevinburke/write_mailmap/releases/download/0.2/write_mailmap-linux-amd64
	chmod 755 $(WRITE_MAILMAP)

force: ;

AUTHORS.txt: force | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt
```

Then run `make authors` and you'll get an AUTHORS.txt and your contributors
don't have to worry about how to install it.
