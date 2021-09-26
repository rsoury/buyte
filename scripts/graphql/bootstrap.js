#!/usr/bin/env node

/*
	In order to run this bootstrap,
	you need to flip the Auth requirements for your AppSync GraphQL engine to AWS IAM.
	Once complete, flip back to Congito Users.
 */

/* eslint-disable no-console */

/*
	Some Setup to use Apollo client in Node.js
 */
global.navigator = () => null;
global.WebSocket = require("ws");
require("es6-promise").polyfill();
require("isomorphic-fetch");
// -- End Apollo Setup

require("colors");
const logger = require("tracer").console();
const gql = require("graphql-tag");
const get = require("lodash.get");
const prompt = require("prompt");
const argv = require("minimist")(process.argv.slice(2));
const AmazonCognitoIdentity = require("amazon-cognito-identity-js");
const { AUTH_TYPE } = require("aws-appsync/lib/link/auth-link");
const { default: AWSAppSyncClient } = require("aws-appsync");
const getAWSConfig = require("../../serverless/resources/get-config");
const getEnv = require("../shared/get-env");

const env = getEnv();

console.log(
	`Bootstrapping Graphql connected Data store in Amplify environment: ${env}`
		.yellow
);
const {
	meta: { api, auth, providers }
} = getAWSConfig(env);

const region = providers.awscloudformation.Region;
const url = api[Object.keys(api)[0]].output.GraphQLAPIEndpointOutput;
// const apiId = api[name].output.GraphQLAPIIdOutput;

const authData = auth[Object.keys(auth)[0]];
const cognitoUserPoolId = authData.output.UserPoolId;
const cognitoClientId = authData.output.AppClientIDWeb;

prompt.start();
prompt.override = {
	Username: argv.u,
	Password: argv.p
};
prompt.get(["Username", "Password"], async (err, { Username, Password }) => {
	if (err) {
		return logger.err(err);
	}

	// Set up Cognito Client
	const userPool = new AmazonCognitoIdentity.CognitoUserPool({
		UserPoolId: cognitoUserPoolId,
		ClientId: cognitoClientId
	});
	const authenticationDetails = new AmazonCognitoIdentity.AuthenticationDetails(
		{
			Username,
			Password
		}
	);
	const cognitoUser = new AmazonCognitoIdentity.CognitoUser({
		Username,
		Pool: userPool
	});

	// Set up Apollo client
	const Client = new AWSAppSyncClient({
		url,
		region,
		auth: {
			type: AUTH_TYPE.AMAZON_COGNITO_USER_POOLS,
			jwtToken: async () => {
				const result = await new Promise((resolve, reject) => {
					cognitoUser.authenticateUser(authenticationDetails, {
						onSuccess(resp) {
							resolve(resp);
						},
						onFailure(error) {
							logger.error(error);
							reject(error);
						}
					});
				});
				return result.getIdToken().getJwtToken();
			}
		},
		disableOffline: true // Uncomment for AWS Lambda
	});

	const client = await Client.hydrated();

	// Create both Mobile Payment Options
	const listMobileWebPayments = await client.query({
		query: gql(`
			query ListMobileWebPayments {
			  listMobileWebPayments {
			    items {
			      id
			      name
			      image
			    }
			  }
			}
		`)
	});
	logger.log(listMobileWebPayments);

	let paymentOptions = get(
		listMobileWebPayments,
		"data.listMobileWebPayments.items",
		[]
	);
	logger.log(paymentOptions);

	// If not mobile payment options exist...
	if (!paymentOptions.length) {
		const CreateMobileWebPayment = await Promise.all(
			[
				{
					name: "Apple Pay",
					image:
						"https://s3-ap-southeast-2.amazonaws.com/buyte.au/assets/mobile-web-payments/apple-pay.png"
				},
				{
					name: "Google Pay",
					image:
						"https://s3-ap-southeast-2.amazonaws.com/buyte.au/assets/mobile-web-payments/google-pay.png"
				}
			].map(input =>
				client.mutate({
					mutation: gql(`
						mutation CreateMobileWebPayment($input: CreateMobileWebPaymentInput!) {
							createMobileWebPayment(input: $input) {
								id
								name
								image
							}
						}
					`),
					variables: {
						input
					}
				})
			)
		);
		paymentOptions = CreateMobileWebPayment.map(
			({ data }) => data.createMobileWebPayment
		);
		logger.log(paymentOptions);

		// If your Payment Options haven't been bootstrapped, then your providers probably haven't either.
		const CreatePaymentProvider = await Promise.all(
			[
				{
					name: "Stripe",
					image:
						"https://s3-ap-southeast-2.amazonaws.com/buyte.au/assets/providers/stripe/logo.png"
				},
				{
					name: "Adyen",
					image:
						"https://s3-ap-southeast-2.amazonaws.com/buyte.au/assets/providers/adyen/logo.png"
				}
			].map(input =>
				client.mutate({
					mutation: gql(`
					mutation CreatePaymentProvider($input: CreatePaymentProviderInput!) {
						createPaymentProvider(input: $input) {
							id
							name
							image
						}
					}
				`),
					variables: {
						input
					}
				})
			)
		);

		const paymentProviders = CreatePaymentProvider.map(
			({ data }) => data.createPaymentProvider
		);

		// Now with payment providers and payment options, you just need to create the providerPaymentOption
		const mutations = [];
		paymentProviders.forEach(({ id: providerId }) => {
			paymentOptions.forEach(({ id: optionId }) => {
				mutations.push(
					client.mutate({
						mutation: gql(`
							mutation CreateProviderPaymentOption(
								$input: CreateProviderPaymentOptionInput!
							) {
								createProviderPaymentOption(input: $input) {
									id
								}
							}
						`),
						variables: {
							input: {
								providerPaymentOptionProviderId: providerId,
								providerPaymentOptionPaymentOptionId: optionId
							}
						}
					})
				);
			});
		});
		await Promise.all(mutations);

		logger.log(`Database is successfully bootstrapped!`.green);
	} else {
		logger.log(`Database already bootstrapped!`.green);
	}

	return null;
});
