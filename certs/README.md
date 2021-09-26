## Getting up and running custom Apple Pay certificates

**Note**: The `certs/` directory is this directory.

Requirements:
- An account in the Apple Developer Program
- A recent version of Go (tested with v1.9.2)
- [`cfssl`](https://github.com/cloudflare/cfssl)
- OpenSSL/libssl-dev
- make

1. Log into your developer account. [In the developer console](https://developer.apple.com/account/ios/identifier/merchant), create a Merchant ID

2. Verify your domain by serving the `apple-developer-merchantid-domain-association` file at `https://yourdomain.com/.well-known/apple-developer-merchantid-domain-association` (the example app will do it for you if you put it in `example/static/.well-known`).
**Note**: Be careful to support one of the [supported cipher suites](https://developer.apple.com/reference/applepayjs#2166536) for HTTPS!

1. Edit the JSON files in the `certs/` directory of this repo with your merchant ID, your domain and your email address.

2. Generate a *Payment Processing Certificate* request by running `make cert-processing.certSigningRequest` in the `certs/` directory.

3. Upload the certificate request to the developer console. Select your merchant ID and click "Create Certificate" in the "Payment Processing Certificate" section

4. Download the signed certificate to `certs/cert-processing.cer`, run `make cert-processing.crt` to convert it to the proper format

5. Repeat steps 4-6 for the *Merchant Identity Certificate*, by running `make cert-merchant.certSigningRequest` and, with the certificate, `make cert-merchant.crt`

6. Move the directory `certs/` to `example/certs/`

7. Deploy the application under your domain.

8.  Go to the running application and try to pay with Apple Pay. If you don't see an Apple Pay button, you are probably visiting from an unsupported browser or device. You will not be charged.

Store your certs (*.crt and *.pem files) somewhere remotely and download them during CI/CD or programmatic environment set up.