/*

Spotify Authorization Using Refreshable User Authorization

** does not fetch the code from /authorization end point, the iOS SDK does that already

*/

package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "net/url"
    "log"
    "strings"
    "io"
    "io/ioutil"
    "os"
)

/////////////  App Specific Constants //////////////////////

// From your Spotify Developer Dashboard
const clientID = "..."
const clientSecret = "..."

// From your app
const redirectURI = "..." // ie. myapp://spotify-login-callback - Used here by Spotify for code validation only, I think

////////////////////////////////////////////////////////////

type GrantType string

const (
  CodeSwap GrantType = "authorization_code"
  RefreshToken GrantType = "refresh_token"
)

// The token swap and token refresh endpoints on Spotify are nearly identical, so using the same function

// 1. get the access code/refresh token
// for the Spotify iOS-SDK, the input to this service is sent url-encoded in body, seems to be oauth2 spec

// 2. send that to Spotify (along with clientid/secret, type (refresh/swap) as url-encoded body

// 3. response from Spotify with access token is JSON encoded body, send pass-thru unaltered back to the client as JSON

func TokenGrant(grantType GrantType, response http.ResponseWriter, request *http.Request, _ httprouter.Params) {

  reqBody, err := ioutil.ReadAll(request.Body)
  check_error(err, false)

  log.Printf("%v", string(reqBody))

  // 1. Parse request for url fields

  reqValues, err := url.ParseQuery(string(reqBody))
  if check_error_message(err, false, "input decoding error") {
    httpRespond(response, request, err)
  }

  // 2. Send Access Code/Refresh to Spotify

  sendValues := url.Values{}
  sendValues.Set("grant_type", string(grantType)) // authorization_code or refresh_token
  sendValues.Add("client_id", clientID)
  sendValues.Add("client_secret", clientSecret)
  sendValues.Add("redirect_uri", redirectURI)
  if grantType == RefreshToken {
    sendValues.Add("refresh_token", reqValues.Get("refresh_token"))
  } else {
    sendValues.Add("code", reqValues.Get("code"))
  }

  spotifyResp, err := http.Post("https://accounts.spotify.com/api/token", "application/x-www-form-urlencoded", strings.NewReader(sendValues.Encode()))
  if check_error_message(err, false, "Spotify api error") {
    httpRespond(response, request, err)
  }

  log.Printf("Spotify %v response %v", grantType, spotifyResp.Status)

  // 3. Pass thru the JSON response from Spotify back to the client

  response.Header().Set("Content-Type", "application/json")

  body, err := ioutil.ReadAll(spotifyResp.Body)
  response.Write(body)

  //response.WriteHeader(http.StatusOK) // auto called above on Write

}

// ------- HTTP End-Point Handlers

func TokenSwap(response http.ResponseWriter, request *http.Request, params httprouter.Params) {
  TokenGrant("authorization_code", response, request, params)
}

func TokenRefresh(response http.ResponseWriter, request *http.Request, params httprouter.Params) {
  TokenGrant("refresh_token", response, request, params)
}

// ------- Main()

// var logWriter *bufio.Writer

// var notify_chan = make(chan *apns2.Notification, 300)
// var resp_chan = make(chan *apns2.Response, 300)

func main() {

  // Command-Line Arguments
  // argv[1] listen port, ie 8000

  log.Printf("Starting up %v with client id %v\n", os.Args[0], clientID)

  if len(os.Args) != 2 {
    log.Println("Arguments Error. [port]")
    os.Exit(1)
  }

  port := os.Args[1]

  // -- Router

  router := httprouter.New()

  router.POST("/token_swap", TokenSwap)
  router.POST("/token_refresh", TokenRefresh)

  // make sure the https key and cert are in the working dir, DER or PEM works
  priv_key := "server.key"
  pub_key := "server.crt"
  err := http.ListenAndServeTLS(":" + port, pub_key, priv_key, router)
  check_error(err, true)

}

// MARK: Utility Functions

func check_error(err error, fatal bool) {
	if err != nil && fatal {
		log.Fatal(err)
	} else if err != nil {
		log.Println(err)
	}
}

func check_error_message(err error, fatal bool, message string) bool {
	if err != nil && fatal {
		log.Fatal(err)
    return true
	} else if err != nil {
		log.Printf("%v: %v\n", message, err)
    return true
	}
  return false
}

func httpRespond(response http.ResponseWriter, request *http.Request, err error) {
  if err != nil {
    response.WriteHeader(http.StatusInternalServerError)
    message := fmt.Sprintf("error: '%v'", err)
    io.WriteString(response, message)
    } else {
      response.WriteHeader(http.StatusCreated)
    }
}
