
# tokensource

[![GoDoc](https://godoc.org/libs.altipla.consulting/tokensource?status.svg)](https://godoc.org/libs.altipla.consulting/tokensource)

Golang OAuth2 Token Source that notifies the refresh operations.


### Basic usage

Basic usage with the `HasChanged()` getter:

```go
package main

import (
  "github.com/altipla-consulting/tokensource"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/slack"
)

func Handler(w http.ResponseWriter, r *http.Request) {
  authConfig := &oauth2.Config{
    ClientID:     "CLIENT_ID_HERE",
    ClientSecret: "CLIENT_SECRET_HERE",
    Scopes:       []string{"SCOPE_FOO", "SCOPE_BAR"},
    Endpoint:     slack.Endpoint,
    RedirectURL:  "https://www.example.com/oauth2/redirect",
  }

  token, err := authConfig.Exchange(r.Context(), r.FormValue("code"))
  if err != nil {
    processError(err)
    return
  }

  storeTokenInDatabase(token)

  // ------------------------------------------------------------------------------
  // Everything under this comment is repeated every time you want to use the token

  ts := tokensource.NewNotify(r.Context(), authConfig, token)
  defer updateToken(ts)

  // use ts.Client(r.Context()) to make the requests
}

func updateToken(ts *tokensource.Notify) {
  if ts.HasChanged() {
    token, err := ts.Token()
    if err != nil {
      processError(err)
      return
    }

    storeTokenInDatabase(token)
  }
}
```


### Notification callback

Notification callback when the token updates:

```go
package main

import (
  "github.com/altipla-consulting/tokensource"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/slack"
)

func Handler(w http.ResponseWriter, r *http.Request) {
  authConfig := &oauth2.Config{
    ClientID:     "CLIENT_ID_HERE",
    ClientSecret: "CLIENT_SECRET_HERE",
    Scopes:       []string{"SCOPE_FOO", "SCOPE_BAR"},
    Endpoint:     slack.Endpoint,
    RedirectURL:  "https://www.example.com/oauth2/redirect",
  }

  token, err := authConfig.Exchange(r.Context(), r.FormValue("code"))
  if err != nil {
    processError(err)
    return
  }
  
  storeTokenInDatabase(token)

  // ------------------------------------------------------------------------------
  // Everything under this comment is repeated every time you want to use the token

  ts := tokensource.NewNotifyHook(r.Context(), authConfig, token, storeTokenInDatabase)

  // use ts.Client(r.Context()) to make the requests
}
```
