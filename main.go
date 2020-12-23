package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	importMap map[string]bool
	sumMap    map[string]bool
	module    string

	reModule = regexp.MustCompile(`^module ([-./\w]+)`)

	reCommentStart = regexp.MustCompile(`/\*`)
	reCommentEnd   = regexp.MustCompile(`\*/`)
	reImport       = regexp.MustCompile(`^(import |\s+)?(\w+ )?"([-./\w]+)"`)
	reFinished     = regexp.MustCompile(`^(func |var |const |type )`)
)

func main() {
	importMap = make(map[string]bool)
	sumMap = make(map[string]bool)

	if err := filepath.Walk(".", walker); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var builtin, inModule, other []string
	for i := range importMap {
		switch {
		case !strings.Contains(i, "."):
			builtin = append(builtin, i)
		case module != "" && strings.HasPrefix(i, module):
			inModule = append(inModule, strings.TrimPrefix(i, module+"/"))
		default:
			other = append(other, i)
		}
	}

	var sumImports []string
	if len(sumMap) > 0 {
		for i := range sumMap {
			if !importMap[i] {
				sumImports = append(sumImports, i)
			}
		}
	}

	printList("builtin packages:", builtin)
	printList("\npackages from this module:", inModule)
	printList("\npackages from the internet:", other)
	printList("\npackages imported by other packages:", sumImports)
}

func printList(header string, list []string) {
	if len(list) == 0 {
		return
	}
	sort.Strings(list)
	fmt.Println(header)
	for _, i := range list {
		fmt.Println(i)
	}
}

func parseGoMod(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for s := bufio.NewScanner(file); s.Scan(); {
		txt := s.Text()
		if !reModule.MatchString(txt) {
			continue
		}
		matches := reModule.FindStringSubmatch(txt)
		if len(matches) < 2 {
			return fmt.Errorf("processing %s: module name not found in %s", path, txt)
		}
		module = matches[1]
		return nil
	}
	return fmt.Errorf("module name not found in %s", path)
}

func parseGoSum(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for s := bufio.NewScanner(file); s.Scan(); {
		txt := s.Text()
		split := strings.Split(txt, " ")
		if len(split) != 3 {
			return fmt.Errorf("parsing `%s`: bad format", txt)
		}
		sumMap[split[0]] = true
	}
	return nil
}

func walker(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	switch info.Name() {
	case "go.mod":
		return parseGoMod(path)
	case "go.sum":
		return parseGoSum(path)
	}
	if filepath.Ext(path) != ".go" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	inComment := false
	for s := bufio.NewScanner(file); s.Scan(); {
		txt := s.Text()
		if !inComment && reCommentStart.MatchString(txt) {
			inComment = true
			continue
		}
		if inComment && reCommentEnd.MatchString(txt) {
			inComment = false
			continue
		}
		if inComment {
			continue
		}
		if reFinished.MatchString(txt) {
			break
		}
		if !reImport.MatchString(txt) {
			continue
		}
		matches := reImport.FindStringSubmatch(txt)
		if len(matches) != 4 {
			return fmt.Errorf("processing %s: parsing `%s`: expected import", path, txt)
		}
		importMap[reImport.FindStringSubmatch(txt)[3]] = true
	}
	return nil
}
