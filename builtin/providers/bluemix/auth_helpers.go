/*
* Licensed Materials - Property of IBM
* (C) Copyright IBM Corp. 2017. All Rights Reserved.
* US Government Users Restricted Rights - Use, duplication or
* disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 */

package bluemix

import (
	json "encoding/json"
	"errors"
	"log"

	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	slsession "github.com/softlayer/softlayer-go/session"
)

// DefaultTimeout default timeout if not specified
const defaultTimeout = time.Second * 60
const defaultEndpoint = "https://api.softlayer.com/xmlrpc/v3"
const defaultRegion = "ng"

// Session stores the information required for communication with the SoftLayer
// and Bluemix API
type Session struct {
	// UserName is the name of the Bluemix API user
	UserName string

	// Password is the password associated with the Bluemix API user
	Password string

	// Endpoint is the API endpoint used by the Bluemix API user
	Endpoint string

	IAMEndpoint string
	IAMClientID string
	IAMSecret   string

	// Timeout specifies a time limit for http requests made by this
	// session. Requests that take longer that the specified timeout
	// will result in an error.
	Timeout time.Duration

	// AccessToken is the token secret for token-based authentication
	AccessToken string

	// AccessToken is the token secret for token-based authentication
	RefreshToken string

	// IdentityCookie is used to aquire a SoftLayer token
	IdentityCookie string

	// SoftLayer IMS token
	SoftLayerIMSToken string

	// SoftLayer IMS account number
	SoftLayerAccountNumber string

	// SoftLayer user ID
	SoftLayerUserID int

	// SoftLayerSesssion is the the SoftLayer session used to connect to the SoftLayer API
	SoftLayerSession *slsession.Session

	// Debug controls logging of request details (URI, parameters, etc.)
	Debug bool
}

// NewSession creates and returns a pointer to a new session object.
func NewSession(args ...interface{}) (*Session, error) {

	keys := map[string]int{"username": 0,
		"password":                 1,
		"identity_cookie":          2,
		"region":                   3, // ng, eu-gb, sydney
		"iam_client_id":            4,
		"iam_secret":               5,
		"timeout":                  6,
		"softlayer_username":       7,
		"softlayer_api_key":        8,
		"softlayer_endpoint_url":   9,
		"softlayer_account_number": 10,
		"softlayer_timeout":        11}
	values := []string{"", "", "", "", "", "", "", "", "", "", "", ""}

	for i := 0; i < len(args); i++ {
		values[i] = args[i].(string)
	}

	// Default to the environment variables
	envFallback(&values[keys["username"]], "username")
	envFallback(&values[keys["password"]], "password")
	envFallback(&values[keys["identity_cookie"]], "identity_cookie")

	username := values[keys["username"]]
	password := values[keys["password"]]
	identityCookie := values[keys["identity_cookie"]]

	// username/password or identity cookie needs to be provided
	if (username == "" || password == "") && (identityCookie == "") {
		return nil, errors.New("Either BlueMix username and password or BlueMix identity cookie (BLUEMIX_IDENTITY_COOKIE environment variable) are required")
	}
	envFallback(&values[keys["region"]], "region")
	envFallback(&values[keys["iam_client_id"]], "iam_client_id")
	envFallback(&values[keys["iam_secret"]], "iam_secret")
	envFallback(&values[keys["timeout"]], "timeout")
	envFallback(&values[keys["softlayer_account_number"]], "softlayer_account_number")

	// Bluemix timeout
	timeout := values[keys["timeout"]]
	var timeoutDuration time.Duration
	if timeout != "" {
		timeoutDuration, _ = time.ParseDuration(fmt.Sprintf("%ss", timeout))
	} else {
		timeoutDuration, _ = time.ParseDuration(fmt.Sprintf("%ss", defaultTimeout))
	}

	//TODO validate the input params
	// Mandatory:
	//	username, password, region
	//  and also softlayer_account_number - if softlayer_account_number isn't specified, need to specify
	// softlayer_api_key and softlayer_username

	//
	// Figure out the bluemix domain name
	// Should be something like ng.bluemix.net or stage1.ng.bluemix.net (basically ${region}.bluemix.net)
	//
	bmDomainName := fmt.Sprintf("%s.bluemix.net", values[keys["region"]])

	bluemixSession := &Session{
		UserName:               username,
		Password:               password,
		IdentityCookie:         identityCookie,
		Endpoint:               fmt.Sprintf("https://login.%s/UAALoginServerWAR", bmDomainName),
		SoftLayerAccountNumber: values[keys["softlayer_account_number"]],
		IAMEndpoint:            fmt.Sprintf("https://iam.%s", bmDomainName),
		IAMClientID:            values[keys["iam_client_id"]],
		IAMSecret:              values[keys["iam_secret"]],
		Timeout:                timeoutDuration,
	}

	//	err
	bluemixSession.authenticate()
	endpointURL := values[keys["softlayer_endpoint_url"]]
	if endpointURL == "" {
		endpointURL = defaultEndpoint
	}
	softlayerSession := slsession.New(
		values[keys["softlayer_username"]], // If not specified , these values will default to the string zero value of ""
		values[keys["softlayer_api_key"]],
		endpointURL,
		values[keys["softlayer_timeout"]],
	)
	// if the SoftLayer IMS account is provided, retrieve the IMS token
	if values[keys["softlayer_account_number"]] != "" {

		if identityCookie == "" {
			// creates the identity cookie
			bluemixSession.createIdentityCookie()
		}
		// obtain an IMS token
		err := bluemixSession.createIMSToken()
		if err != nil {
			return bluemixSession, err
		}
		softlayerSession.UserId = bluemixSession.SoftLayerUserID
		softlayerSession.AuthToken = bluemixSession.SoftLayerIMSToken

	}
	bluemixSession.SoftLayerSession = softlayerSession
	return bluemixSession, nil
}

//Authenticate against Bluemix
func (s *Session) authenticate() error {

	// Create body for token request
	bodyAsValues := url.Values{
		"grant_type": {"password"},
		"username":   {s.UserName},
		"password":   {s.Password},
	}

	authURL := fmt.Sprintf("%s/oauth/token", s.Endpoint)

	authHeaders := map[string]string{
		"Authorization": "Basic Y2Y6",
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	type AuthenticationErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}

	type AuthenticationResponse struct {
		AccessToken  string                      `json:"access_token"`
		TokenType    string                      `json:"token_type"`
		RefreshToken string                      `json:"refresh_token"`
		Error        AuthenticationErrorResponse `json:"error"`
	}

	httpClient := &http.Client{}
	httpClient.Timeout = s.Timeout

	req, err := http.NewRequest("POST", authURL, strings.NewReader(bodyAsValues.Encode()))
	if err != nil {
		return fmt.Errorf("Failed issuing hosted service deletion request: %s", err)
	}
	for k, v := range authHeaders {
		req.Header.Add(k, v)
	}
	response, err := httpClient.Do(req)
	//TODO: error handling

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)

	// unmarshall the response
	var jsonResponse AuthenticationResponse
	err = json.Unmarshal(responseBody, &jsonResponse)

	s.AccessToken = jsonResponse.AccessToken
	s.RefreshToken = jsonResponse.RefreshToken
	return err
}

func (s *Session) createIdentityCookie() error {
	bodyAsValues := url.Values{
		"grant_type":    {"password"},
		"username":      {s.UserName},
		"password":      {s.Password},
		"response_type": {"identity_cookie"},
	}

	authURL := fmt.Sprintf("%s/oauth/token", s.Endpoint)

	authHeaders := map[string]string{
		"Authorization": "Basic Y2Y6",
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	type IdentityCookieErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}

	type IdentityCookieResponse struct {
		Expiration     int64                       `json:"expiration"`
		IdentityCookie string                      `json:"identity_cookie"`
		Error          IdentityCookieErrorResponse `json:"error"`
	}

	httpClient := &http.Client{}
	httpClient.Timeout = s.Timeout

	req, err := http.NewRequest("POST", authURL, strings.NewReader(bodyAsValues.Encode()))
	// TODO: error handling
	for k, v := range authHeaders {
		req.Header.Add(k, v)
	}
	response, err := httpClient.Do(req)
	//TODO: error handling

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)

	// unmarshall the response
	var jsonResponse IdentityCookieResponse
	err = json.Unmarshal(responseBody, &jsonResponse)

	s.IdentityCookie = jsonResponse.IdentityCookie
	return err
}

func (s *Session) createIMSToken() error {
	log.Printf("[INFO] Creating the IMS token...")
	bodyAsValues := url.Values{
		"grant_type":    {"urn:ibm:params:oauth:grant-type:identity-cookie"},
		"cookie":        {s.IdentityCookie},
		"ims_account":   {s.SoftLayerAccountNumber},
		"response_type": {"cloud_iam, ims_portal"},
	}

	authURL := fmt.Sprintf("%s/oidc/token", s.IAMEndpoint)

	authHeaders := map[string]string{
		"Authorization": "Basic Y2Y6",
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	type IMSTokenErrorResponse struct {
		Code        string `json:"error"`
		Description string `json:"error_description"`
	}

	type IMSTokenResponse struct {
		IMSToken   string                `json:"ims_token"`
		IMSUserID  int                   `json:"ims_user_id"`
		TokenType  string                `json:"token_type"`
		ExpiresIn  int                   `json:"expires_in"`
		Expiration int64                 `json:"expiration"`
		Error      IMSTokenErrorResponse `json:"error"`
	}

	count := 10
	for count > 0 {
		httpClient := &http.Client{}
		httpClient.Timeout = s.Timeout

		req, _ := http.NewRequest("POST", authURL, strings.NewReader(bodyAsValues.Encode()))
		for k, v := range authHeaders {
			req.Header.Add(k, v)
		}
		req.SetBasicAuth(s.IAMClientID, s.IAMSecret)

		response, err := httpClient.Do(req)
		if err == nil {
			if response.StatusCode != 200 {
				log.Printf("[ERROR] Response Status: %s", response.Status)
				time.Sleep(1000 * time.Millisecond)
				count--
			} else {
				responseBody, err := ioutil.ReadAll(response.Body)
				if err == nil {
					// unmarshall the response
					var jsonResponse IMSTokenResponse
					err = json.Unmarshal(responseBody, &jsonResponse)

					if jsonResponse.IMSToken != "" {
						log.Printf("[INFO] IMS token aquired")
						s.SoftLayerIMSToken = jsonResponse.IMSToken
						s.SoftLayerUserID = jsonResponse.IMSUserID
						response.Body.Close()
						return nil
					}
					// empty IMS token
					log.Printf("[WARNING] Retrying to aquire the IMS token...")
					time.Sleep(1000 * time.Millisecond)
					count--
				} else {
					log.Println("[WARNING] Error occurred while reading the HTTP response body:  ", err)
					time.Sleep(1000 * time.Millisecond)
					count--
				}
			}
		} else {
			log.Println("[WARNING] Error occurred while aquiring the IMS token:  ", err)
			time.Sleep(1000 * time.Millisecond)
			count--
		}
		log.Printf("[WARNING] Retrying to aquire the IMS token...")
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}
	return errors.New("[ERROR] Failed to retrieve the IMS token")
}

// ValueFromEnv will return the value for param from tne environment if it's set, or "" if not set
func ValueFromEnv(paramName string) string {
	var defValue string

	switch paramName {
	case "username":
		// Prioritize BM_USERNAME
		defValue = os.Getenv("BM_USERNAME")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_USERNAME")
		}

	case "password":
		// Prioritize BM_PASSWORD
		defValue = os.Getenv("BM_PASSWORD")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_PASSWORD")
		}

	case "softlayer_username":
		// Prioritize SL_USERNAME
		defValue = os.Getenv("SL_USERNAME")
		if defValue == "" {
			defValue = os.Getenv("SOFTLAYER_USERNAME")
		}

	case "sofltayer_api_key":
		// Prioritize SL_API_KEY
		defValue = os.Getenv("SL_API_KEY")
		if defValue == "" {
			defValue = os.Getenv("SOFTLAYER_API_KEY")
		}

	case "identity_cookie":
		// Prioritize BM_IDENTITY_COOKIE
		defValue = os.Getenv("BM_IDENTITY_COOKIE")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_IDENTITY_COOKIE")
		}

	case "region":
		// Prioritize BM_REGION
		defValue = os.Getenv("BM_REGION")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_REGION")
		}
		if defValue == "" {
			defValue = defaultRegion
		}

	case "timeout":
		// Prioritize BM_TIMEOUT
		defValue = os.Getenv("BM_TIMEOUT")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_TIMEOUT")
		}

	case "iam_client_id":
		// Prioritize BM_IAM_CLIENT_ID
		defValue = os.Getenv("BM_IAM_CLIENT_ID")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_IAM_CLIENT_ID")
		}

	case "iam_secret":
		// Prioritize BM_IAM_SECRET
		defValue = os.Getenv("BM_IAM_SECRET")
		if defValue == "" {
			defValue = os.Getenv("BLUEMIX_IAM_SECRET")
		}

	case "softlayer_account_number":
		// PRIORITIZE SL_ACCOUNT_NUMBER
		defValue = os.Getenv("SL_ACCOUNT_NUMBER")
		if defValue == "" {
			defValue = os.Getenv("SOFTLAYER_ACCOUNT_NUMBER")
		}
	}

	return defValue
}

func envFallback(value *string, paramName string) {
	if *value == "" {
		*value = ValueFromEnv(paramName)
	}
}
