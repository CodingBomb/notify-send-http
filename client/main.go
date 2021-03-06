package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

var (
	icon, summary, body, notificationServer = "", "", "", ""
	timeout = -1
)

func main() {
	notificationServer = os.Getenv("NOTIFY_SEND_URL")
	if notificationServer == "" {
		fmt.Println("NOTIFY_SEND_URL is not set!")
		os.Exit(1)
	}

	args := parseFlags()
	argsLength := len(args)

	if argsLength > 2 || argsLength < 1 {
		fmt.Println("Invalid number of options")
		os.Exit(1)
	}

	summary = args[0]
	if argsLength == 2 {
		body = args[1]
	}

	triggerNotification(summary, body)
}

func parseFlags() []string {
	// This is required because the Guard notifier sends custom parameters at the end of the command,
	// so here we need to reorganize things
	args := []string{}
	for index, arg := range os.Args {
		if strings.HasPrefix(arg, "-") {
			args = append([]string{arg, os.Args[index+1]}, args...)
		} else if index > 0 && ! strings.HasPrefix(os.Args[index-1], "-") {
			args = append(args, arg)
		}
	}

	os.Args = append([]string{os.Args[0]}, args...)
	flag.StringVar(&icon, "i", icon, "Path to icon")
	flag.StringVar(&icon, "icon", icon, "Path to icon")
	flag.String("u", "", "")
	flag.String("a", "", "")
	flag.String("c", "", "")
	flag.IntVar(&timeout, "t", timeout, "Time to expire notification (milliseconds)")
	flag.String("h", "", "")
	// TODO: https://github.com/guard/guard/blob/19351271941a3362a47176c6808ddcb4a675e3ad/lib/guard/notifiers/notifysend.rb#L15
	flag.Parse()

	return flag.Args()
}

func triggerNotification(summary, body string) {
	if icon != "" {
		notification := map[string]string{
			"summary": summary,
			"body":    body,
			"timeout": strconv.Itoa(timeout),
		}
		request, err := newfileUploadRequest(notificationServer, notification, "icon", icon)
		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{}
		_, err = client.Do(request)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err := http.PostForm(notificationServer, url.Values{
			"summary": {summary},
			"body":    {body},
			"timeout": {strconv.Itoa(timeout)},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Based on http://matt.aimonetti.net/posts/2013/07/01/golang-multipart-file-upload-example/
// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}
