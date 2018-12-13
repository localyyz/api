# 



## Routes

<details>
<summary>`/`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/**
	- _GET_
		- [(*Handler).Routes.func2.1]()

</details>
<details>
<summary>`/carts/*/default/*/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/**
				- _GET_
					- [GetCart]()
				- _PUT_
					- [UpdateCart]()
				- _DELETE_
					- [ClearCart]()

</details>
<details>
<summary>`/carts/*/default/*/*/billing`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/billing**
				- _DELETE_
					- [DeleteCartBilling]()

</details>
<details>
<summary>`/carts/*/default/*/*/checkout`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/checkout**
				- _POST_
					- [CreateCheckouts]()

</details>
<details>
<summary>`/carts/*/default/*/*/checkout/{checkoutID}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/checkout/{checkoutID}**
				- _PUT_
					- [CheckoutCtx]()
					- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/*/default/*/*/items/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/items/***
				- **/{cartItemID}/***
					- [CartItemCtx]()
					- **/**
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/pay**
				- _POST_
					- [CreatePayments]()

</details>
<details>
<summary>`/carts/*/default/*/*/shipping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/default/***
		- [DefaultCartCtx]()
		- **/***
			- **/shipping**
				- _DELETE_
					- [DeleteCartShipping]()
				- _GET_
					- [ListShippingRates]()

</details>
<details>
<summary>`/carts/*/express/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/**
			- _GET_
				- [GetCart]()
			- _DELETE_
				- [DeleteCart]()

</details>
<details>
<summary>`/carts/*/express/*/items`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/items**
			- _POST_
				- [CreateCartItem]()

</details>
<details>
<summary>`/carts/*/express/*/pay`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/pay**
			- _POST_
				- [ExpressShopifyClientCtx]()
				- [CreatePayment]()

</details>
<details>
<summary>`/carts/*/express/*/shipping/address`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/shipping/address**
			- _PUT_
				- [ExpressShopifyClientCtx]()
				- [UpdateShippingAddress]()

</details>
<details>
<summary>`/carts/*/express/*/shipping/estimate`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/shipping/estimate**
			- _GET_
				- [ExpressShopifyClientCtx]()
				- [GetShippingRates]()

</details>
<details>
<summary>`/carts/*/express/*/shipping/method`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/express/***
		- [ExpressCartCtx]()
		- **/shipping/method**
			- _PUT_
				- [ExpressShopifyClientCtx]()
				- [UpdateShippingMethod]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
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
<summary>`/carts/*/{cartID}/*/*/billing`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/billing**
				- _DELETE_
					- [DeleteCartBilling]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/checkout`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/checkout**
				- _POST_
					- [CreateCheckouts]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/checkout/{checkoutID}`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/checkout/{checkoutID}**
				- _PUT_
					- [CheckoutCtx]()
					- [UpdateCheckout]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/items/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/pay`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/pay**
				- _POST_
					- [CreatePayments]()

</details>
<details>
<summary>`/carts/*/{cartID}/*/*/shipping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/carts/***
	- **/{cartID}/***
		- [CartCtx]()
		- **/***
			- **/shipping**
				- _GET_
					- [ListShippingRates]()
				- _DELETE_
					- [DeleteCartShipping]()

</details>
<details>
<summary>`/categories/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/**
		- _GET_
			- [FilterSortCtx]()
			- [CategoryRootCtx]()
			- [List]()

</details>
<details>
<summary>`/categories/*/10/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/10/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/10/products/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/11/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/11/products/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/12/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/12/products/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/13/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/13/products/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/merchants`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/merchants**
		- _POST_
			- [ListMerchants]()

</details>
<details>
<summary>`/categories/*/onsale/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/onsale/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/onsale/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListDiscountProducts]()

</details>
<details>
<summary>`/categories/*/styles`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/styles**
		- _POST_
			- [ListStyles]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/**
			- _GET_
				- [GetCategory]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/merchants`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/merchants**
			- _GET_
				- [ListMerchants]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/categories/*/{categoryID}/*/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/categories/***
	- **/{categoryID}/***
		- [CategoryCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/featured`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/featured**
		- _GET_
			- [ListFeaturedCollection]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/**
			- _GET_
				- [GetCollection]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/collections/*/{collectionID}/*/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/collections/***
	- **/{collectionID}/***
		- [CollectionCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/connect`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/connect**
	- _GET_
		- [Connect]()

</details>
<details>
<summary>`/deals/*/activate`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/activate**
		- _POST_
			- [ActivateDeal]()

</details>
<details>
<summary>`/deals/*/active/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/active/***
		- **/**
			- _GET_
				- [StatusCtx.func1]()
				- [ListDeal]()

</details>
<details>
<summary>`/deals/*/active/*/{dealID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/active/***
		- **/{dealID}/***
			- [DealCtx]()
			- **/**
				- _GET_
					- [GetDeal]()

</details>
<details>
<summary>`/deals/*/comingsoon`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/comingsoon**
		- _GET_
			- [ListUpcomingDeal]()

</details>
<details>
<summary>`/deals/*/featured`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/featured**
		- _GET_
			- [ListFeaturedDeal]()

</details>
<details>
<summary>`/deals/*/history`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/history**
		- _GET_
			- [StatusCtx.func1]()
			- [ListDeal]()

</details>
<details>
<summary>`/deals/*/ongoing`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/ongoing**
		- _GET_
			- [ListOngoingDeal]()

</details>
<details>
<summary>`/deals/*/timed`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/timed**
		- _GET_
			- [ListTimedDeal]()

</details>
<details>
<summary>`/deals/*/upcoming`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/upcoming**
		- _GET_
			- [StatusCtx.func1]()
			- [ListDeal]()

</details>
<details>
<summary>`/deals/*/{dealID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/**
			- _GET_
				- [GetDeal]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/deals/*/{dealID}/*/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/deals/***
	- **/{dealID}/***
		- [DealCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/login`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/oauth/shopify/callback**
	- _GET_
		- [bitbucket.org/moodie-app/moodie-api/lib/connect.(*Shopify).OAuthCb-fm]()

</details>
<details>
<summary>`/ping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/ping**
	- _POST_
		- [LogDeviceData]()

</details>
<details>
<summary>`/places/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/**
		- _GET_
			- [List]()

</details>
<details>
<summary>`/places/*/featured`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/featured**
		- _GET_
			- [ListFeatured]()

</details>
<details>
<summary>`/places/*/{placeID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/**
			- _GET_
				- [GetPlace]()

</details>
<details>
<summary>`/places/*/{placeID}/*/favourite`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/favourite**
			- _POST_
				- [AddFavourite]()
			- _DELETE_
				- [DeleteFavourite]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListProducts]()

</details>
<details>
<summary>`/places/*/{placeID}/*/shipping/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/***
	- **/{placeID}/***
		- [PlaceCtx]()
		- **/shipping/***
			- **/**
				- _POST_
					- [SearchShippingZone]()
				- _GET_
					- [ListShippingZone]()

</details>
<details>
<summary>`/places/internal`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/places/internal**
	- _PUT_
		- [SessionCtx]()
		- [UpdateInternal]()

</details>
<details>
<summary>`/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/**
		- _*_
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/brands**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/categories**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/colors**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/favourite/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/favourite/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/favourite/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [ListFavourite]()

</details>
<details>
<summary>`/products/*/feedv3/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/**
			- _GET_
				- [ListFeedV3]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/onsale/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/onsale/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Onsale]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/feedv3/*/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/feedv3/***
		- [FilterSortCtx]()
		- **/products/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListFeedV3Products]()

</details>
<details>
<summary>`/products/*/history`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/history**
		- _GET_
			- [ListHistoryProduct]()

</details>
<details>
<summary>`/products/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/prices**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/sizes**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/stores**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/subcategories**
		- _*_
			- [WithFilterBy.func1]()
			- [ListProducts]()

</details>
<details>
<summary>`/products/*/trend`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/trend**
		- _GET_
			- [ListTrending]()

</details>
<details>
<summary>`/products/*/{productID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/**
			- _GET_
				- [GetProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/collections/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/collections/***
			- **/**
				- _DELETE_
					- [DeleteFromAllCollections]()

</details>
<details>
<summary>`/products/*/{productID}/*/collections/*/{collectionID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/collections/***
			- **/{collectionID}/***
				- [UserCollectionCtx]()
				- **/**
					- _DELETE_
						- [DeleteProductFromCollection]()

</details>
<details>
<summary>`/products/*/{productID}/*/favourite`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/favourite**
			- _POST_
				- [DeviceCtx]()
				- [AddFavouriteProduct]()
			- _DELETE_
				- [DeviceCtx]()
				- [DeleteFavouriteProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/**
				- _*_
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/brands**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/categories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/colors**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/prices**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/sizes**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/stores**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/related/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/related/***
			- [FilterSortHijacksCtx]()
			- **/subcategories**
				- _*_
					- [WithFilterBy.func1]()
					- [ListRelatedProduct]()

</details>
<details>
<summary>`/products/*/{productID}/*/variant`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/products/***
	- [FilterSortHijacksCtx]()
	- **/{productID}/***
		- [ProductCtx]()
		- **/variant**
			- _GET_
				- [GetVariant]()

</details>
<details>
<summary>`/search/*/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/**
			- _*_
				- [Search]()

</details>
<details>
<summary>`/search/*/*/brands`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/brands**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/categories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/categories**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/colors`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/colors**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/prices`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/prices**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/sizes`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/sizes**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/stores`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/stores**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/*/subcategories`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/***
		- [FilterSortHijacksCtx]()
		- **/subcategories**
			- _*_
				- [WithFilterBy.func1]()
				- [Search]()

</details>
<details>
<summary>`/search/*/related`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/related**
		- _POST_
			- [RelatedTags]()

</details>
<details>
<summary>`/search/*/similar`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/search/***
	- **/similar**
		- _POST_
			- [SimilarSearch]()

</details>
<details>
<summary>`/session/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/session/***
	- **/**
		- _DELETE_
			- [Logout]()

</details>
<details>
<summary>`/signup`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/signup**
	- _POST_
		- [EmailSignup]()

</details>
<details>
<summary>`/users/*/collections/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/collections/***
		- **/**
			- _GET_
				- [ListUserCollections]()
			- _POST_
				- [CreateUserCollection]()

</details>
<details>
<summary>`/users/*/collections/*/{collectionID}/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/collections/***
		- **/{collectionID}/***
			- [UserCollectionCtx]()
			- **/**
				- _PUT_
					- [UpdateUserCollection]()
				- _DELETE_
					- [DeleteUserCollection]()
				- _GET_
					- [GetUserCollection]()

</details>
<details>
<summary>`/users/*/collections/*/{collectionID}/*/products/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/collections/***
		- **/{collectionID}/***
			- [UserCollectionCtx]()
			- **/products/***
				- **/**
					- _GET_
						- [GetUserCollectionProducts]()
					- _POST_
						- [CreateProductInCollection]()

</details>
<details>
<summary>`/users/*/me/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/**
			- _GET_
				- [GetUser]()
			- _PUT_
				- [UpdateUser]()

</details>
<details>
<summary>`/users/*/me/*/address/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
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
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/address/***
			- **/{addressID}/***
				- [AddressCtx]()
				- **/**
					- _GET_
						- [GetAddress]()
					- _PUT_
						- [UpdateAddress]()
					- _DELETE_
						- [RemoveAddress]()

</details>
<details>
<summary>`/users/*/me/*/orders/*`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/orders/***
			- **/**
				- _GET_
					- [ListOrders]()

</details>
<details>
<summary>`/users/*/me/*/ping`</summary>

- [RealIP]()
- [NoCache]()
- [RequestID]()
- [RequestLogger.func1]()
- [(*Handler).Routes.func1]()
- [Verifier.func1]()
- [SessionCtx]()
- [UserRefresh]()
- [DeviceCtx]()
- [PaginateCtx]()
- **/users/***
	- **/me/***
		- [MeCtx]()
		- **/ping**
			- _GET_
				- [Ping]()

</details>

Total # of routes: 193
