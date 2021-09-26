#!/usr/bin/env node

/* eslint-disable no-console */

require("colors");
const fs = require("fs");
const path = require("path");
const AWS = require("aws-sdk");
const logger = require("tracer").console();
const getAWSConfig = require("../../serverless/resources/get-config");
const getEnv = require("../shared/get-env");

const env = getEnv();

console.log(`Styling Cognito Hosted UI in environments: ${env}`.yellow);
const config = getAWSConfig(env);
const {
	meta: { auth }
} = config;
const { UserPoolId: userPoolId } = auth[Object.keys(auth)[0]].output;

const styles = fs
	.readFileSync(path.resolve(__dirname, "./cognito.css"), "utf8")
	.replace(/(\r\n\t|\n|\r\t)/gm, "") // remove newlines
	.replace("/*.+?*/"); // remove comments
const image = fs.readFileSync(path.resolve(__dirname, "./logo.png"));

const cognitoISP = new AWS.CognitoIdentityServiceProvider();
const params = {
	UserPoolId: userPoolId,
	CSS: styles,
	ImageFile: Buffer.from(image)
};

cognitoISP.setUICustomization(params, (err, data) => {
	if (err) logger.error(err, err.stack);
	// error
	else
		logger.log(
			`Successfully updated, new css version:  ${data.UICustomization.CSSVersion}`
				.green
		); // successful response
});
