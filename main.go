package main

import (
	"crypto/tls"
	"encoding/json"
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

func main() {
	config, err := getConf()
	if err != nil {
		log.Fatalf("Unable to retrieve config options: %s", err.Error())
	}

	mappings, err := getMappings(config.mappings)
	if err != nil {
		log.Fatalf("Unable to retrieve mappings: %s", err.Error())
	}

	if config.debug {
		log.Printf("Config: %+v\n", config)
		log.Printf("Mappings: %+v\n", mappings)
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
		log.Println("Starting HTTP server on port :80")
		log.Fatalln(httpServer.ListenAndServe())
	}()

	// HTTPS server
	httpsRouter := httprouter.New()

	for _, filePath := range mappings.FilePaths {
		httpsRouter.GET(filePath.Path, handleFile(filePath.File))
	}
	for _, pushFilePath := range mappings.PushFilePaths {
		httpsRouter.GET(pushFilePath.Path, handlePush(pushFilePath.File, pushFilePath.PushPaths))
	}

	httpsServer := &http.Server{
		ReadTimeout:  config.readTimeout,
		WriteTimeout: config.writeTimeout,
		IdleTimeout:  config.idleTimeout,
		Handler:      httpsRouter,
		Addr:         ":443",
		TLSConfig:    &tls.Config{GetCertificate: certManager.GetCertificate},
	}

	log.Println("Starting HTTPS server on :443")
	log.Fatalln(httpsServer.ListenAndServeTLS("", ""))
}

func handlePush(fileName string, pushPaths []string) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read %s", fileName)
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if debug {
			log.Printf("%s PUSH", fileName)
		}

		if pusher, ok := w.(http.Pusher); ok {
			for _, pushPath := range pushPaths {
				if debug {
					log.Printf("PUSH path %s", pushPath)
				}
				if err := pusher.Push(pushPath, &http.PushOptions{
					Method: "GET",
				}); err != nil {
					log.Printf("Failed to push %s: %v", pushPath, err)
				}
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
	mappings       string
	debug          bool
	cirtificateDir string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
}

func getConf() (config, error) {
	var domain, cirtificateDir, mappings string
	var readTimeout, writeTimeout, idleTimeout time.Duration
	var debug bool
	flag.StringVar(&domain, "domain", "", "REQUIRED: the domain to point to, i.e. www.ccochrane.com")
	flag.StringVar(&mappings, "mappings", "mappings.json", "REQUIRED: the mapping file for endpoints")
	flag.StringVar(&cirtificateDir, "cirt_dir", ".", "the directory to store generated tls certificates")
	flag.DurationVar(&readTimeout, "read_timeout", time.Second, "HTTP read timeout")
	flag.DurationVar(&writeTimeout, "write_timeout", time.Second, "HTTP write timeout")
	flag.DurationVar(&idleTimeout, "idle_timeout", time.Second, "HTTP idle timeout")
	flag.BoolVar(&debug, "debug", false, "turn on debug logging")
	flag.Parse()

	if domain == "" {
		return config{}, errors.New("-domain flag not provided")
	}

	return config{
		domain:         domain,
		mappings:       mappings,
		debug:          debug,
		cirtificateDir: cirtificateDir,
		readTimeout:    readTimeout,
		writeTimeout:   writeTimeout,
		idleTimeout:    idleTimeout,
	}, nil
}

type Mappings struct {
	FilePaths     []FilePath     `json:"file-paths"`
	PushFilePaths []PushFilePath `json:"push-file-paths"`
}

type PushFilePath struct {
	File      string   `json:"file"`
	Path      string   `json:"path"`
	PushPaths []string `json:"push-paths"`
}

type FilePath struct {
	File string `json:"file"`
	Path string `json:"path"`
}

func getMappings(mappingFile string) (Mappings, error) {
	mappingJSON, err := ioutil.ReadFile(mappingFile)
	if err != nil {
		return Mappings{}, fmt.Errorf("Unable to read mappings JSON file %s", mappingFile)
	}

	var mappings Mappings
	return mappings, json.Unmarshal(mappingJSON, &mappings)
}
