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
		- [New.func2]()

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
<summary>`/places/autocomplete`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/places**
	- **/autocomplete**
		- _POST_
			- [AutoComplete]()

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
<summary>`/promos/:promoID/claim`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/:promoID**
		- [PromoCtx]()
		- **/claim**
			- _POST_
				- [ClaimCtx]()
				- [ClaimPromo]()

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
			- _GET_
				- [GetClaims]()

</details>
<details>
<summary>`/promos/:promoID/save`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/promos**
	- **/:promoID**
		- [PromoCtx]()
		- **/save**
			- _DELETE_
				- [ClaimCtx]()
				- [UnSavePromo]()

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
<summary>`/users/:userID/*`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/:userID**
		- [UserCtx]()
		- **/***
			- **/**
				- _GET_
					- [GetUser]()

</details>
<details>
<summary>`/users/:userID/*/device`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/:userID**
		- [UserCtx]()
		- **/***
			- **/device**
				- _PUT_
					- [SetDeviceToken]()

</details>
<details>
<summary>`/users/:userID/*/shoppinglist`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/:userID**
		- [UserCtx]()
		- **/***
			- **/shoppinglist**
				- _GET_
					- [GetShoppingList]()

</details>
<details>
<summary>`/users/me/*`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/***
			- **/**
				- _GET_
					- [GetUser]()

</details>
<details>
<summary>`/users/me/*/device`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/***
			- **/device**
				- _PUT_
					- [SetDeviceToken]()

</details>
<details>
<summary>`/users/me/*/shoppinglist`</summary>

- [NoCache]()
- [Logger]()
- [Recoverer]()
- [New.func1]()
- **/users**
	- **/me**
		- [MeCtx]()
		- **/***
			- **/shoppinglist**
				- _GET_
					- [GetShoppingList]()

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

Total # of routes: 37
