# gamma

## Test
```
go test ./src/pkg/keystore
go test ./src/pkg/wallet
go test ./src/pkg/cmd/walletcli -v

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

