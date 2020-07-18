package azuretoken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Tokenresponse struct {
	Tokentype    string `json:"token_type"`
	Expiresin    string `json:"expires_in"`
	Extexpiresin string `json:"ext_expires_in"`
	Expireson    string `json:"expires_on"`
	Notbefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	Accesstoken  string `json:"access_token"`
}

type GraphClient struct {
	TenantID      string
	ApplicationID string
	ClientSecret  string
}

type GraphToken struct {
	AccessToken string
	TokenType   string
	Resource    string
	ExpiresOn   time.Time
	NotBefore   time.Time
}

type Token struct{}

// GetToken implements a function that gets an access token from the Microsoft API and returns it
func (t Token) GetToken(r *strings.Reader, url string) string {
	req, err := http.NewRequest("POST", url, r)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	if err != nil {
		log.Println(err)
	}

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string([]byte(body)))
	if err != nil {
		log.Println(err)
	}

	var token Tokenresponse
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Println(err)
	}
	//fmt.Println(token.Accesstoken)
	return token.Accesstoken

}

// GetGraphToken implements a function that gets an access token from the Microsoft Graph API and returns it
func (t Token) GetGraphToken(graphClient GraphClient) GraphToken {
	resource := fmt.Sprintf("/%v/oauth2/token", graphClient.TenantID)
	data := url.Values{}

	data.Add("grant_type", "client_credentials")
	data.Add("client_id", graphClient.ApplicationID)
	data.Add("client_secret", graphClient.ClientSecret)
	data.Add("resource", "https://graph.microsoft.com")

	u, err := url.ParseRequestURI("https://login.microsoftonline.com")
	if err != nil {
		log.Println(err)
	}

	u.Path = resource
	req, err := http.NewRequest("POST", u.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	httpClient := &http.Client{Timeout: time.Second * 10}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	tmpToken := struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		ExpiresOn    int64  `json:"expires_on"`
		ExtExpiresIn int64  `json:"ext_expires_in"`
		NotBefore    int64  `json:"not_before"`
		Resource     string `json:"resource"`
		TokenType    string `json:"token_type"`
	}{}

	json.Unmarshal(body, &tmpToken)
	return (GraphToken{AccessToken: tmpToken.AccessToken, TokenType: tmpToken.TokenType, Resource: tmpToken.Resource, ExpiresOn: time.Unix(tmpToken.ExpiresOn, 0), NotBefore: time.Unix(tmpToken.NotBefore, 0)})
}
