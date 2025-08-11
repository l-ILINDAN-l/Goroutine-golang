package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"wget/internal/network"
	"wget/internal/parser"
	"wget/internal/storage"
	"wget/internal/structers"
)

var (
	levelRecursive  int    // -l, --level
	outputDocument  string // -O, --output-document
	directoryPrefix string // -P, --directory-prefix
	mirror          bool   // -m, --mirror
	pageRequisites  bool   // -p, --page-requisites
	timeout         int    // -T, --timeout
)
var rootCmd = &cobra.Command{
	Use:   "wget",
	Short: "simple wget command for downloading websites",
	Long:  "CLI-util for downloading sites with all attached content",
	Run:   runWget,
}

func init() {
	rootCmd.Flags().IntVarP(&levelRecursive, "level", "l", 5, "Recursion depth (0 - infinite)")
	rootCmd.Flags().StringVarP(&outputDocument, "output-document", "O", "", "Save the document to the specified file")
	rootCmd.Flags().StringVarP(&directoryPrefix, "directory-prefix", "P", ".", "Save files to the specified directory")
	rootCmd.Flags().BoolVarP(&mirror, "mirror", "m", false, "download the website to a local machine")
	rootCmd.Flags().BoolVarP(&pageRequisites, "page-requisites", "p", false, "Download all necessary resources while the site is loading (images, CSS, JS)")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "T", 60, "Server connection timeout")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Task struct {
	URL   *url.URL
	Depth int
}

func runWget(_ *cobra.Command, args []string) {
	if err := os.MkdirAll(directoryPrefix, 0755); err != nil {
		log.Fatalf("error creating directory %s: %v", directoryPrefix, err)
	}

	strg := storage.NewStorage(directoryPrefix)
	visitedURLs := structers.NewSet[string]()
	queue := structers.NewQueue[Task]()

	startURL := args[0]
	startURLParsed, err := url.Parse(startURL)
	if err != nil {
		log.Fatalf("error parsing start url: %v", err)
	}
	queue.Enqueue(Task{URL: startURLParsed, Depth: 1})

	for !queue.IsEmpty() {
		currentTask, _ := queue.Dequeue()
		currentURL := currentTask.URL.String()

		if visitedURLs.Contains(currentURL) {
			continue
		}
		visitedURLs.Add(currentURL)

		body, contentType, err := network.Download(currentURL, timeout)
		if err != nil {
			log.Printf("error downloading %s: %v", currentURL, err)
			continue
		}

		localPath := strg.URLToPath(currentTask.URL)

		if pageRequisites && strings.Contains(contentType, "text/html") {
			bodyBytes, err := io.ReadAll(body)
			if err != nil {
				log.Printf("error reading HTML %s: %v", currentTask.URL.Path, err)
				continue
			}
			err = body.Close()
			if err != nil {
				log.Printf("error closing body: %v", err)
			}

			rewrittenBody, err := parser.RewriteLinks(bytes.NewReader(bodyBytes), currentTask.URL, strg)
			if err != nil {
				log.Printf("Ошибка переписывания ссылок для %s: %v", currentURL, err)
				continue
			}

			if err := strg.Save(localPath, rewrittenBody); err != nil {
				log.Printf("error saving %s: %v", currentTask.URL.Path, err)
				continue
			}

			if currentTask.Depth < levelRecursive || levelRecursive == 0 {
				newLinks, err := parser.ExtractLinks(bytes.NewReader(bodyBytes), currentTask.URL)
				if err != nil {
					log.Printf("error pase HTML by URL %s extracting links: %v", currentTask.URL, err)
					continue
				}

				for _, newLink := range newLinks {
					newLinkURL, err := url.Parse(newLink)
					if err != nil {
						log.Printf("error parsing link to URL %s: %v", newLinkURL, newLink)
						continue
					}
					if newLinkURL.Host == startURLParsed.Host {
						queue.Enqueue(Task{URL: newLinkURL, Depth: currentTask.Depth + 1})
					}
				}
			}
		} else {
			if err := strg.Save(localPath, body); err != nil {
				log.Printf("error saving resources by path %s: %v", localPath, err)
			}
			err := body.Close()
			if err != nil {
				log.Printf("error closing body: %v", err)
			}
		}
	}
}
