package cart

import (
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
)

func CreateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	newCart := &data.Cart{
		UserID: user.ID,
	}

	//creds, _ := data.DB.ShopifyCred.FindByPlaceID(2311)
	//api := shopify.NewClient(nil, creds.AccessToken)
	//api.BaseURL, _ = url.Parse(creds.ApiURL)

	//cart := &shopify.CheckoutRequest{
	//&shopify.Checkout{
	//Email: user.Email,
	//},
	//}
	//_, resp, err := api.Checkout.Create(r.Context(), cart)
	//lg.Warn(err, resp.Status)

	//b, err := ioutil.ReadAll(resp.Body)
	//lg.Warn(string(b))

	if err := data.DB.Cart.Save(newCart); err != nil {
		return
	}
}

func UpdateCart(w http.ResponseWriter, r *http.Request) {
}

func DeleteCart(w http.ResponseWriter, r *http.Request) {
}
