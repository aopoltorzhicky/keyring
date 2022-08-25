# Keyring

Library for key management


## Install binary
```bash
go install github.com/aopoltorzhicky/keyring/cmd/go-keyring@latest
```

## Get package
```bash
go get github.com/aopoltorzhicky/keyring
```

## Usage binary

```bash
# set key
go-keyring set
```

```bash
# get key
go-keyring get
```

```bash
# remove key
go-keyring delete
```

## Usage libary

Support only `File` realization which store encrypted keys in YAML file. It realize interface `Keyring` which contains next methods:

```go
Set(service, user, password string) error
Get(service, user string) (string, error)
Delete(service, user string) error
```