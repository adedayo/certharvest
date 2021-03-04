package certharvest

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sync"
	"time"
)

//GetServerCertificates gets the certificate chain associated with a URL
//This is non-blocking. Use the blocking version GetServerCertificatesBlocking if suitable
func GetServerCertificates(config Config, urls ...string) <-chan CertificatesOrError {
	certsAndErrs := make([]<-chan CertificatesOrError, 0)
	for _, url := range urls {
		certsAndErrs = append(certsAndErrs, getCert(url, config))
	}
	return mergeChannels(certsAndErrs...)
}

func getCert(url string, config Config) <-chan CertificatesOrError {
	result := make(chan CertificatesOrError)
	go func() {
		defer close(result)
		certs := []*x509.Certificate{}
		var certError error
		client := getClient(&certs, &certError)
		resp, err := client.Get(url)
		if err != nil {
			result <- CertificatesOrError{url, certs, err}
			return
		}
		defer resp.Body.Close()
		if certError != nil {
			result <- CertificatesOrError{url, certs, err}
			return
		}
		result <- CertificatesOrError{url, certs, nil}
		return
	}()
	return result
}

func mergeChannels(channels ...<-chan CertificatesOrError) <-chan CertificatesOrError {
	var wg sync.WaitGroup
	out := make(chan CertificatesOrError)
	output := func(c <-chan CertificatesOrError) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go output(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//GetServerCertificatesBlocking gets the certificate chain associated with a URL
func GetServerCertificatesBlocking(config Config, urls ...string) (out []CertificatesOrError) {
	for o := range GetServerCertificates(config, urls...) {
		out = append(out, o)
	}
	return
}

//CertificatesOrError contains the certificate chain associated with a URL, or error if there was a problem
type CertificatesOrError struct {
	URL              string
	CertificateChain []*x509.Certificate
	Error            error
}

func certificateInterceptor(certs *[]*x509.Certificate, certError *error) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		for _, cc := range rawCerts {

			if cert, err := x509.ParseCertificate(cc); err == nil {
				*certs = append(*certs, cert)
			} else {
				certError = &err
				return err
			}

		}
		return nil
	}
}

func getClient(certs *[]*x509.Certificate, err *error) http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify:    true, //yep this is deliberate :-)
		VerifyPeerCertificate: certificateInterceptor(certs, err),
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, //10 second timeout //TODO parametrise this
	}
}

//Config is the configuration for harvesting certificates
type Config struct {
	//Timeout fir the HTTP client connection
	TimeOut time.Duration
}
