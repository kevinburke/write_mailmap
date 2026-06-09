package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const Version = "0.4"

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

type authorMapper func([]string) ([]string, error)

func mailmapAuthors(authors []string) ([]string, error) {
	if len(authors) == 0 {
		return nil, nil
	}
	cmd := exec.Command("git", "check-mailmap", "--stdin")
	cmd.Stdin = strings.NewReader(strings.Join(authors, "\n") + "\n")
	bits, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("git check-mailmap --stdin: %w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("git check-mailmap --stdin: %w", err)
	}

	mappedAuthors := make([]string, 0, len(authors))
	scanner := bufio.NewScanner(bytes.NewReader(bits))
	for scanner.Scan() {
		mappedAuthors = append(mappedAuthors, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(mappedAuthors) != len(authors) {
		return nil, fmt.Errorf("git check-mailmap --stdin returned %d authors for %d input authors", len(mappedAuthors), len(authors))
	}
	return mappedAuthors, nil
}

// parseCoAuthors scans commit message bodies for "Co-Authored-By:" trailers,
// applies mailmap normalization, and returns any new authors not already
// present in seenAuthors. Found authors are added to seenAuthors as a side
// effect.
func parseCoAuthors(body []byte, seenAuthors map[string]bool, mapAuthors authorMapper) ([]string, error) {
	var rawAuthors []string
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Match "Co-Authored-By:" case-insensitively.
		if len(line) < len("Co-Authored-By:") {
			continue
		}
		if !strings.EqualFold(line[:len("Co-Authored-By:")], "Co-Authored-By:") {
			continue
		}
		author := strings.TrimSpace(line[len("Co-Authored-By:"):])
		if author == "" {
			continue
		}
		if _, err := mail.ParseAddress(author); err != nil {
			return nil, fmt.Errorf("invalid Co-Authored-By trailer %q: %w", author, err)
		}
		rawAuthors = append(rawAuthors, author)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	mappedAuthors, err := mapAuthors(rawAuthors)
	if err != nil {
		return nil, err
	}
	if len(mappedAuthors) != len(rawAuthors) {
		return nil, fmt.Errorf("author mapper returned %d authors for %d input authors", len(mappedAuthors), len(rawAuthors))
	}

	var authors []string
	for _, author := range mappedAuthors {
		author = strings.TrimSpace(author)
		if _, err := mail.ParseAddress(author); err != nil {
			return nil, fmt.Errorf("invalid mailmapped Co-Authored-By trailer %q: %w", author, err)
		}
		lowerAuthor := strings.ToLower(author)
		if seenAuthors[lowerAuthor] {
			continue
		}
		authors = append(authors, author)
		seenAuthors[lowerAuthor] = true
	}
	return authors, nil
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

	// Also parse Co-Authored-By trailers from commit messages.
	coAuthorCmd := exec.Command("git", "log", "--use-mailmap", "--format=%B")
	coAuthorBits, err := coAuthorCmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	coAuthors, err := parseCoAuthors(coAuthorBits, seenAuthors, mailmapAuthors)
	if err != nil {
		log.Fatal(err)
	}
	authors = append(authors, coAuthors...)

	sort.Strings(authors)
	if _, err := os.Stdout.WriteString(strings.Join(authors, "\n") + "\n"); err != nil {
		log.Fatal(err)
	}
}
