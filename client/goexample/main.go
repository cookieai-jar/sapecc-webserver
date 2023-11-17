package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const timeout = time.Second * 120
const defaultHttpsPort = 9443
const defaultHttpPort = 9090

type SapEccServer struct {
	Host         string `json:"host"`
	SystemNumber string `json:"systemNumber"`
	Client       string `json:"client"`
	JcoUser      string `json:"jcoUser"`
	JcoPassword  string `json:"jcoPassword"`

	IsTestingServer bool `json:"isTestingServer"`
}

type SapEccLockUserRequest struct {
	Server   SapEccServer `json:"server"`
	Username string       `json:"username"`
}

type SapEccCreateUserRequest struct {
	Server      SapEccServer      `json:"server"`
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	Firstname   string            `json:"firstname"`
	Lastname    string            `json:"lastname"`
	LicenseType string            `json:"licenseType"`
	Parameters  map[string]string `json:"parameters"`
}

// This is the Role of SAP.
type SapActivityGroup struct {
	Group    string `json:"group"`
	FromDate string `json:"fromDate"` // veza server only support MM/dd/yyyy format
	ToDate   string `json:"toDate"`
}

type SapEccAssignUserGroupRequest struct {
	Server     SapEccServer       `json:"server"`
	Username   string             `json:"username"`
	UserGroups []SapActivityGroup `json:"userGroups"`
}

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
	url := "https://127.0.0.1"
	port := 9443
	client := NewClient(nil, "hcmsbxas01.sap.digitalriver.com", "300", "00", "DRVEZATEST", "Veza123!")
	fmt.Println("Now check if the server is up")
	version, err := client.GetVersion(ctx, url, port)
	if err != nil {
		fmt.Println("Unable to connect with SAP Webserver")
		return
	}
	fmt.Println("The version is " + version + " \n")

	fmt.Println("Now ping the server")
	err = client.Ping(ctx, url, port)
	if err != nil {
		fmt.Println("Unable to Ping SAP Webserver")
		return
	}
	fmt.Println("Server is OK")

	username := "TESTUSER6"
	password := "Veza123!"
	firstname := "FirstnameSix"
	lastname := "John"
	licenseType := "91"
	parameters := map[string]string{"/BA1/F4_EXCH": "Test", "/SPE/IF_QUEUE_LOG": "S"}
	groups := []SapActivityGroup{{Group: "/IPRO/MANAGER"}}

	fmt.Println("Now create user " + username)
	err = client.CreateUser(ctx, url, port, username, password, firstname, lastname, licenseType, parameters)
	if err != nil {
		fmt.Println("Unable to Create User " + username)
		return
	}
	fmt.Println("Create user is OK")

	fmt.Println("Now assign group for user " + username)
	err = client.AssignUserGroups(ctx, url, port, username, groups)
	if err != nil {
		fmt.Println("Unable to Assign user group " + username)
		return
	}
	fmt.Println("Assign user group is OK")

	fmt.Println("Now lock a user " + username)
	err = client.Lock(ctx, url, port, username)
	if err != nil {
		fmt.Println("Unable to Lock User " + username)
		return
	}
	fmt.Println("Lock user is OK")
}

func (c *Client) GetVersion(ctx context.Context, vezaServerUrl string, port int) (string, error) {
	url := fmt.Sprintf("%s:%d/about", vezaServerUrl, port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Println("Unable to create a request for /about. err: " + err.Error())
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for /about. err: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) getSapServer() SapEccServer {
	return SapEccServer{
		Host:         c.hostname,
		SystemNumber: c.systemNumber,
		Client:       c.clientID,
		JcoUser:      c.username,
		JcoPassword:  c.password,

		// IsTesting?
		IsTestingServer: true,
	}
}

func (c *Client) Ping(ctx context.Context, vezaServerUrl string, port int) error {
	url := fmt.Sprintf("%s:%d/ping", vezaServerUrl, port)
	sapServer := c.getSapServer()
	body, err := json.Marshal(sapServer)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("Unable to create a request for /ping. err:" + err.Error())
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for /ping. err: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
	}
	return nil
}

func (c *Client) Lock(ctx context.Context, vezaServerUrl string, port int, username string) error {
	url := fmt.Sprintf("%s:%d/lock", vezaServerUrl, port)
	sapServer := c.getSapServer()
	request := SapEccLockUserRequest{
		Server:   sapServer,
		Username: username,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	fmt.Println("the body is " + string(body))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("Unable to create a request for /lock.")
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for /lock. err: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
	}
	return nil
}

func (c *Client) CreateUser(ctx context.Context, vezaServerUrl string, port int, username, password, firstname, lastname, licenseType string, parameters map[string]string) error {
	url := fmt.Sprintf("%s:%d/create_user", vezaServerUrl, port)
	sapServer := c.getSapServer()
	request := SapEccCreateUserRequest{
		Server:      sapServer,
		Username:    username,
		Password:    password,
		Firstname:   firstname,
		Lastname:    lastname,
		LicenseType: licenseType,
		Parameters:  parameters,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	fmt.Println("the body is " + string(body))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("Unable to create a request for /create_user.")
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for /create_user. err: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
	}
	return nil
}

func (c *Client) AssignUserGroups(ctx context.Context, vezaServerUrl string, port int, username string, groups []SapActivityGroup) error {
	url := fmt.Sprintf("%s:%d/assign_groups", vezaServerUrl, port)
	sapServer := c.getSapServer()
	request := SapEccAssignUserGroupRequest{
		Server:     sapServer,
		Username:   username,
		UserGroups: groups,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}
	fmt.Println("the body is " + string(body))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("Unable to create a request for /assign_groups. err: " + err.Error())
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Unable to get response for /assign_groups.")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("Invalid status code %d", resp.StatusCode))
	}
	return nil
}
