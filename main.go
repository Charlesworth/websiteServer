package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/acme/autocert"
)

var debug bool

func handleIndexPush() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	file, err := ioutil.ReadFile("charlesworth.github.io/index.html")
	if err != nil {
		log.Fatalf("Unable to read index.html")
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if debug {
			log.Println("index.html PUSH requested")
		}

		if pusher, ok := w.(http.Pusher); ok {
			if err := pusher.Push("/me.jpg", &http.PushOptions{
				Method: "GET",
			}); err != nil {
				log.Printf("Failed to push: %v", err)
			}

			w.Write(file)
			w.(http.Flusher).Flush()
		}
	}
}

func handleFile(fileName string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read %s", fileName)
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if debug {
			log.Printf("%s requested", fileName)
		}
		w.Write(file)
	}
}

type config struct {
	domain         string
	debug          bool
	cirtificateDir string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
}

func getConf() (config, error) {
	var domain, cirtificateDir string
	var readTimeout, writeTimeout, idleTimeout time.Duration
	var debug bool
	flag.StringVar(&domain, "domain", "", "REQUIRED: the domain to point to, i.e. www.ccochrane.com")
	flag.StringVar(&cirtificateDir, "cirt_dir", ".", "the directory to store generated tls certificates")
	flag.DurationVar(&readTimeout, "read_timeout", time.Second*5, "HTTP read timeout")
	flag.DurationVar(&writeTimeout, "write_timeout", time.Second*5, "HTTP write timeout")
	flag.DurationVar(&idleTimeout, "idle_timeout", time.Second*5, "HTTP idle timeout")
	flag.BoolVar(&debug, "debug", false, "turn on debug logging")
	flag.Parse()

	if domain == "" {
		return config{}, errors.New("-domain flag not provided")
	}

	return config{
		domain:         domain,
		debug:          debug,
		cirtificateDir: cirtificateDir,
		readTimeout:    readTimeout,
		writeTimeout:   writeTimeout,
		idleTimeout:    idleTimeout,
	}, nil
}

func main() {
	config, err := getConf()
	if err != nil {
		log.Fatalf("Unable to retrieve config options: %s", err.Error())
	}

	if config.debug {
		fmt.Printf("Config: %+v\n", config)
		debug = true
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.domain),
		Cache:      autocert.DirCache(config.cirtificateDir),
	}

	// HTTP server
	httpServer := &http.Server{
		ReadTimeout:  config.readTimeout,
		WriteTimeout: config.writeTimeout,
		IdleTimeout:  config.idleTimeout,
		Handler:      certManager.HTTPHandler(nil),
		Addr:         ":80",
	}

	go func() {
		fmt.Println("Starting HTTP server on port :80")
		log.Fatalln(httpServer.ListenAndServe())
	}()

	// HTTPS server
	httpsRouter := httprouter.New()
	httpsRouter.GET("/", handleIndexPush())
	httpsRouter.GET("/index.html", handleFile("charlesworth.github.io/index.html"))
	httpsRouter.GET("/me.jpg", handleFile("charlesworth.github.io/me.jpg"))
	httpsRouter.GET("/favicon.png", handleFile("charlesworth.github.io/favicon.png"))
	httpsRouter.GET("/CVCharlesCochrane.pdf", handleFile("charlesworth.github.io/CVCharlesCochrane.pdf"))
	httpsRouter.GET("/keybase.txt", handleFile("charlesworth.github.io/keybase.txt"))
	httpsServer := &http.Server{
		ReadTimeout:  config.readTimeout,
		WriteTimeout: config.writeTimeout,
		IdleTimeout:  config.idleTimeout,
		Handler:      httpsRouter,
		Addr:         ":443",
		TLSConfig:    &tls.Config{GetCertificate: certManager.GetCertificate},
	}

	fmt.Println("Starting HTTPS server on :443")
	log.Fatalln(httpsServer.ListenAndServeTLS("", ""))
}
