package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// define Debug flag as global variable
var Debug = flag.Bool("d", false, "Enable debug mode")

func main() {
	flag.Parse()

	// Initialise new logger and set the default log to slog
	Logger := NewLogger()
	slog.SetDefault(Logger)

	// Get user args, len of args
	userArgs := flag.Args()
	userArgsLen := len(userArgs)

	// Response channel for go routine responses
	respchan := make(chan string, userArgsLen)

	// If the user input has 1 or more arg(s), rangeUserArgsHandler() is called
	// If userArgsLen is 0, getUserInput() is called
	if userArgsLen >= 1 {
		if *Debug {
			slog.Debug("User args greater than 1, rangeUserArgs() called")
		}
		rangeUserArgsHandler(userArgs, respchan)
	} else {
		if *Debug {
			slog.Debug("No args provided, getUserInput() called")
		}
		consoleArgs := getUserInput()
		rangeUserArgsHandler(consoleArgs, respchan)
	}

	//
	slog.Info("Finished?")
}

// Get user input from console, maximum of one line and break with <CR>, split the string at " " and append to a new slice
func getUserInput() []string {
	var consoleArgs []string

	fmt.Print("Enter search term: ")
	reader := bufio.NewReader(os.Stdin)
	consoleInput, err := reader.ReadString('\r')
	for {
		if err != nil {
			fmt.Println("Error reading input:", err)
		}
		if strings.HasSuffix(consoleInput, "\r") || strings.HasSuffix(consoleInput, "\n") {
			strings.TrimSuffix(consoleInput, "\r")
			strings.TrimSuffix(consoleInput, "\n")
			consoleArgs = append(consoleArgs, consoleInput)
			break
		}
		continue
	}

	if *Debug {
		fmt.Println("Items entered:", consoleArgs)
	}

	return consoleArgs
}

type DebugRangeOverInfo struct {
	i          int
	searchItem string
}

func (d DebugRangeOverInfo) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("i", d.i),
		slog.String("searchItem", d.searchItem))
}

// rangeUserArgsHandler(userArgs, respchan)
func (d DebugRangeOverInfo) rangeUserArgsHandler(userArgs []string, msg chan string) {
	var wg sync.WaitGroup
	// make a var queue to store the return channel msg
	var queue string

	// slog.Debug("rangeUserArgs", slog.String("userArgs", strings.Join(userArgs, " ")))
	for i, arg := range userArgs {
		debugMatchUserArgsHandler := DebugRangeOverInfo{i, arg}
		slog.Debug("rangeUserArgs", "debug", debugMatchUserArgsHandler)
		wg.Add(1)
		go func(i int, arg string) {
			defer wg.Done()
			matchUserArgsHandler(i, arg, msg)
		}(i, arg)
		// print the output of the returned goroutine values
		queue = <-msg
		openBrowser(queue)
		// slog.Debug(queue)
		wg.Wait()
	}
}

func matchUserArgsHandler(i int, arg string, res chan string) {
	debugMatchUserArgsHandler := DebugRangeOverInfo{i, arg}
	var (
		kb_url   string = "https://portal.nutanix.com/kb/"
		jira_url string = "https://jira.nutanix.com/browse/"
		jql_url  string = "https://jira.nutanix.com/secure/QuickSearch.jspa?searchString="
	)

	// Check if the user arg "begins with"
	prefix := func(matchAgainst, userMatch string) bool {
		return strings.HasPrefix(matchAgainst, userMatch)
	}

	searchSplit := strings.Split(arg, "-")
	searchItem := strings.ToUpper(arg)

	// check if split searchSplit is all numbers
	re, err := regexp.Compile(`\d+`)

	if re {
		if prefix(searchItem, "KB") {
			slog.Debug("Matches KB", "debug", debugMatchUserArgsHandler)
			url_concat := kb_url + searchSplit[1]
			res <- url_concat
		} else if prefix(searchItem, "ENG") || prefix(searchItem, "ONCALL") || prefix(searchItem, "TH") || prefix(searchItem, "UT") {
			slog.Debug("Matches JIRA", "debug", debugMatchUserArgsHandler)
			url_concat := jira_url + searchItem
			res <- url_concat
		}
	} else {
		var url_concat string = jql_url + `text ~ "` + arg + "\""
		res <- url_concat
		slog.Debug("No match, searching JIRA", "debug", debugMatchUserArgsHandler)
	}
}

// openBrowser
//   - Open the browser (if debug is not enabled)
func openBrowser(url string) {
	var err error
	if !*Debug {
		switch runtime.GOOS {
		case "darwin":
			err := exec.Command("open", url).Start()
			if err != nil {
				log.Fatal(err)
			}
		case "linux":
			err := exec.Command("xdg-open", url).Start()
			if err != nil {
				log.Fatal(err)
			}
		case "windows": // windows lol?
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			if err != nil {
				log.Fatal(err)
			}
		default:
			err = fmt.Errorf("error with runtime.GOOS: %v", err)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		slog.Debug("Not opening browser in debug mode")
	}
}
