<div align="center">
  <h1>Buyte</h1>
</div>

<div align="center">
  <strong>Digital wallet payment orchestration on a serverless stack</strong>
</div>

<div align="center">
  Accelerate your eCommerce website checkout with an enhanced digital wallet payment orchestration. Integrate native Apple Pay and Google Pay with your Payment Processor(s).
</div>
<br/>

![Buyte Banner](https://github.com/rsoury/buyte/blob/master/examples/images/banner-1544x500.jpg)

<div align="center">
   <a href="https://www.youtube.com/watch?v=fKnVh8_HLwk">Demo</a>
   <span> | </span>
   <a href="https://github.com/rsoury/buyte/blob/master/examples/images/">Example images</a>
   <span> | </span>
   <a href="https://github.com/rsoury/buyte/blob/master/docs/Buyte-Architecture.jpeg">Architecture</a>
   <span> | </span>
   <a href="https://github.com/rsoury/buyte/blob/master/docs/Buyte-Checkout-Sequence-Diagram.jpeg">Sequence diagram</a>
</div>

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Supported Digital Wallets](#supported-digital-wallets)
- [Supported Payment Processors](#supported-payment-processors)
- [Overview](#overview)
- [Installation Requirements](#installation-requirements)
- [Supporting Repositories](#supporting-repositories)
- [Getting Started](#getting-started)
  - [1. Amplify](#1-amplify)
  - [2. ApplePay Certificates](#2-applepay-certificates)
  - [3. Serverless](#3-serverless)
  - [3. CLI](#3-cli)
  - [5. Finalise Cognito](#5-finalise-cognito)
- [Database Set Up](#database-set-up)
- [(Optional) Update Database Schema](#optional-update-database-schema)
- [Testing](#testing)
- [Caveats](#caveats)
- [Development Endpoints](#development-endpoints)
- [Contribution](#contribution)
- [Enterprise Support](#enterprise-support)
- [Found this repo interesting?](#found-this-repo-interesting)

## Features

- **Frictionless:** Skip passwords, account forms and the standard checkout flow. Minimise time to checkout, maximising conversions. 
- **Familiar:** Allow your users to checkout using the same technology they use in-store.
- **Secure:** Apple and Google's security infrastructure prevents card/payment details from leaving your user's devices.
- **Widgetised:** Complete your checkout from anywhere on your website with the [Buyte Checkout](https://github.com/rsoury/buyte-checkout).
- **Extensible:** Bring your Payment Processor and pass on raw/decrypted payment data.
- **Serverless:** Scalable by default.

## Supported Digital Wallets

- Apple Pay
- Google Pay

## Supported Payment Processors

- Stripe
- Adyen
- [**Add your own**](#contribution)

## Overview

This repository responsible for tokenisation of digital wallet payloads before passing on raw/decrypted payment data to the connected Payment Processor.

It is comprised of the Serverless API, CLI, Scripts, and AWS Amplify Configuration

The produced Binary is a CLI tool capable of running commands for administrative and development purposes as well as a command for starting the Tokenisation and Payment Processing API.

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

## Database Set Up

*We highly advise configuring your `.env.development` or `.env.production` files before proceeding to minimise the number of flags passed to each command.*

1. Create yourself a super user
   ```
   buyte create-super-user -e youremail@example.com -p somepassword
   ```
   1. Add your `ADMIN_USERNAME` and `ADMIN_PASSWORD` environment variables to your `.env` file
2. Set up Cognito Custom User Attributes - for [Dashboard](https://github.com/rsoury/buyte-dashboard)
   ```
   buyte auth-setup
   ```
3. Create your payment options
   ```
   buyte payments add --name "Apple Pay" --image https://s3.url/to-imaage.png
   buyte payments add --name "Google Pay"
   ```
4. Create your payment providers
   ```
   buyte providers add --name Adyen
   buyte providers add --name Stripe
   ```
5. Use the List commands to identify the Ids of each Payment and Provider record. ie. `buyte payments list` or `buyte providers list`
6. Connect your Payment Options to each of your Payment Providers.
   ```
   buyte providers connect --provider-id adyen-xxxx-xxxx-xxxx --payment-id applepay-yyyy-yyyy-yyyy
   buyte providers connect --provider-id stripe-xxxx-xxxx-xxxx --payment-id applepay-yyyy-yyyy-yyyy
   buyte providers connect --provider-id adyen-xxxx-xxxx-xxxx --payment-id googlepay-yyyy-yyyy-yyyy
   buyte providers connect --provider-id stripe-xxxx-xxxx-xxxx --payment-id googlepay-yyyy-yyyy-yyyy
   ```
7. List your providers to check which payment options are connected - `buyte providers list`

You should see an output of the Provider details and their associated Payment Options.

## (Optional) Update Database Schema

In case you have unique storage requirements that fall outside of the schema, here's a simple way to update your schema. 

1. Create a symlink to the Amplify backend directory.
   ```
   ln -s ./amplify/schema.graphql ./amplify/dev/amplify/backend/api/buytedev/schema.graphql
   ```
2. Make modifications to `./amplify/schema.graphql`
3. `cd ./amplify/dev/`
4. `amplify api update`
5. `amplify push`

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

## Contribution

Simply fork this repo and make it your own, or create a pull request and we can build something awesome together!

## Enterprise Support

Whether you're looking to integrate a Legacy Payment Processor or Banking API, or looking for managed deployment and operation in your cloud, you can contact us at [Web Doodle](https://www.webdoodle.com.au/?ref=github-buyte) to discuss tailored solutions.

## Found this repo interesting?

Star this project ⭐️⭐️⭐️, and feel free to follow me on [Github](https://github.com/rsoury), [Twitter](https://twitter.com/@ryan_soury) or [Medium](https://rsoury.medium.com/)