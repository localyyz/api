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
<summary>`/categories`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/categories**
	- **/**
		- _GET_
			- [ListCategories]()

</details>
<details>
<summary>`/categories/:category`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/categories**
	- **/:category**
		- [CategoryCtx]()
		- **/**
			- _GET_
				- [GetCategory]()

</details>
<details>
<summary>`/categories/:category/places`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/categories**
	- **/:category**
		- [CategoryCtx]()
		- **/places**
			- _GET_
				- [ListPlaces]()

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
			- _DELETE_
				- [UnfollowPlace]()
			- _POST_
				- [FollowPlace]()

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
<summary>`/places/all`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/all**
		- _GET_
			- [ListPlaces]()

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
<summary>`/places/manage`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/manage**
		- **/**
			- _GET_
				- [ListManagable]()

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
			- [Nearby]()

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
			- [Recent]()

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
<summary>`/promos/:promoID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/:promoID**
		- [PromoCtx]()
		- **/**
			- _GET_
				- [GetPromo]()

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
<summary>`/promos/active`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/active**
		- _GET_
			- [ListActive]()

</details>
<details>
<summary>`/promos/history`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/history**
		- _GET_
			- [ListHistory]()

</details>
<details>
<summary>`/promos/manage`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/manage**
		- [PromoManageCtx]()
		- **/**
			- _POST_
				- [CreatePromo]()
			- _GET_
				- [ListManagable]()

</details>
<details>
<summary>`/promos/manage/:promoID`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/manage**
		- [PromoManageCtx]()
		- **/:promoID**
			- [PromoCtx]()
			- **/**
				- _PUT_
					- [UpdatePromo]()
				- _DELETE_
					- [DeletePromo]()

</details>
<details>
<summary>`/promos/manage/preview`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/manage**
		- [PromoManageCtx]()
		- **/preview**
			- _POST_
				- [PreviewPromo]()

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
<summary>`/session/verify`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/session**
	- **/verify**
		- _POST_
			- [VerifySession]()

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

Total # of routes: 39
