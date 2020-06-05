package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"gopkg.in/fatih/set.v0"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func clangQuery(source string, query string, args []string) ([]Match, error) {
	allArgs := append([]string{source}, args...)
	proc := exec.Command("clang-query", allArgs...)

	pin, _ := proc.StdinPipe()
	pout, _ := proc.StdoutPipe()
	perr, _ := proc.StderrPipe()

	proc.Start()
	pin.Write([]byte(query + "\n"))
	pin.Close()

	out, _ := ioutil.ReadAll(pout)
	proc.Wait()

	matches := ParseMatches(string(out))

	errorMessage, _ := ioutil.ReadAll(perr)
	if string(errorMessage) != "" {
		log.Fatal(string(errorMessage))
		return matches, errors.New(string(errorMessage))
	}

	return matches, nil
}

func getExtraArgs(args []string) [][]string {
	for k, v := range args {
		if v == "--" {
			return [][]string{args[k+1:]}
		}
	}
	return [][]string{}
}

// Match a clang-query match
type Match struct {
	info string
}

func (m Match) String() string {
	return m.info
}

// ParseMatches parse matches from a clang-query output stream
func ParseMatches(matches string) []Match {
	var results []Match
	lines := strings.Split(matches, "\n")
	if len(lines) < 6 {
		return []Match{}
	}

	var match []string
	activeMatch := false
	for _, line := range lines {
		if len(line) > 5 && line[0:5] == "Match" {
			activeMatch = true
			if len(match) != 0 { // the match on the stack is done
				results = append(results, Match{strings.Join(match, "\n")})
				match = []string{}
			}
		} else if activeMatch && line != "" {
			match = append(match, line)
		}
	}
	if len(match) != 0 {
		// Store the currently processing match.
		// The last line contains the number of matches which we do not care about.
		results = append(results, Match{strings.Join(match[:int(math.Max(0., float64(len(match))))], "\n")})
		match = []string{}
	}

	return results
}

type flags struct {
	Listen    string
	Files     []string
	ClangArgs []string
}

func getflags() flags {

	listen := flag.String("listen", ":3333", "Where to listening for queries")

	flag.Parse()

	args := flag.Args()
	files := []string{}
	clangArgs := []string{}

	for k, v := range args {
		if v == "---" {
			files = args[:k]
			clangArgs = args[k+1:]
			break
		}
		files = args
	}

	flags := &flags{
		Listen:    *listen,
		Files:     files,
		ClangArgs: clangArgs}

	return *flags
}

type clangError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

type response struct {
	Matches []string `json:"matches"`
}

func main() {
	flags := getflags()

	log.Println("Working on files:")
	for _, file := range flags.Files {
		log.Println("    " + file)
	}

	ln, _ := net.Listen("tcp", flags.Listen)
	conn, _ := ln.Accept()

	for {
		query, _ := bufio.NewReader(conn).ReadString('\n')

		log.Printf("Message Received: %s\n", query)

		// result sets
		results := set.New()
		errors := set.New()

		// limited concurrency, http://jmoiron.net/blog/limiting-concurrency-in-go/
		concurrency := runtime.NumCPU()
		sem := make(chan bool, concurrency)

		// spawn workers
		for _, file := range flags.Files {
			sem <- true
			go func(this_file string) {
				defer func() { <-sem }()

				matches, err := clangQuery(this_file, query, flags.ClangArgs)

				if err != nil {
					log.Fatal(err)
				}

				for _, match := range matches {
					results.Add(match.String())
				}

				if err != nil {
					errors.Add(clangError{this_file, err.Error()})
					log.Println(err)
				}

			}(file)
		}
		// join workers
		for i := 0; i < cap(sem); i++ {
			sem <- true
		}

		response := &response{
			Matches: []string{}}

		// accumulate results
		for _, match := range set.StringSlice(results) {
			response.Matches = append(response.Matches, match)
		}

		jsonResponse, _ := json.Marshal(response)

		conn.Write([]byte(string(jsonResponse) + "\n"))

		log.Printf("Found %d matches", results.Size())

		conn.Close()
		conn, _ = ln.Accept()
	}
}
