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

var clientID string // $SPOTIFY_CLIENT_ID
var clientSecret string // $SPOTIFY_CLIENT_SECRET

// the URI user will be redirected to after login
// ie "myapp://spotify_login_callback"
// used by Spotify not just for return view/page, but also code validation i think
var redirectURI string

////////////////////////////////////////////////////////////

type GrantType string

const (
  CodeSwap GrantType = "authorization_code"
  RefreshToken GrantType = "refresh_token"
)

// The token swap and token refresh endpoints on Spotify are nearly identical
// common function for both endpoints

// 1. the input to this service is with the auto code is url-encoded, not JSON, seems to be oauth2 spec
// 2. together with clientid/secret, grant_type (refresh/swap) sent to Spotify as url-encoded body
// 3. response from Spotify is JSON encoded body, send pass-thru unaltered back to the client as JSON

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
    log.Fatal("Arguments Error. [port]")
  }

  port := os.Args[1]

  // -- Env

  clientID = os.Getenv("SPOTIFY_CLIENT_ID")
  clientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
  redirectURI = os.Getenv("SPOTIFY_AUTH_REDIRECT_URI")
  if len(clientID) == 0 || len(clientSecret) == 0 || len(redirectURI) == 0 {
  log.Fatal("Set env SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET and SPOTIFY_AUTH_REDIRECT_URI")
  }

  // -- Router

  router := httprouter.New()

  router.POST("/token_swap", TokenSwap)
  router.POST("/token_refresh", TokenRefresh)

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
