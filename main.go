package main

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	oauthConfig *oauth2.Config
)

func init() {
	oauthConfig = &oauth2.Config{
		RedirectURL:  "https://fhirbit.cfapps.io/callback",
		ClientID:     os.Getenv("FITBIT_CLIENT_ID"),
		ClientSecret: os.Getenv("FITBIT_CLIENT_SECRET"),
		Scopes:       []string{"activity", "profile", "weight", "heartrate", "settings", "location", "sleep", "nutrition", "social"},
		Endpoint:     endpoints.Fitbit,
	}
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `<html>
<body>
	<a href="/login">FitBit Log In</a>
</body>
</html>`
	fmt.Fprintf(w, htmlIndex)
}

var (
	// TODO: randomize it
	oauthStateString = "pseudo-random"
)

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Content: %s\n", content)
}

func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := oauthConfig.Exchange(oauth2.NoContext, code)

	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.fitbit.com/1/user/-/profile.json", nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleGoogleLogin)
	http.HandleFunc("/callback", handleOAuth2Callback)
	http.ListenAndServe(":"+port, nil)
}
