#!/usr/bin/env node

/* eslint-disable */

require("colors");
const AWS = require("aws-sdk");
const argv = require("minimist")(process.argv.slice(2));
const inquirer = require("inquirer");
const logger = require("tracer").console();
const getAWSConfig = require("../../serverless/resources/get-config");
const getEnv = require("../shared/get-env");

const env = getEnv();

console.log(
	`Setup merchant user attributes in Amplify environment: ${env}`.yellow
);
const {
	meta: { auth, providers }
} = getAWSConfig(env);

let { userId: Username } = argv;

const { Region: region } = providers.awscloudformation;
const { UserPoolId } = auth[Object.keys(auth)[0]].output;

if (!Username) {
	throw new Error('Username "userId" is a required CLI flag!');
}

const client = new AWS.CognitoIdentityServiceProvider({
	apiVersion: "2016-04-18",
	region
});

const questions = [
	{
		type: "input",
		name: "phone_number",
		message: "What is the contact phone number?"
	},
	{
		type: "input",
		name: "custom:store_name",
		message: "What is the store name?"
	},
	{
		type: "input",
		name: "website",
		message: "What is the website?"
	},
	{
		type: "input",
		name: "custom:currency",
		message: "What is the store currency code?"
	},
	{
		type: "input",
		name: "custom:country",
		message: "What is the store country code?"
	},
	{
		type: "input",
		name: "custom:logo",
		message: "What is the store logo URL?"
	},
	{
		type: "input",
		name: "custom:cover_image",
		message: "What is the store cover image URL?"
	},
	{
		type: "input",
		name: "locale",
		message: "What is the store locale?"
	}
];

inquirer.prompt(questions).then(answers => {
	const UserAttributes = [];
	Object.entries(answers).forEach(([key, value]) => {
		if (typeof value === "number" || !!value) {
			UserAttributes.push({
				Name: key,
				Value: value
			});
		}
	});
	client.adminUpdateUserAttributes(
		{
			UserAttributes,
			UserPoolId,
			Username
		},
		(err, data) => {
			if (err) {
				logger.error(`Error`.red);
				logger.error(err);
				logger.error(err.stack);
			} else {
				const attrObj = UserAttributes.reduce(
					(accumulator, currentValue) => {
						accumulator[currentValue.Name] = currentValue.Value;
						return accumulator;
					},
					{}
				);
				logger.info(`Success!`.green);
				logger.info(
					`Updated username: ${`${Username}`.cyan} with values ${
						`${JSON.stringify(attrObj)}`.cyan
					}`
				);
			}
		}
	);
});
