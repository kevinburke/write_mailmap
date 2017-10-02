package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const Version = "0.2"

func init() {
	flag.Usage = func() {
		os.Stderr.WriteString(`write_mailmap

Runs 'git log' on your codebase, rewriting commit authors using a .mailmap file
if it exists, and deduplicates any authors that are present. The sorted list 
of authors is printed to stdout.

`)
		flag.PrintDefaults()
	}
}

func main() {
	var version = flag.Bool("version", false, "Print the version string and exit")
	flag.Parse()
	if flag.Arg(0) == "version" || *version {
		fmt.Fprintf(os.Stderr, "write_mailmap version %s\n", Version)
		os.Exit(2)
	}
	cmd := exec.Command("git", "log", "--use-mailmap", "--format='%aN <%aE>'")
	bits, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	seenAuthors := make(map[string]bool)
	authors := make([]string, 0)

	scanner := bufio.NewScanner(bytes.NewReader(bits))
	for scanner.Scan() {
		author := strings.Trim(strings.TrimSpace(scanner.Text()), "'")
		lowerAuthor := strings.ToLower(author)
		if _, ok := seenAuthors[lowerAuthor]; ok {
			continue
		} else {
			authors = append(authors, author)
		}
		seenAuthors[lowerAuthor] = true
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(authors)
	if _, err := os.Stdout.WriteString(strings.Join(authors, "\n") + "\n"); err != nil {
		log.Fatal(err)
	}
}
