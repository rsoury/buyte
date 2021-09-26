#!/usr/bin/env node

/* eslint-disable no-console */

require("colors");
const path = require("path");
const mkdirp = require("mkdirp");
const fs = require("fs");
const rimraf = require("rimraf");
const childProcess = require("child_process");

const copy = (source, destDir) => {
	const err = fs.copyFileSync(
		source,
		path.resolve(destDir, path.basename(source))
	);
	if (err) {
		console.log(
			`Failed to copy ${source} -> ${destDir}/aws-exports.js`.red
		);
	} else {
		console.log(`Copied ${source} -> ${destDir}/aws-exports.js`.green);
	}
};

const rootDir = path.resolve(__dirname, "../../");

const environments = ["dev", "production"];

// Get env
environments.forEach(amplifyEnv => {
	// Load env
	console.log(
		`Generating config for Amplify environment: ${amplifyEnv}`.yellow
	);
	childProcess.execFileSync("amplify", ["env", "checkout", amplifyEnv], {
		stdio: "inherit",
		cwd: rootDir
	});
	// Make env config dir
	const configDir = path.resolve(__dirname, `../../src/config/${amplifyEnv}`);
	rimraf.sync(configDir);
	mkdirp.sync(configDir);
	// Copy config files to config env dir.
	const awsExportsPath = path.resolve(__dirname, "../../src/aws-exports.js");
	copy(awsExportsPath, configDir);
	const amplifyMetaPath = path.resolve(
		__dirname,
		"../../amplify/backend/amplify-meta.json"
	);
	copy(amplifyMetaPath, configDir);
});
