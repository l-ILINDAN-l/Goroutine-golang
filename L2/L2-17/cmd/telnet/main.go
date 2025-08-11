package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var rootCmd = &cobra.Command{
	Short: "Teletype Network is a protocol that allows one computer to control another from a remote location",
	Long: `Usage: telnet [OPTION...] [HOST [PORT]]
Options:
	- --timeout number seconds
A client-server protocol based on the exchange of text data over TCP connections. It allows you to remotely control computers using text input and output.`,
	Args: cobra.ExactArgs(2),
	Run:  runTelnet,
}

var timeout int

func init() {
	rootCmd.Flags().IntVar(&timeout, "timeout", 10, "timeout in seconds")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runTelnet(_ *cobra.Command, args []string) {
	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		log.Fatalf("fail to connect server error: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("fail to close connection: %v", err)
		}
	}(conn)

	done := make(chan struct{})
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil && err != io.EOF {
			log.Printf("error connection in time listen: %v", err)
		}

		close(done)
	}()

	_, err = io.Copy(conn, os.Stdin)
	if err != nil && err != io.EOF {
		log.Printf("error connection in time listen: %v", err)
	}
	<-done

	log.Println("Connection closed.")
}
