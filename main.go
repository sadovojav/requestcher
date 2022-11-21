package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	counter = 1
	green   = color.New(color.FgGreen).SprintFunc()
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	fmt.Printf("%s: %s\n", green(counter), time.Now().Format("02.01.06 15:4:05"))
	fmt.Printf("%s: %s\n", green("Method"), r.Method)
	fmt.Printf("%s: %s\n", green("Request URI"), r.RequestURI)
	fmt.Printf("%s: %s\n", green("Remote Addr"), r.RemoteAddr)
	color.Green("Headers:\n")

	for headerName, headerValue := range r.Header {
		fmt.Printf("%s = %s\n", headerName, strings.Join(headerValue, ", "))
	}

	parseUrlParams(r)

	parseBody(r)

	counter++

	fmt.Printf("\n")
}

func parseUrlParams(r *http.Request) {
	values := r.URL.Query()

	if len(values) != 0 {
		color.Green("URL Params:\n")

		for k, v := range values {
			fmt.Printf("%s = %s\n", k, v[0])
		}
	}
}

func parseBody(r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm() err: %v", err)
		return
	}

	contentType := r.Header.Get("Content-type")

	if contentType == "application/x-www-form-urlencoded" {
		values := r.PostForm

		if len(values) != 0 {
			color.Green("Params:\n")

			for k, v := range values {
				fmt.Printf("%s = %s\n", k, v[0])
			}
		}
	} else if contentType == "application/json" {
		body, err := io.ReadAll(r.Body)

		if err != nil {
			panic(err)
		}

		color.Green("Body:\n")

		jsonStr := string(body)

		prettyJson, _ := prettyJsonFormatter(jsonStr)

		fmt.Printf("%s\n", prettyJson)
	}
}

func prettyJsonFormatter(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}

func main() {
	var port string

	app := &cli.App{
		Name:        "requestcher",
		Description: "HTTP request catcher",
		Usage:       "HTTP request catcher",
		Version:     "v1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Value:       "8080",
				Usage:       "Listening port",
				Destination: &port,
				DefaultText: "8080",
			},
		},
		Action: func(*cli.Context) error {
			http.HandleFunc("/", handler)

			fmt.Printf("Starting server at port %s\n", port)

			if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
				log.Fatal(err)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
