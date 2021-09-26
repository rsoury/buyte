const CDP = require("chrome-remote-interface");

module.exports = async () => CDP.Version();
