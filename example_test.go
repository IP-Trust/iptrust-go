package iptrust_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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
