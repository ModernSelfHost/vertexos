package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "health":
		cmdHealth(os.Args[2:])
	case "auth":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "login":
			cmdLogin(os.Args[3:])
		default:
			usage()
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("vertexctl: health | auth login")
}

func httpClient() *http.Client {
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
}

func apiBase() string {
	if v := os.Getenv("VERTEX_API"); v != "" {
		return v
	}
	return "https://127.0.0.1:8443"
}

func cmdHealth(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	insecure := fs.Bool("k", true, "skip TLS verify")
	fs.Parse(args)
	_ = insecure
	resp, err := httpClient().Get(apiBase() + "/api/v1/health")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())
}

func cmdLogin(args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("u", "", "username")
	pass := fs.String("p", "", "password")
	fs.Parse(args)
	body, _ := json.Marshal(map[string]string{"username": *user, "password": *pass})
	resp, err := httpClient().Post(apiBase()+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())
}
