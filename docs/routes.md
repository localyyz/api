# 



## Routes

<details>
<summary>`/`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/**
	- _GET_
		- [(*Handler).Routes.func2.1]()

</details>
<details>
<summary>`/carts/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/**
		- _GET_
			- [ListCarts]()
		- _POST_
			- [CreateCart]()

</details>
<details>
<summary>`/carts/*/default/*/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/**
				- _GET_
					- [GetCart]()
				- _DELETE_
					- [ClearCart]()

</details>
<details>
<summary>`/carts/*/default/*/*/checkout/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/checkout/***
				- **/**
					- _POST_
						- [CreateCheckout]()
					- _PUT_
						- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/*/default/*/*/items/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/items/***
				- **/**
					- _POST_
						- [CreateCartItem]()

</details>
<details>
<summary>`/carts/*/default/*/*/items/*/count`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/items/***
				- **/count**
					- _GET_
						- [CountCartItem]()

</details>
<details>
<summary>`/carts/*/default/*/*/items/*/{cartItemID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/items/***
				- **/{cartItemID}/***
					- [CartItemCtx]()
					- **/**
						- _GET_
							- [GetCartItem]()
						- _PUT_
							- [UpdateCartItem]()
						- _DELETE_
							- [RemoveCartItem]()

</details>
<details>
<summary>`/carts/*/default/*/*/pay`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/pay**
				- _POST_
					- [CreatePayment]()

</details>
<details>
<summary>`/carts/*/default/*/*/shipping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/shipping**
				- _GET_
					- [ListShippingRates]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/**
				- _DELETE_
					- [ClearCart]()
				- _GET_
					- [GetCart]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/checkout/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/checkout/***
				- **/**
					- _POST_
						- [CreateCheckout]()
					- _PUT_
						- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/items/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/items/***
				- **/**
					- _POST_
						- [CreateCartItem]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/items/*/count`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/items/***
				- **/count**
					- _GET_
						- [CountCartItem]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/items/*/{cartItemID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/items/***
				- **/{cartItemID}/***
					- [CartItemCtx]()
					- **/**
						- _DELETE_
							- [RemoveCartItem]()
						- _GET_
							- [GetCartItem]()
						- _PUT_
							- [UpdateCartItem]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/pay`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/pay**
				- _POST_
					- [CreatePayment]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/shipping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/shipping**
				- _GET_
					- [ListShippingRates]()

</details>
<details>
<summary>`/categories/*/gender/{gender}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/categories/***
	- **/gender/{gender}**
		- _GET_
			- [GenderCtx]()
			- [ListCategory]()

</details>
<details>
<summary>`/collections/*/featured`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/collections/***
	- **/featured**
		- _GET_
			- [FeaturedScopeCtx]()
			- [ListCollection]()

</details>
<details>
<summary>`/collections/*/man`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/collections/***
	- **/man**
		- _GET_
			- [MaleScopeCtx]()
			- [ListCollection]()

</details>
<details>
<summary>`/collections/*/woman`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/collections/***
	- **/woman**
		- _GET_
			- [FemaleScopeCtx]()
			- [ListCollection]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/**
			- _GET_
				- [GetCollection]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products**
			- _GET_
				- [ListProduct]()

</details>
<details>
<summary>`/connect`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/connect**
	- _GET_
		- [Connect]()

</details>
<details>
<summary>`/echo`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/echo**
	- _POST_
		- [echoPush]()

</details>
<details>
<summary>`/leaderboard`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/leaderboard**
	- _GET_
		- [leaderBoard]()

</details>
<details>
<summary>`/locales/*/cities`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/locales/***
	- **/cities**
		- _GET_
			- [ListCities]()

</details>
<details>
<summary>`/locales/*/{localeID}/*/places`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/locales/***
	- **/{localeID}/***
		- [LocaleCtx]()
		- **/places**
			- _GET_
				- [ListPlaces]()

</details>
<details>
<summary>`/login`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/login**
	- _POST_
		- [EmailLogin]()

</details>
<details>
<summary>`/login/facebook`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/login/facebook**
	- _POST_
		- [FacebookLogin]()

</details>
<details>
<summary>`/oauth/shopify/callback`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/oauth/shopify/callback**
	- _GET_
		- [bitbucket.org/moodie-app/moodie-api/lib/connect.(*Shopify).OAuthCb-fm]()

</details>
<details>
<summary>`/places/*/approval`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/approval**
		- _POST_
			- [HandleApproval]()

</details>
<details>
<summary>`/places/*/following`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/following**
		- _GET_
			- [ListFollowing]()

</details>
<details>
<summary>`/places/*/{placeID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/**
			- _GET_
				- [GetPlace]()

</details>
<details>
<summary>`/places/*/{placeID}/*/categories/*/gender/{gender}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/categories/***
			- **/gender/{gender}**
				- _GET_
					- [GenderCtx]()
					- [ListCategory]()

</details>
<details>
<summary>`/places/*/{placeID}/*/follow`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/follow**
			- _POST_
				- [FollowPlace]()
			- _DELETE_
				- [UnfollowPlace]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products**
			- _GET_
				- [ListProduct]()

</details>
<details>
<summary>`/places/*/{placeID}/*/share`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/share**
			- _POST_
				- [Share]()

</details>
<details>
<summary>`/products/*/featured`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/featured**
		- _GET_
			- [ListFeaturedProduct]()

</details>
<details>
<summary>`/products/*/gender/{gender}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/gender/{gender}**
		- _GET_
			- [GenderCtx]()
			- [CategoryCtx]()
			- [ListGenderProduct]()

</details>
<details>
<summary>`/products/*/recent`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/recent**
		- _GET_
			- [ListRecentProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/{productID}/***
		- [ProductCtx]()
		- **/**
			- _GET_
				- [GetProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/{productID}/***
		- [ProductCtx]()
		- **/related**
			- _GET_
				- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/variant`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/products/***
	- **/{productID}/***
		- [ProductCtx]()
		- **/variant**
			- _GET_
				- [GetVariant]()

</details>
<details>
<summary>`/register`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/register**
	- _POST_
		- [RegisterSignup]()

</details>
<details>
<summary>`/search/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/search/***
	- **/**
		- _POST_
			- [OmniSearch]()

</details>
<details>
<summary>`/search/*/city/{locale}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/search/***
	- **/city/{locale}**
		- _POST_
			- [LocaleShorthandCtx]()
			- [SearchCity]()

</details>
<details>
<summary>`/session/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/session/***
	- **/**
		- _DELETE_
			- [Logout]()

</details>
<details>
<summary>`/session/*/heartbeat`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/session/***
	- **/heartbeat**
		- _POST_
			- [PostHeartbeat]()

</details>
<details>
<summary>`/signup`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/signup**
	- _GET_
		- [GetSignupPage]()
	- _POST_
		- [EmailSignup]()

</details>
<details>
<summary>`/users/*/me/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/**
			- _GET_
				- [GetUser]()

</details>
<details>
<summary>`/users/*/me/*/address/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/address/***
			- **/**
				- _POST_
					- [CreateAddress]()
				- _GET_
					- [ListAddresses]()

</details>
<details>
<summary>`/users/*/me/*/address/*/{addressID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/address/***
			- **/{addressID}/***
				- [AddressCtx]()
				- **/**
					- _PUT_
						- [UpdateAddress]()
					- _DELETE_
						- [RemoveAddress]()
					- _GET_
						- [GetAddress]()

</details>
<details>
<summary>`/users/*/me/*/device`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/device**
			- _PUT_
				- [SetDeviceToken]()

</details>
<details>
<summary>`/users/*/me/*/ping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/ping**
			- _GET_
				- [Ping]()

</details>
<details>
<summary>`/webhooks/shopify`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Recoverer]()
- [(*JwtAuth).Verify.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [PaginateCtx]()
- **/webhooks/shopify**
	- _POST_
		- [ShopifyStoreWhCtx]()
		- [WebhookHandler]()

</details>

Total # of routes: 55
