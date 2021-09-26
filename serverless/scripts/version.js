const jsonfile = require("jsonfile");
const path = require("path");

module.exports = () => {
	const package = jsonfile.readFileSync(
		path.resolve(__dirname, "../../package.json")
	);
	return package.version.split(".")[0];
};
