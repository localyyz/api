# Localyyz Tool

collection of random scripts to be run against the database and shopify api
`make run-prod` to run against the production database (dangers!)


## Other tools

brew install python (3.*)
make sure pip is in your path, ie `export PATH=/usr/local/Celler/python/<version>/bin:$PATH`

if you run into some error with pip, you might have to upgrade to the newest
major version of it... if you can't install at all, try this script:

`curl https://bootstrap.pypa.io/get-pip.py | python3`

install and set up a virtual python environment

`pip3 install --user virtualenv`

export virtual env into your path with:

`export PATH=$HOME/Library/Python/3.6/bin:$PATH`

now you can run virtualenv and create a local environment

`virtualenv ENV`

and activate it by

`vitualenv ENV/bin/activate`

and now if you run pip and python it will activate a local version of python/pip



### reports

merchant report:

curl -XGET localhost:3331/places/active | in2csv -f json -v > all_merchants.csv
