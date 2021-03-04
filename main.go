package main

import (
	"fmt"
	"time"

	harvest "github.com/adedayo/certharvest/pkg"
)

//Some examples demonstrating the library usage
func main() {

	//Asynchronous
	for cerr := range harvest.GetServerCertificates(harvest.Config{TimeOut: 5 * time.Second}, "https://google.com", "https://microsoft.com") {
		if cerr.Error == nil {
			fmt.Printf("%s: %#v\n", cerr.URL, cerr)
		} else {
			fmt.Printf("Error: %s\n", cerr.Error.Error())
		}
	}

	//Synchronous
	for _, cerr := range harvest.GetServerCertificatesBlocking(harvest.Config{TimeOut: 5 * time.Second}, "https://google.com", "https://microsoft.com") {
		if cerr.Error == nil {
			fmt.Printf("%s: %#v\n", cerr.URL, cerr)
		} else {
			fmt.Printf("Error: %s", cerr.Error.Error())
			println(cerr.Error.Error())
		}
	}
}
