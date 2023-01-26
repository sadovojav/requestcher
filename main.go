package main

import (
	"encoding/json"
	"fmt"
	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	log         *logrus.Logger
	counter     = 1
	green       = color.New(color.FgGreen).SprintFunc()
	requestInfo RequestInfo
)

type RequestInfo struct {
	Counter    int
	DateTime   string
	Method     string
	RequestURI string
	RemoteAddr string
	Headers    map[string]string
	UrlParams  map[string]string
	FormData   map[string]string
	Body       interface{}
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func handler(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)

	if (*r).Method == "OPTIONS" {
		return
	}

	requestInfo = RequestInfo{}

	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	requestInfo.Counter = counter
	requestInfo.DateTime = time.Now().Format("02.01.06 15:4:05")
	requestInfo.Method = r.Method
	requestInfo.RequestURI = r.RequestURI
	requestInfo.RemoteAddr = r.RemoteAddr

	parseHeaders(r)
	parseUrlParams(r)
	parseBody(r)

	printRequestInfo()

	log.WithFields(logrus.Fields{
		"request": requestInfo,
	}).Info()

	counter++

	w.WriteHeader(http.StatusOK)
	
	json.NewEncoder(w).Encode(requestInfo)
}

func parseHeaders(r *http.Request) {
	m := make(map[string]string)
	for name, value := range r.Header {
		m[name] = strings.Join(value, ", ")
	}
	requestInfo.Headers = m
}

func parseUrlParams(r *http.Request) {
	values := r.URL.Query()

	if len(values) != 0 {
		m := make(map[string]string)
		for name, value := range values {
			m[name] = value[0]
		}

		requestInfo.UrlParams = m
	}
}

func parseBody(r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm() err: %v", err)
		return
	}

	contentType := r.Header.Get("Content-type")
	log.Println(contentType)
	if contentType == "application/x-www-form-urlencoded" {
		values := r.PostForm

		if len(values) != 0 {
			m := make(map[string]string)
			for name, value := range values {
				m[name] = value[0]
			}

			requestInfo.FormData = m
		}
	} else if contentType == "application/json" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var f interface{}
		if err := json.Unmarshal(body, &f); err != nil {
			fmt.Println("Can not unmarshal JSON")
		}

		requestInfo.Body = f
	}
}

func printRequestInfo() {
	fmt.Printf("%s: %s\n", green(requestInfo.Counter), requestInfo.DateTime)
	fmt.Printf("%s: %s\n", green("Method"), requestInfo.Method)
	fmt.Printf("%s: %s\n", green("Request URI"), requestInfo.RequestURI)
	fmt.Printf("%s: %s\n", green("Remote Addr"), requestInfo.RemoteAddr)

	if len(requestInfo.Headers) != 0 {
		fmt.Printf("%s:\n", green("Headers"))
		for name, value := range requestInfo.Headers {
			fmt.Printf("%s = %s\n", name, value)
		}
	}

	if len(requestInfo.UrlParams) != 0 {
		fmt.Printf("%s:\n", green("UrlParams"))
		for name, value := range requestInfo.UrlParams {
			fmt.Printf("%s = %s\n", name, value)
		}
	}

	if len(requestInfo.FormData) != 0 {
		fmt.Printf("%s:\n", green("FormData"))
		for name, value := range requestInfo.FormData {
			fmt.Printf("%s = %s\n", name, value)
		}
	}

	if requestInfo.Body != nil {
		fmt.Printf("%s:\n", green("Body"))

		f := colorjson.NewFormatter()
		f.Indent = 4

		s, _ := f.Marshal(requestInfo.Body)
		fmt.Println(string(s))
	}

	fmt.Printf("\n")
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
			if _, err := os.Stat("./logs"); os.IsNotExist(err) {
				os.Mkdir("./logs", os.ModePerm)
			}

			log = logrus.New()
			log.SetFormatter(&logrus.JSONFormatter{})

			runID := time.Now().Format("run-2006-01-02-15-04-05")
			logFile := runID + ".log"
			f, err := os.OpenFile("./logs/"+logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				log.Out = f
			} else {
				log.Info("Failed to log to file, using default stderr")
			}

			defer f.Close()

			http.HandleFunc("/", handler)

			fmt.Printf("%s: %s\n", green("Starting server at port"), port)
			fmt.Printf("%s: %s\n", green("Log file"), logFile)

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
