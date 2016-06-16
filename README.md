### Moodie Api

Routes:

`GET /`
	Health check URI

`POST /login/facebook`
	login with facebook
 
 required:
 	
 	`token: string`
 	exchange short-lived facebook token for an user object
 	
 resp:
 
 	`{
 		"id": "1",
 		"name": "Paul", 
 		"email": "paul@moodie.ca",
 		"avatarUrl": "xxx",
 		"jwt": "xxxx"
 	 }`
 	 
 notice the jwt returned. **This will be the endpoint** that returns the jwt. If you lose it, you'll have to reauthenticate again. For all authed endpoints, you need to pass in the jwt in the `Authorization` header as `BEARER xxx`
 
`CRUD /posts`

POST:
	
	{
		user_id: 1,
		caption: "this is great",
		imageUrl: "http://someimageurl.jpg",
	}

