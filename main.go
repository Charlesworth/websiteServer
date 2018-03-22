package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("https request")
	io.WriteString(w, `<html><body>Welcome!</body></html>`)
}

type config struct {
	domain         string
	cirtificateDir string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	idleTimeout    time.Duration
}

func getConf() (config, error) {
	var domain, cirtificateDir string
	var readTimeout, writeTimeout, idleTimeout time.Duration
	flag.StringVar(&domain, "domain", "", "REQUIRED: the domain to point to, i.e. www.ccochrane.com")
	flag.StringVar(&cirtificateDir, "cirt_dir", ".", "the directory to store generated tls certificates")
	flag.DurationVar(&readTimeout, "read_timeout", time.Second*5, "HTTP read timeout")
	flag.DurationVar(&writeTimeout, "write_timeout", time.Second*5, "HTTP write timeout")
	flag.DurationVar(&idleTimeout, "idle_timeout", time.Second*5, "HTTP idle timeout")
	flag.Parse()

	if domain == "" {
		return config{}, errors.New("-domain flag not provided")
	}

	return config{
		domain:         domain,
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

	fmt.Printf("Config: %+v\n", config)

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
	httpsMux := &http.ServeMux{}
	httpsMux.HandleFunc("/", handleIndex)
	httpsServer := &http.Server{
		ReadTimeout:  config.readTimeout,
		WriteTimeout: config.writeTimeout,
		IdleTimeout:  config.idleTimeout,
		Handler:      httpsMux,
		Addr:         ":443",
		TLSConfig:    &tls.Config{GetCertificate: certManager.GetCertificate},
	}

	fmt.Println("Starting HTTPS server on :443")
	log.Fatalln(httpsServer.ListenAndServeTLS("", ""))
}
