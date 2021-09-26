import babel from "rollup-plugin-babel";
import resolve from "rollup-plugin-node-resolve";
import commonjs from "rollup-plugin-commonjs";
import json from "rollup-plugin-json";
import { sizeSnapshot } from "rollup-plugin-size-snapshot";
import cleanup from "rollup-plugin-cleanup";

module.exports = {
	input: "src/index.js",
	output: {
		file: "build/bundle.js",
		format: "umd",
		exports: "named",
		name: "buyte-amplify-bundle",
		compact: true,
		indent: false
	},
	plugins: [
		babel({
			exclude: "node_modules/**"
		}),
		resolve(),
		commonjs(),
		json({
			exclude: "node_modules/**",
			compact: true
		}),
		sizeSnapshot(),
		cleanup()
	]
};
