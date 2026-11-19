package iptrust_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/IP-Trust/iptrust-go"
)

func ExampleClient_LookupIP() {
	client, err := iptrust.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.LookupIP(context.Background(), "9.9.9.9")
	if err != nil {
		log.Fatal(err)
	}

	if res.Location.Valid {
		fmt.Println(res.Location.Country, res.Location.Resolution)
	}
	if res.ASN.Valid {
		fmt.Println(res.ASN.Number, res.ASN.RiskLabel)
		if res.ASN.Risk.Valid {
			fmt.Println("risk score:", res.ASN.Risk.Value)
		}
	}
	if res.VPN.Valid && res.VPN.Detected {
		if res.VPN.Provider.Valid {
			fmt.Println("VPN:", res.VPN.Provider.Value)
		}
		if res.VPN.LastDetected.Valid {
			fmt.Println("last seen", res.VPN.LastDetected.Value.Format("2006-01-02"))
		}
	}
}

func ExampleClient_LookupIPs() {
	client, err := iptrust.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ips := []string{"9.9.9.9", "2001:4860:4860::8888"}
	results, err := client.LookupIPs(context.Background(), ips)
	if err != nil {
		log.Fatal(err)
	}

	for _, ip := range ips {
		res := results[ip]
		if res.Location.Valid {
			fmt.Println(ip, res.Location.Country)
		}
	}
}

func ExampleClient_LookupASN() {
	client, err := iptrust.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.LookupASN(context.Background(), "AS13335")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.ASN.Company, res.IPv4.TotalIPs, "IPv4 addresses in", res.IPv4.TotalRoutes, "routes")

	// IPv6 address counts are too large for an int64, so they are strings.
	if total, ok := new(big.Int).SetString(res.IPv6.TotalIPsStr, 10); ok {
		fmt.Println(total.BitLen(), "bits of IPv6 addresses")
	}

	for _, country := range res.Countries {
		fmt.Println(country.Country, country.IPv4IPs, country.IPv4Routes)
	}
	fmt.Println(res.Anonymization.DetectedVPNIPs, "VPN and", res.Anonymization.DetectedProxyIPs, "proxy addresses")
}

func ExampleError() {
	client, err := iptrust.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.LookupIP(context.Background(), "not-an-ip")

	var apiErr *iptrust.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusBadRequest:
			fmt.Println("invalid IP:", apiErr.Detail)
		case http.StatusTooManyRequests:
			fmt.Println("quota exceeded, back off")
		default:
			fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.ErrorID)
		}
	}
}

func ExampleClient_DownloadDatabase() {
	client, err := iptrust.NewClient("your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	dl, body, err := client.DownloadDatabase(context.Background(), iptrust.DatabaseGeolocation, iptrust.FormatMMDB)
	if err != nil {
		log.Fatal(err)
	}
	defer body.Close()

	f, err := os.Create(dl.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		log.Fatal(err)
	}
}
