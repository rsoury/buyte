/* eslint-disable */
// Set options as a parameter, environment variable, or rc file.
require = require("esm")(module /*, options*/);
module.exports = environment => require("./get-" + environment + "-config");
