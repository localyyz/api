### Localyyz Api

### Gcloud CLI

# Install

```
brew tap caskroom/cask
brew cask install google-cloud-sdk
```

tip: you can enable auto completion by sourcing the
auto complete file (follow on screen instructions).

# Init and Authenticate

```
gcloud init
```

Build and deploy

1. make sure `docker` is running locally
2. build and deploy with `sup production deploy`

```
gcloud clean up tags
```

gcloud container images delete gcr.io/<project>/api@sha256:<tag> --force-delete-tags

### Troubleshoot

Sup connection error. Check Sup documentation.
if `ssh-add -l` returns `The agent has no identities`
do `ssh-add` and enter the passphrase.


### Refresh search word materialized view

`REFRESH MATERIALIZED VIEW search_words;`
