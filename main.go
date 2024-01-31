package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	debug   = flag.Bool("d", false, "Enable debug logging")
	version = flag.String("version", "version: 0.1", "v")
)

// example:  gqo ENG-665245, UT-1234
func main() {
	flag.Parse()

	envLogPath := os.Getenv("HOME")
	envLog := envLogPath + "/quickopen.log"

	logToFile, err := os.OpenFile(envLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logToFile.Close()

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	logger := slog.New(slog.NewJSONHandler(logToFile, handlerOpts))
	slog.SetDefault(logger)

	if *debug {
		logger.Warn("verbose debug logging is enabled")
	}

	userArgs := flag.Args()
	lArgs := len(userArgs)

	var wg sync.WaitGroup
	respchan := make(chan string, lArgs)

	if lArgs >= 1 {
		for i, arg := range userArgs {
			if *debug {
				logger.Debug(
					"debug",
					slog.Int("RANGE - arg", int(i)),
					slog.String("%T", arg),
					slog.String("msg", arg),
				)
			}
			wg.Add(1)
			go func(i int, arg string) {
				defer wg.Done()
				matchArgs(i, arg, respchan)
				queue := <-respchan
				logger.Info(
					"test",
					slog.Int("id", int(i)),
					slog.String("result", queue),
				)
				openBrowser(queue)
			}(i, arg)
		}
		wg.Wait()
	} else {
		logger.Error("Use gqo <args>")
	}
}

func matchArgs(i int, arg string, res chan string) {
	if *debug {
		fmt.Printf("goroutine[%d] : %s\n", i, arg)
	}

	var (
		kb_url   string = "https://portal.nutanix.com/kb/"
		jira_url string = "https://jira.nutanix.com/browse/"
		jql_url  string = "https://jira.nutanix.com/secure/QuickSearch.jspa?searchString="
	)

	searchItem := strings.ToUpper(arg)
	searchSplit := strings.Split(arg, "-")
	searchPrefix := strings.ToUpper(searchSplit[0])

	if strings.HasPrefix(searchPrefix, "KB") {
		if *debug {
			fmt.Printf("[%d] MATCHES KB : %v\n", i, searchItem)
		}
		url_concat := kb_url + searchSplit[1]
		res <- url_concat
	} else if strings.HasPrefix(searchPrefix, "ENG") || strings.HasPrefix(searchPrefix, "ONCALL") || strings.HasPrefix(searchPrefix, "UT") {
		if *debug {
			fmt.Printf("[%d] MATCHES JIRA : %v\n", i, searchItem)
		}
		url_concat := jira_url + searchItem
		res <- url_concat
	} else {
		if *debug {
			fmt.Printf("[%d] NO MATCH, SEARCHING JIRA : %v\n", i, searchItem)
		}
		jql_text := `text ~ "`
		jql_orderBy := `" ORDER BY created DESC`
		url_concat := jql_url + jql_text + arg + jql_orderBy
		res <- url_concat
	}
}

func openBrowser(url string) {
	if !*debug {
		var err error
		switch runtime.GOOS {
		case "darwin":
			err = exec.Command("open", url).Start()
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		default:
			err = fmt.Errorf("error with runtime.GOOS: %v", err)
		}
	} else {
		fmt.Println("Not opening link due to debug mode")
	}
}
