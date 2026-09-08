# MoiPayWay Go SDK

Official Go SDK for the [MoiPayWay API](https://documenter.getpostman.com/view/11919136/2s93Joz7Bq).

Use this client to call every endpoint in that collection: wallets, collections, transfers, users, verification, cards, catalogs, omnichain, and the rest of the documented surface.

```bash
go get github.com/mpwsocial/go
```

```go
package main

import (
	"log"
	"os"

	moipayway "github.com/mpwsocial/go"
)

func main() {
	client, err := moipayway.New(os.Getenv("MOIPAYWAY_API_KEY"), "test")
	if err != nil {
		log.Fatal(err)
	}

	countries, err := client.Countries()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(countries["status"])

	wallet, err := client.CreateWallet(map[string]any{
		"code": "NGN",
		"meta": map[string]any{"name": "Operations", "user_id": "user-uuid"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(wallet)
}
```

## API docs

Paths, methods, and request bodies are in the public collection:

[MoiPayWay API (Postman)](https://documenter.getpostman.com/view/11919136/2s93Joz7Bq)

Call any of those endpoints with `Request()`:

```go
client.Request("POST", "wallet/details", map[string]any{"wallet_id": "..."}, true)
client.Request("GET", "user/misc/countries", nil, false)
```

- `test` → `https://dev.moipayway.co`
- `live` → `https://api.moipayway.co`
- Auth: `Authorization: Bearer <api_key>` (pass `auth=false` for documented catalog GETs)

## Webhooks

```go
ok := moipayway.VerifyWebhook(rawBody, signatureHeader, timestampHeader, apiKey, 300)
```
