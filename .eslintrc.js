module.exports = {
	extends: ["airbnb/base", "plugin:prettier/recommended"],
	env: {
		node: true
	},
	globals: {
		serverless: "writable"
	},
	rules: {
		// See: https://github.com/benmosher/eslint-plugin-import/issues/496
		"import/no-extraneous-dependencies": 0,
		"no-prototype-builtins": 0,
		"no-console": 0
	}
};
