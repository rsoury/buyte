const jsonfile = require("jsonfile");
const path = require("path");

module.exports = () => {
	const libPackage = jsonfile.readFileSync(
		path.resolve(__dirname, "../../package.json")
	);
	return libPackage.version.split(".")[0];
};
