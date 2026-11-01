# MoiPayWay SDK

Official merchant REST SDK. Environments: `test` → `https://dev.moipayway.co`, `live` → `https://api.moipayway.co`.

API reference: [https://documenter.getpostman.com/view/11919136/2sBYHXPdcb](https://documenter.getpostman.com/view/11919136/2sBYHXPdcb)

Collection paths become nested methods (example):

- `POST wallet/create` → client `wallet.create`
- `POST wallet/collection/initiate` → client `wallet.collection.initiate`
- `POST omnichain/tokenization/standards/erc20/deploy` → client `omnichain.tokenization.standards.erc20.deploy`

Deprecated flat tokenization paths (`omnichain/tokenization/erc20/...`) remain as aliases and call the `/standards/` routes.

```bash
go get github.com/mpwsocial/go@v0.2.0
```

```go
client, err := moipayway.New(os.Getenv("MOIPAYWAY_API_KEY"), "test")
wallet, err := client.Wallet.Create(map[string]any{"code": "NGN", "meta": map[string]any{"name": "Ops", "user_id": "user-uuid"}})
```
