package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/yanatan16/golang-instagram/instagram"
)

var (
	client = &instagram.Api{
		AccessToken: "31483978.4f76e3b.39af1b07617d49ad9b4fd7a65014b93a",
	}
	clientID     = "4f76e3bdf20749c590e707314ccf219b"
	clientSecret = "17cca07c55ee420faf6487601cee3264"
	redirectURI  = "http://localhost:1234/callback"
	authURL      = "https://api.instagram.com/oauth/authorize/?client_id=%s&redirect_uri=%s&response_type=code&scope=basic+public_content+follower_list"
)

func Auth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, authURL, 302)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	args := url.Values{}
	args.Add("client_id", clientID)
	args.Add("client_secret", clientSecret)
	args.Add("redirect_uri", redirectURI)
	args.Add("grant_type", "authorization_code")
	args.Add("code", code)

	resp, err := http.PostForm("https://api.instagram.com/oauth/access_token", args)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(b))
}

func main() {
	fmt.Println("starting instaload")

	//authURL = fmt.Sprintf(authURL, clientID, redirectURI)

	//http.HandleFunc("/auth", Auth)
	//http.HandleFunc("/callback", Callback)
	//http.ListenAndServe(":1234", nil)

	resp, _ := client.GetUserFollows("self", nil)
	for _, u := range resp.UsersResponse.Users {
		fmt.Println(u.Id, u.Username)
	}
}
