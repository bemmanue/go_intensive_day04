package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Data struct {
	Money      int
	CandyType  string
	CandyCount int
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {
	candyType := flag.String("k", "", "two-letter abbreviation for the candy type")
	candyCount := flag.Int("c", 0, "count of candy to buy")
	money := flag.Int("m", 0, "amount of money given to machine")
	flag.Parse()

	if !isFlagPassed("k") || !isFlagPassed("c") || !isFlagPassed("m") {
		log.Fatalln("Wrong arguments")
	}

	var data Data
	data.Money = *money
	data.CandyCount = *candyCount
	data.CandyType = *candyType

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalln("Error encoding data: ", err)
	}

	rootCA, err := ioutil.ReadFile("../server-cert/cert.pem")
	if err != nil {
		log.Fatalf("reading cert failed : %v", err)
	}
	rootCAPool := x509.NewCertPool()
	rootCAPool.AppendCertsFromPEM(rootCA)

	client := http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			IdleConnTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAPool,
				GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
					c, err := tls.LoadX509KeyPair("../client-cert/cert.pem", "../client-cert/key.pem")
					if err != nil {
						log.Printf("Error loading key pair: %v\n", err)
						return nil, err
					}
					return &c, nil
				},
			},
		}}

	req, err := http.NewRequest(http.MethodPost, "https://localhost:3333/buy_candy", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalln("Error creating request: ", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln("Error sending request: ", err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
