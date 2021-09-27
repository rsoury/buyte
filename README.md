# Buyte - WIP

**Open Source Digital Wallet Checkout**.  
Accelerate your eCommerce website checkout with an all-in-one solution that integrates Apple Pay and Google Pay with your Payment Processor(s). Completely serverless tokenisation and payment processing for digital wallet solutions built for the AWS Cloud.

![Buyte Logo](https://github.com/rsoury/buyte/blob/master/docs/logo.png)

[See a demo on YouTube](https://www.youtube.com/watch?v=fKnVh8_HLwk)  
[See some example images of the application](https://github.com/rsoury/buyte/blob/master/examples/images/)

## Features

- Hosted 
- Checkout for each 

## Supported Digital Wallets

- Apple Pay
- Google Pay

## Supported Payment Processors

- Stripe
- Adyen

## Overview

This repository includes all core services (Serverless API, CLI, Scripts, and AWS Amplify Configuration).  

The produced Binary is a CLI tool capable of running commands for administrative and development purposes as well as a command for starting the Tokenisation and Payment Processing API.

## Architecture

![Buyte Architecture](https://github.com/rsoury/buyte/blob/master/docs/Buyte-Architecture.jpeg)

## Installation Requirements

- Go 1.16.0+
- Node.js 10.0+

## Supporting Repositories

Once this codebase has been set up, please visit:

- [Buyte Dashboard](https://github.com/rsoury/buyte-dashboard)
  - Set up the Administration Portal where Checkouts can be created and connected to different Payment Processors.
- [Buyte Checkout](https://github.com/rsoury/buyte-checkout)
  - Configure and then install the Buyte Checkout JS library into your website referencing the Checkout ID produced in your Buyte Dashboard.

## Getting Started

1. Clone the repository `git clone git@github.com:rsoury/buyte.git`
2. Install Node.js Dependencies: `yarn`

### 1. Amplify

1. Set up your Amplify Configuration
   1. `amplify configure`
   2. Make a `dev` directory under the amplify directory. In each directory (`dev` or `prod`), you can manage an environment for your Amplify configurations. We advise committing these configurations to a private repository. These configurations will include references to components in your cloud infrastructure.
      1. `mkdir -p ./amplify/dev`
      2. `cd ./amplify/dev`
   3. `amplify init`
2. Add a Data Storage (DynamoDB and AppSync GraphQL) to Amplify
   `amplify api add`
   1. Select **"GraphQL"** for the interface and **"Amazon Cognito User Pools"** for authentication
   2. Select **"Yes"** for "Do you want to configure advanced settings for the GraphQL API" and provide the path to the GraphQL Schema `../graphql.schema`
3. Add Auth (Cognito) to Amplify  
   `amplify add auth`  
   You should receive a message that Auth has already been added.  
4. Push your Amplify configuration  
   `amplify push`
   Ensure you auto-generate code from GraphQL schema when prompted.
5. Add the GraphQL Endpoint to your Environment file `.env.development` or `.env.production`

### 2. ApplePay Certificates

[Visit the Certs directory](https://github.com/rsoury/buyte/blob/master/certs/) and follow the guide to produce your Apple Pay Certificates

### 3. Serverless

1. Deploy to AWS
   For Production - `sls deploy --env production --stage prod`
   For Development - `sls deploy`

For development, use `sls offline` to test requests to a locally hosted web server.

### 3. CLI

1. Install Golang dependencies - `go mod download`
2. Build the binary - `make`
3. Run the API in Development - `buyte api`
4. Run the API in Production - `buyte api --production`

For development, use `make init && make watch` to rebuild the binary on file change.

### 5. Finalise Cognito

Go to your AWS Console and visit the Cognito Portal.

1. Add the Serverless Lambda Functions as the Cognito Triggers.
2. Add a Domain Name to your Hosted UI
3. Update your `COGNITO_CLIENTID` and `COGNITO_USERPOOLID` your Environment file(s).

Further configuration for Cognito will continue in the [Buyte Dashboard](https://github.com/rsoury/buyte-dashboard) set up.

## Testing

```shell
go test -v
```

## Caveats

- [ApplePay](https://github.com/rsoury/applepay/) dependency has some caveats:
  - You may need to change your `PKG_CONFIG_PATH` to include OpenSSL. For example, on my Mac I use `PKG_CONFIG_PATH=$(brew --prefix openssl)/lib/pkgconfig go test`.
  - After Serverless Deploy, go to AWS Cognito in AWS console and save the Cognito Triggers page.  
  **There is a bug here where without saving manually, they will not run when required.**

## Development Endpoints

Development endpoints were made to assist in spinning up a landing page with the appropriate digital wallet.
