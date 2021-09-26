const fs = require("fs");
const path = require("path");
const LaunchChrome = require("@serverless-chrome/lambda");
const CDP = require("chrome-remote-interface");
const version = require("./chrome/version");

const chromeSettings = {
	flags: [
		"headless",
		"disable-gpu",
		"no-first-run",
		"no-default-browser-check",
		"no-sandbox",
		"window-size=1280,1696",
		"hide-scrollbars"
	]
};

module.exports.cse = async event => {
	console.log(event);

	// Setup Chrome
	await LaunchChrome(chromeSettings);

	const [tab] = await CDP.List();
	const client = await CDP({ host: "127.0.0.1", target: tab });

	// May only need Page to evaluate JS
	const { Runtime } = client;
	await Runtime.enable();

	// Construct JS Browser Script
	const filepath = path.resolve(
		__dirname,
		"../../../node_modules/adyen-cse-web/js/adyen.encrypt.nodom.min.js"
	);
	const adyenPackage = await new Promise((resolve, reject) => {
		fs.readFile(filepath, "utf8", (err, data) => {
			if (err) {
				return reject(err);
			}
			return resolve(data);
		});
	});
	// Get data from event.
	const { cseKey, ...toEncrypt } = event;
	// Inject data
	toEncrypt.generationtime = new Date().toISOString();
	let toEncryptString = "";
	Object.entries(toEncrypt).forEach(([key, value]) => {
		toEncryptString += `${key}: "${value}",`;
	});
	const script = `
		${adyenPackage}
		(function(){
			var cseInstance = window.adyen.encrypt.createEncryption("${cseKey}", {
				enableValidations: false
			});
			return cseInstance.encrypt({
				${toEncryptString}
			});
		})()
	`;
	// console.log(script);

	// Evaluate Script
	const { result, exceptionDetails } = await Runtime.evaluate({
		expression: script
	});
	if (exceptionDetails) {
		console.log(exceptionDetails);
	}

	// It's important that we close the websocket connection,
	// or our Lambda function will not exit properly
	await client.close();

	return result;
};

module.exports.chrome_version = async () => {
	// Setup Chrome
	await LaunchChrome(chromeSettings);

	let responseBody;

	console.log("Getting version info...");

	try {
		responseBody = await version();
	} catch (error) {
		console.error("Error getting version info");
		throw error;
	}

	return {
		statusCode: 200,
		body: JSON.stringify(responseBody),
		headers: {
			"Content-Type": "application/json"
		}
	};
};
