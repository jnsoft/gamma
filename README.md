# gamma

## Test
```
go test ./src/pkg/crypto
go test ./src/pkg/keystore
go test ./src/pkg/wallet

go test ./src/cmd/walletcli

go test ./src/pkg/database
go test ./src/pkg/database -run TestInitDataDirectory_CreatesGenesisAndDb

```

## Wallet CLI

```
go run ./src/cmd/walletcli create -file wallet.json -password secret
go run ./src/cmd/walletcli address -file wallet.json -password secret

walletcli create -file mywallet.json -password secret
walletcli address -file mywallet.json -password secret
walletcli sign \
  -file mywallet.json \
  -password secret \
  -to deadbeef... \
  -amount 100

```

## Node
```
go run ./src/node -v -ip 127.0.0.1 -p 8081
```


Application code (e.g. `main` or your node package) should orchestrate:

```go
if err := database.InitDataDirectory(dataDir); err != nil {
    // handle
}

st, err := database.LoadState(dataDir)
if err != nil {
    // handle
}

// use st...

if err := database.PersistState(dataDir, st); err != nil {
    // handle
}
```

