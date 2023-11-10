package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const timeout = time.Second * 120
const defaultHttpsPort = 9443
const defaultHttpPort = 9090

type Client struct {
	hostname     string
	clientID     string
	systemNumber string
	username     string
	password     string

	httpClient *http.Client
}

func NewClient(httpClient *http.Client, hostname, clientID, systemNumber, username, password string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout, Transport: tr}
	}
	return &Client{
		hostname:     hostname,
		clientID:     clientID,
		systemNumber: systemNumber,
		username:     username,
		password:     password,
		httpClient:   httpClient,
	}
}

func main() {
	ctx := context.Background()
	client := NewClient(nil, "hcmsbxas01.sap.digitalriver.com", "300", "00", "DRVEZATEST", "Veza123!")
	fmt.Println("Now check if the server is up")
	err := client.RunHelloWorld(ctx)
	if err != nil {
		fmt.Println("Unable to connect with SAP Webserver")
	}

}

func (c *Client) RunHelloWorld(ctx context.Context) error {
	url := "https://127.0.0.1:9443/helloworld"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("Unable to create a request for helloworld.")
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for helloworld.")
		return err
	}
	fmt.Printf("the status code is %d \n", resp.StatusCode)

	return nil
}
