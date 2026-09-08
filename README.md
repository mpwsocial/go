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

	wallet, err := client.Wallet.Create(map[string]any{
		"code": "NGN",
		"meta": map[string]any{"name": "Operations", "user_id": "user-uuid"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(wallet)
}
```

Collection paths are methods on the client:

- `POST wallet/create` → `client.Wallet.Create(body)`
- `POST wallet/collection/initiate` → `client.Wallet.Collection.Initiate(body)`
- `POST user/account/individual/create` → `client.User.Account.Individual.Create(body)`
- `GET user/misc/job-types` → `client.User.Misc.JobTypes(nil)`
- `POST verification/lookup` → `client.Verification.Lookup(body)`
- `POST omnichain/wallet/evm/eoa/create-wallet` → `client.Omnichain.Wallet.Evm.Eoa.CreateWallet(body)`

`test` uses `https://dev.moipayway.co`. `live` uses `https://api.moipayway.co`.

```go
ok := moipayway.VerifyWebhook(rawBody, signatureHeader, timestampHeader, apiKey, 300)
```
