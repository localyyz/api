BlogTO Scraper
===

###Neighbourhoods:

You can get a list of neighbourhoods and their geolocations from
`http://www.blogto.com/neighbourhoods/`.

In the html source, find `var neighbourhoods`.
In developer tools, `copy(neighbourhoods)` and paste into
`./data/locale.json`

###Listing stores in a Neighbourhood
`GET http://www.blogto.com/api/v2/listings/`

with url parameters:

    bundle_type=large
    default_neighborhood=<neighbourhood_id>
    limit=<n>
    offset=<n>
    ordering=smart_review
    type=<type_id>

neighbourhood id:

    get from previous step

listing types:

    Restaurants: 1
    Bars: 2
    Cafes: 3
    Design: 4
    Fashion: 5
    Grocery: 6
    Galleries: 7
    Bookstores: 8
    Bakeries: 9
    Fitness: 10
    Hotels: 11
    Services: 12

sample response:

    {
        "count": 0,
        "next": "",
        "previous": "",
        "results": []
    }

same response item:

    {
        "id": 11897,
        "name": "Livsstil",
        "share_url": "http://www.blogto.com/fashion/livsstil-toronto/",
        "image_url": "http://media.blogto.com/listings/3b83-20160715-livsstil.jpg",
        "date_published": "2016-07-15T00:00:00",
        "address": " 445 Adelaide St W",
        "type": {
            "id": 5,
            "name": "Fashion",
            "share_url": "http://www.blogto.com/fashion/"
        },
        "sub_type": {
            "id": 61,
            "name": "Men's and Women's Clothing",
            "share_url": "http://www.blogto.com/fashion/c/toronto/mens-and-womens-clothing/"
        },
        "default_neighborhood": {
            "id": 31,
            "name": "King West",
            "path": "/kingwest/"
        },
        "coordinates": {
            "latitude": "43.64585",
            "longitude": "-79.39899"
        },
        "appears_in_best_of_lists": false,
        "phone": "416.703.2916",
        "rating": 0,
        "dinesafe_establishment_id": "",
        "website": "http://www.livsstilshop.com/"
    }


