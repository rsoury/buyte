const getAWSConfig = require("./get-config");

const environments = {
	DEV: "dev",
	PROD: "prod"
};

module.exports.dev = () => getAWSConfig(environments.DEV);

module.exports.prod = () => getAWSConfig(environments.PROD);

module.exports.devPoolId = () =>
	getAWSConfig(environments.DEV).public.aws_user_pools_id;
module.exports.devWebClientId = () =>
	getAWSConfig(environments.DEV).public.aws_user_pools_web_client_id;

module.exports.prodPoolId = () =>
	getAWSConfig(environments.PROD).public.aws_user_pools_id;
module.exports.prodWebClientId = () =>
	getAWSConfig(environments.PROD).public.aws_user_pools_web_client_id;
