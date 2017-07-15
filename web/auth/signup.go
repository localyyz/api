package auth

import (
	"io/ioutil"
	"net/http"

	"github.com/pressly/chi/render"
)

func GetSignupPage(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadFile("./web/auth/signup.html")
	render.HTML(w, r, string(b))
}
