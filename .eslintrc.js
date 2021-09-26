module.exports = {
	extends: ["airbnb/base", "plugin:prettier/recommended"],
	parser: "babel-eslint",
	plugins: ["babel"],
	env: {
		browser: true
	},
	rules: {
		// See: https://github.com/benmosher/eslint-plugin-import/issues/496
		"import/no-extraneous-dependencies": 0
	}
};
