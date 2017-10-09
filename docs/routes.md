# 



## Routes

<details>
<summary>`/`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/**
	- _GET_
		- [New.func2.1]()

</details>
<details>
<summary>`/carts`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/**
		- _POST_
			- [CreateCart]()
		- _GET_
			- [ListCarts]()

</details>
<details>
<summary>`/carts/:cartID/*`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/**
				- _PUT_
					- [UpdateCart]()
				- _DELETE_
					- [ClearCart]()
				- _GET_
					- [GetCart]()

</details>
<details>
<summary>`/carts/:cartID/*/checkout`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/checkout**
				- **/**
					- _POST_
						- [CreateCheckout]()
					- _PUT_
						- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/:cartID/*/items`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/items**
				- **/**
					- _POST_
						- [CreateCartItem]()

</details>
<details>
<summary>`/carts/:cartID/*/items/:cartItemID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/items**
				- **/:cartItemID**
					- [CartItemCtx]()
					- **/**
						- _PUT_
							- [UpdateCartItem]()
						- _DELETE_
							- [RemoveCartItem]()
						- _GET_
							- [GetCartItem]()

</details>
<details>
<summary>`/carts/:cartID/*/items/quantities`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/items**
				- **/quantities**
					- _GET_
						- [CheckQuantities]()

</details>
<details>
<summary>`/carts/:cartID/*/payment`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/payment**
				- **/**
					- _POST_
						- [CreatePayment]()

</details>
<details>
<summary>`/carts/:cartID/*/payment/methods`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/payment**
				- **/methods**
					- _GET_
						- [ListPaymentMethods]()

</details>
<details>
<summary>`/carts/:cartID/*/shipping`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/:cartID**
		- [CartCtx]()
		- **/***
			- **/shipping**
				- **/**
					- _PUT_
						- [UpdateShippingMethod]()
					- _GET_
						- [ListShippingRates]()

</details>
<details>
<summary>`/carts/default/*`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/**
				- _PUT_
					- [UpdateCart]()
				- _DELETE_
					- [ClearCart]()
				- _GET_
					- [GetCart]()

</details>
<details>
<summary>`/carts/default/*/checkout`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/checkout**
				- **/**
					- _POST_
						- [CreateCheckout]()
					- _PUT_
						- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/default/*/items`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/items**
				- **/**
					- _POST_
						- [CreateCartItem]()

</details>
<details>
<summary>`/carts/default/*/items/:cartItemID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/items**
				- **/:cartItemID**
					- [CartItemCtx]()
					- **/**
						- _PUT_
							- [UpdateCartItem]()
						- _DELETE_
							- [RemoveCartItem]()
						- _GET_
							- [GetCartItem]()

</details>
<details>
<summary>`/carts/default/*/items/quantities`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/items**
				- **/quantities**
					- _GET_
						- [CheckQuantities]()

</details>
<details>
<summary>`/carts/default/*/payment`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/payment**
				- **/**
					- _POST_
						- [CreatePayment]()

</details>
<details>
<summary>`/carts/default/*/payment/methods`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/payment**
				- **/methods**
					- _GET_
						- [ListPaymentMethods]()

</details>
<details>
<summary>`/carts/default/*/shipping`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/carts**
	- **/default**
		- [DefaultCartCtx]()
		- **/***
			- **/shipping**
				- **/**
					- _GET_
						- [ListShippingRates]()
					- _PUT_
						- [UpdateShippingMethod]()

</details>
<details>
<summary>`/connect/:shopID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/connect/:shopID**
	- _GET_
		- [Connect]()

</details>
<details>
<summary>`/echo`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/echo**
	- _POST_
		- [echoPush]()

</details>
<details>
<summary>`/leaderboard`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/leaderboard**
	- _GET_
		- [leaderBoard]()

</details>
<details>
<summary>`/locales`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/locales**
	- **/**
		- _GET_
			- [ListLocale]()

</details>
<details>
<summary>`/locales/:localeID/places`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/locales**
	- **/:localeID**
		- [LocaleCtx]()
		- **/places**
			- _GET_
				- [ListPlaces]()

</details>
<details>
<summary>`/login`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/login**
	- _POST_
		- [EmailLogin]()

</details>
<details>
<summary>`/login/facebook`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/login/facebook**
	- _POST_
		- [FacebookLogin]()

</details>
<details>
<summary>`/oauth/shopify/callback`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/oauth/shopify/callback**
	- _GET_
		- [bitbucket.org/moodie-app/moodie-api/lib/connect.(*Shopify).OAuthCb-fm]()

</details>
<details>
<summary>`/places/:placeID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/**
			- _GET_
				- [GetPlace]()

</details>
<details>
<summary>`/places/:placeID/follow`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/follow**
			- _POST_
				- [FollowPlace]()
			- _DELETE_
				- [UnfollowPlace]()

</details>
<details>
<summary>`/places/:placeID/prices`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/prices**
			- _GET_
				- [ListProductPrices]()

</details>
<details>
<summary>`/places/:placeID/products`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/products**
			- _GET_
				- [ListProduct]()

</details>
<details>
<summary>`/places/:placeID/share`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/share**
			- _POST_
				- [Share]()

</details>
<details>
<summary>`/places/:placeID/tags`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/tags**
			- _GET_
				- [ListProductTags]()

</details>
<details>
<summary>`/places/following`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/following**
		- _GET_
			- [ListFollowing]()

</details>
<details>
<summary>`/places/nearby`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/nearby**
		- _GET_
			- [ListNearby]()

</details>
<details>
<summary>`/places/recent`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/recent**
		- _GET_
			- [ListRecent]()

</details>
<details>
<summary>`/products/:productID/variant`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/products**
	- **/:productID**
		- [ProductCtx]()
		- **/variant**
			- _GET_
				- [GetVariant]()

</details>
<details>
<summary>`/register`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/register**
	- _POST_
		- [RegisterSignup]()

</details>
<details>
<summary>`/search`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/search**
	- _POST_
		- [(*JwtAuth).Verify.func1]()
		- [SessionCtx]()
		- [UserRefresh]()
		- [OmniSearch]()

</details>
<details>
<summary>`/session`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/session**
	- **/**
		- _DELETE_
			- [Logout]()

</details>
<details>
<summary>`/session/heartbeat`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/session**
	- **/heartbeat**
		- _POST_
			- [PostHeartbeat]()

</details>
<details>
<summary>`/signup`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/signup**
	- _POST_
		- [EmailSignup]()
	- _GET_
		- [GetSignupPage]()

</details>
<details>
<summary>`/users/me`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/**
			- _GET_
				- [GetUser]()

</details>
<details>
<summary>`/users/me/address`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/address**
			- **/**
				- _POST_
					- [CreateAddress]()
				- _GET_
					- [ListAddresses]()

</details>
<details>
<summary>`/users/me/device`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/device**
			- _PUT_
				- [SetDeviceToken]()

</details>
<details>
<summary>`/users/me/ping`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/ping**
			- _GET_
				- [Ping]()

</details>
<details>
<summary>`/webhooks/shopify`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/webhooks/shopify**
	- _POST_
		- [ShopifyStoreWhCtx]()
		- [WebhookHandler]()

</details>

Total # of routes: 46
