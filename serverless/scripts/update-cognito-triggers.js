/* eslint-disable no-console */

const colors = require("colors");
const AWS = require("aws-sdk");
const get = require("lodash.get");

if (typeof serverless === "undefined") {
	throw new Error("Run script in Serverless hook: Cannot find serverless.");
}

const region = get(serverless, "service.provider.region");
const cognitoISP = new AWS.CognitoIdentityServiceProvider({
	region
});
const sts = new AWS.STS({ apiVersion: "2011-06-15" });
const getAccountId = () =>
	new Promise((resolve, reject) => {
		sts.getCallerIdentity({}, (err, data) => {
			if (err) {
				return reject(err);
			}
			return resolve(data.Account);
		});
	});
const describeUserPool = UserPoolId =>
	new Promise((resolve, reject) => {
		cognitoISP.describeUserPool({ UserPoolId }, (err, data) => {
			if (err) {
				return reject(err);
			}
			return resolve(data);
		});
	});
const updateUserPool = params =>
	new Promise((resolve, reject) => {
		cognitoISP.updateUserPool(params, (err, data) => {
			if (err) {
				return reject(err);
			}
			return resolve(data);
		});
	});

(async () => {
	const accountId = await getAccountId();
	const service = get(serverless, "service.service");
	const stage = get(serverless, "service.provider.stage");
	const userPoolId = get(
		serverless,
		`service.custom.cognito.${stage}.userPoolId`
	);
	console.log(
		colors.cyan(
			`Region: ${region}  |  Account: ${accountId}  |  Service: ${service}  |  Stage: ${stage}`
		)
	);
	console.log("");
	if (!region) {
		throw new Error("Cannot find Region");
	}
	if (!accountId) {
		throw new Error("Cannot find Account");
	}
	if (!service) {
		throw new Error("Cannot find Service");
	}
	if (!stage) {
		throw new Error("Cannot find Provider Stage");
	}
	if (!userPoolId) {
		throw new Error("Cannot find User Pool ID");
	}
	const getLambdaArn = funcName =>
		!funcName.startsWith("arn")
			? `arn:aws:lambda:${region}:${accountId}:function:${service}-${stage}-${funcName}`
			: funcName;
	const newLambdaConfig = Object.entries(serverless.service.functions).reduce(
		(accumulator, [key, value]) => {
			if (value.hasOwnProperty("cognitoTrigger")) {
				if (value.cognitoTrigger) {
					accumulator[value.cognitoTrigger] = getLambdaArn(key);
				}
			}
			return accumulator;
		},
		{}
	);
	console.log(colors.cyan(`New Lambda Config:`));
	console.log(newLambdaConfig);

	const userPoolKeys = [
		"AdminCreateUserConfig",
		"AutoVerifiedAttributes",
		"DeviceConfiguration",
		"EmailConfiguration",
		"EmailVerificationMessage",
		"EmailVerificationSubject",
		"LambdaConfig",
		// "MfaConfiguration",
		"Policies",
		"SmsAuthenticationMessage",
		"SmsConfiguration",
		"SmsVerificationMessage",
		"UserPoolAddOns",
		"UserPoolTags",
		"VerificationMessageTemplate"
	];
	const data = await describeUserPool(userPoolId);
	const params = userPoolKeys.reduce((accumulator, key) => {
		if (data.UserPool.hasOwnProperty(key)) {
			accumulator[key] = data.UserPool[key];
		}
		return accumulator;
	}, {});
	params.Policies.PasswordPolicy.TemporaryPasswordValidityDays = 7;
	delete params.AdminCreateUserConfig.UnusedAccountValidityDays;
	params.MfaConfiguration = "OFF"; // TODO: Update this... but am forcing no MFA for now.
	params.UserPoolId = userPoolId;
	params.LambdaConfig = Object.assign(
		{},
		params.LambdaConfig || {},
		newLambdaConfig
	);

	// console.log(util.inspect(params, false, null, true));

	await updateUserPool(params);

	console.log(
		colors.green(
			`Successfully updated Cognito triggers for UserPool:  ${userPoolId}`
		)
	);
})();
