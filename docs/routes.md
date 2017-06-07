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
<summary>`/places/:placeID/promos`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/promos**
			- _GET_
				- [ListPromo]()

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
<summary>`/places/:placeID/shopify/sync`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/:placeID**
		- [PlaceCtx]()
		- **/shopify**
			- **/sync**
				- _POST_
					- [CredCtx]()
					- [ClientCtx]()
					- [SyncProduct]()

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
<summary>`/products/:productID/claim`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/products**
	- **/:productID**
		- [ProductCtx]()
		- **/claim**
			- _POST_
				- [ClaimProduct]()

</details>
<details>
<summary>`/promos/:promoID/claims`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/:promoID**
		- [PromoCtx]()
		- **/claims**
			- [ClaimCtx]()
			- **/**
				- _GET_
					- [GetClaims]()
				- _DELETE_
					- [RemoveClaim]()

</details>
<details>
<summary>`/promos/:promoID/claims/complete`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/:promoID**
		- [PromoCtx]()
		- **/claims**
			- [ClaimCtx]()
			- **/complete**
				- _PUT_
					- [CompleteClaim]()

</details>
<details>
<summary>`/search`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/search**
	- **/**
		- _POST_
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

</details>
<details>
<summary>`/users/me/cart`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/cart**
			- _GET_
				- [GetCart]()

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
<summary>`/users/me/nda`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/nda**
			- _POST_
				- [AcceptNDA]()

</details>
<details>
<summary>`/webhooks/shopify`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/webhooks/shopify**
	- _POST_
		- [WebhookHandler]()

</details>

Total # of routes: 27
