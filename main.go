package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

const Version = "0.2.1"

var options struct {
	origin              string
	printVersion        bool
	insecure            bool
	readOnly            bool
	numberOfConcurrency int
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "ws URL",
		Short: "websocket tool",
		Run:   root,
	}
	rootCmd.Flags().StringVarP(&options.origin, "origin", "o", "", "websocket origin")
	rootCmd.Flags().BoolVarP(&options.printVersion, "version", "v", false, "print version")
	rootCmd.Flags().BoolVarP(&options.insecure, "insecure", "k", false, "skip ssl certificate check")
	rootCmd.Flags().BoolVarP(&options.readOnly, "readonly", "r", false, "read only")
	rootCmd.Flags().IntVarP(&options.numberOfConcurrency, "number", "n", 1, "number of concurrency")

	rootCmd.Execute()
}

func root(cmd *cobra.Command, args []string) {
	if options.printVersion {
		fmt.Printf("ws v%s\n", Version)
		os.Exit(0)
	}

	if len(args) != 1 {
		cmd.Help()
		os.Exit(1)
	}

	dest, err := url.Parse(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var origin string
	if options.origin != "" {
		origin = options.origin
	} else {
		originURL := *dest
		if dest.Scheme == "wss" {
			originURL.Scheme = "https"
		} else {
			originURL.Scheme = "http"
		}
		origin = originURL.String()
	}

	var wg sync.WaitGroup
	for n := 0; n < options.numberOfConcurrency; n++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			err = connect(dest.String(), origin, &readline.Config{}, options.insecure, options.readOnly)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}(&wg)
	}
	wg.Wait()
}
