package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func init() {
	flag.Usage = func() {
		os.Stderr.WriteString(`write_mailmap

Runs 'git log' on your codebase, rewriting commit authors using a .mailmap file
if it exists, and deduplicates any authors that are present. The sorted list 
of authors is printed to stdout.
`)
		os.Exit(2)
	}
}

func main() {
	flag.Parse()
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
