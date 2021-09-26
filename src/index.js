/* eslint-disable no-prototype-builtins */

import * as mutations from "./graphql/mutations";
import * as queries from "./graphql/queries";
import * as subscriptions from "./graphql/subscriptions";

import devAwsExports from "./config/dev/aws-exports";
import devAmplifyMetaConfig from "./config/dev/amplify-meta.json";
import prodAwsExports from "./config/production/aws-exports";
import prodAmplifyMetaConfig from "./config/production/amplify-meta.json";

export const graphql = {
	mutations,
	queries,
	subscriptions
};

export const environments = {
	DEV: "dev",
	PROD: "production"
};

const config = {
	[environments.DEV]: {
		public: devAwsExports,
		meta: devAmplifyMetaConfig
	},
	[environments.PROD]: {
		public: prodAwsExports,
		meta: prodAmplifyMetaConfig
	}
};

const getConfig = environment =>
	config.hasOwnProperty(environment) ? config[environment] : {};

export const getAWSConfig = getConfig;

export default getConfig;
