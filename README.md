# gamma

## Test
```
go test ./src/pkg/crypto
go test ./src/pkg/keystore
go test ./src/pkg/wallet

go test ./src/cmd/walletcli

```

## Wallet CLI

```
walletcli create -file mywallet.json -password secret
walletcli address -file mywallet.json -password secret
walletcli sign \
  -file mywallet.json \
  -password secret \
  -to deadbeef... \
  -amount 100

```

