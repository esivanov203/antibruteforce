package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esivanov203/antibruteforce/internal/httpapi"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

func main() {
	var host, port, subnet, login, ip string

	rootCmd := &cobra.Command{
		Use:   "cli",
		Short: "Anti brute service admin utility",
		Run: func(cmd *cobra.Command, args []string) {
			url := fmt.Sprintf("http://%s:%s", host, port)
			resp, err := healthTestService(url)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp)
			}
		},
	}

	// глобальные флаги
	rootCmd.PersistentFlags().StringVar(&host, "server", "localhost", "server hostname or IP")
	rootCmd.PersistentFlags().StringVar(&port, "port", "8090", "server port")

	// версия сервиса
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show calendar service version",
		Run:   printVersion,
	})

	// reset bucket
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Show calendar service version",
		Run: func(cmd *cobra.Command, args []string) {
			url := fmt.Sprintf("http://%s:%s/bucket/reset", host, port)
			status, resp, err := doJsonRequest("POST", url, httpapi.ResetBucketRequest{
				Login: login,
				IP:    ip,
			})
			printResult(status, resp, err)
		},
	}

	resetCmd.PersistentFlags().StringVar(&login, "login", "", "login")
	resetCmd.PersistentFlags().StringVar(&ip, "ip", "", "ip")

	// blacklist команда
	blacklistCmd := &cobra.Command{
		Use:   "blacklist",
		Short: "Manage blacklist",
	}

	blacklistCmd.PersistentFlags().StringVar(&subnet, "subnet", "", "subnet")

	blacklistCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add subnet to blacklist",
		Run: func(cmd *cobra.Command, args []string) {
			handleList("POST", "blacklist", host, port, subnet)
		},
	})

	blacklistCmd.AddCommand(&cobra.Command{
		Use:   "remove",
		Short: "Remove subnet from blacklist",
		Run: func(cmd *cobra.Command, args []string) {
			handleList("DELETE", "blacklist", host, port, subnet)
		},
	})

	// whitelist команда
	whitelistCmd := &cobra.Command{
		Use:   "whitelist",
		Short: "Manage whitelist",
	}

	whitelistCmd.PersistentFlags().StringVar(&subnet, "subnet", "", "subnet")

	whitelistCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add subnet to whitelist",
		Run: func(cmd *cobra.Command, args []string) {
			handleList("POST", "whitelist", host, port, subnet)
		},
	})

	whitelistCmd.AddCommand(&cobra.Command{
		Use:   "remove",
		Short: "Remove subnet from whitelist",
		Run: func(cmd *cobra.Command, args []string) {
			handleList("DELETE", "whitelist", host, port, subnet)
		},
	})

	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(blacklistCmd)
	rootCmd.AddCommand(whitelistCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleList(method, listType, host, port, subnet string) {
	var url string
	if method == "POST" {
		url = fmt.Sprintf("http://%s:%s/%s", host, port, listType)
		status, resp, err := doJsonRequest(method, url, httpapi.SubnetRequest{Subnet: subnet})
		printResult(status, resp, err)
		return
	}

	url = fmt.Sprintf("http://%s:%s/%s?subnet=%s", host, port, listType, subnet)
	status, resp, err := doJsonRequest(method, url, struct{}{})
	fmt.Println(method, status, resp, err)
	printResult(status, resp, err)
}

func printResult(status int, resp httpapi.ResponseBody, err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("HTTP status:", status)
	if len(resp.Errors) > 0 {
		fmt.Println("Errors:", resp.Errors)
	} else if status < 400 {
		if resp.OK {
			fmt.Println("OK:", resp.OK)
		} else {
			fmt.Println("OK: true")
		}
	}
}

func doJsonRequest(method, url string, request interface{}) (int, httpapi.ResponseBody, error) {
	var body httpapi.ResponseBody

	var req *http.Request
	var err error
	if request != nil {
		bj, err := json.Marshal(request)
		if err != nil {
			return 0, body, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(bj))
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer([]byte("{}")))
	}

	if req == nil {
		return 0, body, errors.New("request cannot be nil")
	}

	if err != nil {
		return 0, body, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, body, err
	}
	defer resp.Body.Close()

	if method == "POST" {
		err = json.NewDecoder(resp.Body).Decode(&body)
	}

	return resp.StatusCode, body, err
}

func healthTestService(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err

	}

	return string(respBody), nil
}
