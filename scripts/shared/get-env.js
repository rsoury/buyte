/* eslint-disable no-console */

require("colors");
const argv = require("minimist")(process.argv.slice(2));

const environments = {
	DEV: "dev",
	PROD: "prod"
};

module.exports = function getEnv() {
	let envArg = argv.env || argv.e;
	envArg = envArg === "prod" ? "production" : envArg; // For now.
	const env = Object.entries(environments).find(
		([key, value]) => value === envArg || key === envArg // eslint-disable-line no-unused-vars
	);
	if (typeof env === "undefined") {
		console.log(
			`Please provide a valid Amplify environment to the --env or -e flag ie. prod || dev`
				.magenta
		);
		process.exit(1);
	}

	return env[1];
};
