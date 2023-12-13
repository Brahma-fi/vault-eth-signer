[![Go Report Card](https://goreportcard.com/badge/github.com/Brahma-fi/vault-eth-signer)](https://goreportcard.com/report/github.com/Brahma-fi/vault-eth-signer)
[![Vault Ethereum Signer Plugin](https://github.com/Brahma-fi/vault-eth-signer/actions/workflows/vault-eth-signer.yml/badge.svg)](https://github.com/Brahma-fi/vault-eth-signer/actions/workflows/vault-eth-signer.yml)

# Vault-Eth-Signer
Eth key plugin is a HashiCorp Vault plugin that supports `ecdsa.secp256k1` based signing, with an API interface
that turns vault into a software based hardware security module(HSM) device.

The plugin exposes some endpoints to allow the client to store or generate signing keys for the `secp256k1`
curve suitable for signing input hash and transaction. You can add a private key to the key-manager or 
generate it automatically only by sending the service name to the API. Private keys are saved in the vault
as a secret, and it never exposes the private keys.

## Build

Dependencies:
* go 1.21

To build the binary you should the following command in the root of the project:
```sh
$ make build
```

This command creates an executable file with name of `vault-eth-signer` and its SHA256 checksum hash:
e.g.,
```sh
$ make build
go version go1.21.1 linux/amd64
all modules verified
cleaning...
building...
-rwxrwxr-x 1 user1 user1 20M Aug 29 17:10 /go/src/github.com/Brahma-fi/vault-eth-signer/.bin/debug/vault-eth-signer
b46018f9d398843a8003a7bf4cd2c97c9be09545d9b113fb84f10b10a74f796d  /go/src/github.com/Brahma-fi/vault-eth-signer/.bin/debug/vault-eth-signer
```

first segment of last line is the SHA256 checksum hash `b46018f9d398843a8003a7bf4cd2c97c9be09545d9b113fb84f10b10a74f796d`.

## Registering the plugin on Vault server
The plugin must be registered and enabled on the vault server as a secret engine.

### Enabling on a non-dev mode server
Before enabling the plugin on the server, it must first be registered. To register the plugin you need two things,
plugin binary file and its SHA256 checksum hash. You create both of them by running `make build` command from [here](#Build).

First copy the binary to the plugin folder for the server (consult the configuration file for the plugin
folder location). 

To register plugin in `v1.14.x` use the following commands:

#### copy the plugin binary from local to vault pod
```sh
$ kubectl cp vault-eth-signer-v1.0.0 -n vault vault-0:/vault/data/plugins/vault-eth-signer-v1.0.0
```

List all available secret plugins in the catalog:
```sh
$ vault plugin list secret

Name            Version
----            -------
aws              v0.14.0+builtin
...
```
Register a new secret plugin to the catalog:
```sh
$ vault plugin register \
  -sha256=a7d4c45a69b741f337a2f10fcb979ad3f41b92da95b4816af195bd906d725d0d \
  secret vault-eth-signer
  
Success! Registered plugin: vault-eth-signer
```

Deregister:
```sh
vault plugin deregister secret vault-eth-signer
```

> If the target vault server is enabled for TLS, and is using a self-signed certificate or other non-verifiable TLS certificate, then the command value needs to contain the switch to turn off TLS verify: `command="vault-eth-signer -tls-skip-verify"`

Once registered, just like in dev mode, it's ready to be enabled as a secret engine:
```sh
$ ./vault secrets enable -path=ethereum -description="Ethereum signer" -plugin-name=vault-eth-signer plugin
```

Disable:
```sh
$ ./vault secrets disable ethereum
```

Get information about a plugin in the catalog:
```sh
$ vault plugin info secret vault-eth-signer

Key                   Value
---                   -----
args                  []
builtin               false
command               vault-eth-signer
deprecation_status    n/a
name                  vault-eth-signer
sha256                a7d4c45a69b741f337a2f10fcb979ad3f41b92da95b4816af195bd906d725d0d
version               n/a
```

## Upgrade Plugin on a Non-Dev Mode Server

#### copy the plugin binary from local to vault pod
```sh
$ kubectl cp vault-eth-signer-v0.0.2 -n vault vault-0:/vault/data/plugins/vault-eth-signer-v0.0.2
```

1. Register a second version of your plugin. You must use the same plugin type and name (the last two arguments) as the plugin being upgraded. This is true regardless of whether the plugin being upgraded is built-in or external.

> Note: You need to increment versions manually.
> 
```sh
$ vault plugin register \
  -sha256=a7d4c45a69b741f337a2f10fcb979ad3f41b92da95b4816af195bd906d725d0d \
  -command=vault-eth-signer-v0.0.2 \
  -version=v0.0.2 \
  secret vault-eth-signer
```

2. Tune the existing mount(in our case: `ethereum` path) to configure it to use the newly registered version.
> Note: Here we need to tune the secret path we enabled after the registration process.
```sh
$ vault secrets tune -plugin-version=v0.0.2 ethereum
```

1. If you wish, you can check the updated configuration. Notice the "Version" is now different from the "Running Version".

```sh
$ vault secrets list -detailed
# many lines and columns omitted for brevity
Path      Plugin         Version  Running Version Running SHA256                                                  
----      ------         -------  --------------- --------------                                                  
...
ethereum/ vault-eth-signer v0.0.2   n/a             a7d4c45a69b741f337a2f10fcb979ad3f41b92da95b4816af195bd906d725d0d
...
# first running version wasn't registered over a version, then you are seeing: n/a
```

4. Finally, trigger a plugin [reload to reload](https://developer.hashicorp.com/vault/docs/commands/plugin/reload) all mounted backends using that plugin or a subset of the mounts using that plugin with either the plugin or mounts flag respectively.

```sh
$ vault plugin reload -plugin vault-eth-signer

# after reload you can execute again:
$ vault secrets list -detailed
# many lines omitted for brevity
Path      Plugin         Version  Running Version Running SHA256                                                  
----      ------         -------  --------------- --------------                                                  
ethereum/ vault-eth-signer v0.0.2   v0.0.2          93b2ea278065d4659ebfb664a14781d9bc87f36b3d458c6a66e4be9f9414d79f
```

5. Test the new version of the plugin.

```sh
$ vault list ethereum/key-managers
$ vault read ethereum/key-managers/user-service
```

6. Last but not least. Please remove the previous plugin binary version

for example:
```sh
$ rm -rf /vault/data/plugins/vault-eth-signer-v0.0.1
```

## Interacting with the vault-eth-signer Plugin

### Creating A New Key-manager
Create a new key-manager in the vault by POSTing to the `/key-managers` endpoint.

Using vault cli:
```sh
$ vault write ethereum/key-managers serviceName="user-service"
Key             Value
---             -----
public_key      04c29a18f57de204ce7797314c7fe61f4a903bd7d25204b63f227be660610e055faca9fec68e74d51dbdde9ebac6286f2975889ccbb97eeb998eefbf54c3f31803
service_name    user-service
```

Using the REST API:
```sh
$ curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"serviceName": "user-service"}' http://localhost:8200/v1/ethereum/key-managers |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "service_name": "user-service"
    "public_key": "045809f2cb46e0a05b7e535e765dc3c658d2a196170f80570900483a46c7875720a2a885656d77181d1107bee5b2f2758a5be3fe58037693c10e7adf16746367bc"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Importing An Existing Private Key
You can also create a new key-manager by importing from an existing private key. The private key is 
passed in as a hexidecimal string, without the '0x' prfix.

Using the REST API:
```sh
$ curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"serviceName": "user-service", "privateKey":"3ee65159f7aa057c482b1041f18f37ce90ef5e460cb46fd3fa0c40fbae41c7e1"}' http://localhost:8200/v1/ethereum/key-managers |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "service_name": "user-service"
    "public_key": "045809f2cb46e0a05b7e535e765dc3c658d2a196170f80570900483a46c7875720a2a885656d77181d1107bee5b2f2758a5be3fe58037693c10e7adf16746367bc"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### List Existing Key-managers
The list command only returns the service name that owned the key-manager. 

Using Vault CLI:
```sh
$ vault list ethereum/key-managers
Keys
----
user-service
```

Using the REST API:
```sh
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/ethereum/key-managers?list=true |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "keys": [
      "user-service",
      "test-service"
    ]
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Reading Individual Key-managers
Inspect the key using the service name. Only the service name and public key of the key-manager is returned.

Using Vault CLI:
```sh
$ vault read ethereum/key-managers/user-service
Key             Value
---             -----
public_key      04c29a18f57de204ce7797314c7fe61f4a903bd7d25204b63f227be660610e055faca9fec68e74d51dbdde9ebac6286f2975889ccbb97eeb998eefbf54c3f31803
service_name    user-service
```

Using the REST API:
```sh
$  curl -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/ethereum/key-managers/user-service |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "service_name": "user-service"
    "public_key": "045809f2cb46e0a05b7e535e765dc3c658d2a196170f80570900483a46c7875720a2a885656d77181d1107bee5b2f2758a5be3fe58037693c10e7adf16746367bc"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Sign a hashed data
Use one of the key-managers to sign a hash.

Using the REST API:
```
$  curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/ethereum/key-managers/user-service/sign -d '{"hash":"0xaf41db230000000000000000000000000000000000000000000000000000000000000023"}' |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signature": "af3a5d16ea4c1fbd8700927df140d4626047a0340b55507f9aa6ede27fac86e91337df9839628af70d6bd520af85ad4853c9d90396c3d32e2f94c2be24b4619b01",
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

### Sign a transaction

```shell
$  curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/ethereum/key-managers/user-service/txn/sign -d '{"data":"0x41ef56c20000000000000000000000000000000000000000000000000000000000000012","gas":2123,"gasPrice":0,"nonce":"0x0","to":"0xaa1fd73c4981aeb9d051eff70b055eb50ba94c32"}' |jq

{
  "request_id": "4ab32813-ca32-a143-8791-e9dbc5264123",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signed_transaction": "0xf781208083015f9094b401069f06a24155774bf8a0f6654ea299c8f68780a460fe47b10000000000000000000000000000000000000000000000000000000000000014840ea23e3fa088f4f5505f6f1da6c9a543863d5c7537e0dfc58618dbf34517c80875283d1e07a0583ecdc23ba3333a3f25611fffe0ec7fb585e9b9af93941f6e3ef8c8ef843513",
    "transaction_hash": "0x8b2a7960a9398ae73b994c46fcb8834068195a2d3468c40a1eaad7ed4a15e32a"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```
