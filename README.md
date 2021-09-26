# Buyte

This repository includes all core services (Serverless API, CLI, Scripts, and AWS Amplify Configuration).  

The Binary built is a CLI tool capable of running commands for administrative and development purposes as well as a command for starting the Tokenisation and Payment Processing API.

## Requirements

- Go 1.9.1+
- CFSSL: `go get -u github.com/cloudflare/cfssl/cmd/...`

## Getting Started

1. Download: `git clone git@github.com:rsoury/buyte.git`
2. Install Node.js Depdencies: `yarn`
3. Build: `make`
4. Run: `buyte api --production`

## Testing

```shell
go test -v
```

## Caveats

- [ApplePay](https://github.com/rsoury/applepay/) depdency has some caveats
  - You may need to change your `PKG_CONFIG_PATH` to include OpenSSL. For example, on my Mac I use `PKG_CONFIG_PATH=$(brew --prefix openssl)/lib/pkgconfig go test`.
  - After Serverless Deploy, go to AWS Cognito and save the Cognito Triggers page.  
There is a bug here where without the save, they will not run when required.

## Development Endpoints

Development endpoints were made to assist in spinning up a landing page with the appropirate digital wallet.