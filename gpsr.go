package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var GWS_API_URL = "https://groups.uw.edu/group_sws/v3/group/"
var PWS_API_URL = "https://ws.admin.washington.edu/identity/v2/person/"

// Elements of the groups API response for members.
type GroupsMember struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Groups API response.
type GroupsMembersResponse struct {
	Data []GroupsMember `json:"data"`
}

// Updater defines the structure of objects in the "updaters" array.
type Updater struct {
	Type string `json:"type"`
	Name string `json:"name"`
	ID   string `json:"id"`
}

// Data defines the structure of the "data" object.
type Data struct {
	Updaters []Updater `json:"updaters"`
}

// Response defines the top-level structure of the JSON.
type MemberManagerResponse struct {
	Data Data `json:"data"`
}

func GetClient(CertFile string, KeyFile string) *http.Client {
	// Load the client's key pair (certificate and private key)
	cert, _ := tls.LoadX509KeyPair(CertFile, KeyFile)

	// Create an HTTP client with the custom transport
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:  []tls.Certificate{cert},
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
			DisableKeepAlives: true,
		},
	}

	return client
}

func GetGroupMembers(client *http.Client, GroupName string) []byte {
	API_REQ := fmt.Sprintf("%s%s/member", GWS_API_URL, GroupName)

	// Make the GET request
	resp, err := client.Get(API_REQ)
	if err != nil {
		log.Fatalf("Error Calling Groups API: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Groups API Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Groups API Response: %v\n", err)
		os.Exit(1)
	}

	return body
}

func GetMemberManagers(client *http.Client, GroupName string) []byte {
	API_REQ := fmt.Sprintf("%s%s", GWS_API_URL, GroupName)

	// Make the GET request
	resp, err := client.Get(API_REQ)
	if err != nil {
		log.Fatalf("Error Calling Groups API: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Groups API Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Groups API Response: %v\n", err)
		os.Exit(1)
	}

	return body
}

func GetHomeDepartment(client *http.Client, netID string) string {
	// Make the GET request
	// https://ws.admin.washington.edu/identity/v2/entity/npho/full
	API_REQ := fmt.Sprintf("%s%s%s", PWS_API_URL, netID, "/full")

	// Make the GET request
	resp, err := client.Get(API_REQ)
	if err != nil {
		log.Fatalf("Error Calling Groups API: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("PWS API Status Code: %d\n", resp.StatusCode)
		fmt.Printf("PWS API Response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		return "None"
	}

	// Create a new go-query document from the HTML string
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(body)))

	// Define the CSS selector to find the target element
	selector := "span.EmployeeHomeDepartment"

	// Find the element and get its text
	HomeDepartment := doc.Find(selector).Text()

	return HomeDepartment
}

func main() {
	// Paths to your client certificate and private key files
	CertFile := "/etc/pki/tls/certs/user-reports.hyakm.washington.edu.crt"
	KeyFile := "/etc/pki/tls/private/user-reports.hyakm.washington.edu.key"

	client := GetClient(CertFile, KeyFile)

	var response GroupsMembersResponse
	body := GetGroupMembers(client, "u_hyak_klone")
	json.Unmarshal([]byte(body), &response)

	var GroupNames []string
	for _, member := range response.Data {
		if member.Type == "group" {
			GroupNames = append(GroupNames, member.ID)
		}
	}
	fmt.Printf("Group Names: %v\n", len(GroupNames))

	for i, group := range GroupNames {
		account, _ := strings.CutPrefix(group, "u_hyak_")

		// Get group count.
		body = GetGroupMembers(client, group)
		json.Unmarshal([]byte(body), &response)
		var GroupNetIDs []string
		for _, member := range response.Data {
			if member.Type == "uwnetid" {
				GroupNetIDs = append(GroupNetIDs, member.ID)
			}
		}

		// Member managers processing.
		body = GetMemberManagers(client, group)
		var response MemberManagerResponse
		json.Unmarshal([]byte(body), &response)
		MemberManagerName := "None"
		MemberManagerNetID := "None"
		if len(response.Data.Updaters) > 0 {
			MemberManagerName = response.Data.Updaters[0].Name
			MemberManagerNetID = response.Data.Updaters[0].ID
		}

		// Get departments.
		HomeDepartment := GetHomeDepartment(client, MemberManagerNetID)

		fmt.Printf("%d,%s,%s,%d,%s,%s,%s\n", i+1, group, account, len(GroupNetIDs), MemberManagerName, MemberManagerNetID, HomeDepartment)
	}
}
