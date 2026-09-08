# MoiPayWay Go SDK

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

Any documented path:

```go
client.Request("POST", "wallet/details", map[string]any{"wallet_id": "..."}, true)
```

Webhook verification:

```go
ok := moipayway.VerifyWebhook(rawBody, signatureHeader, timestampHeader, apiKey, 300)
```
