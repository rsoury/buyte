/* eslint-disable no-prototype-builtins, import/no-unresolved, import/extensions */

/**
 * Files are based in ES6 for compatibility with auto-generated Amplify files.
 */

import * as mutations from "../../amplify/prod/src/graphql/mutations";
import * as queries from "../../amplify/prod/src/graphql/queries";
import * as subscriptions from "../../amplify/prod/src/graphql/subscriptions";

import awsExports from "../../amplify/prod/src/aws-exports";
import amplifyMetaConfig from "../../amplify/prod/src/amplify-meta.json";

export const graphql = {
	mutations,
	queries,
	subscriptions
};

export const environments = {
	DEV: "dev",
	PROD: "prod"
};

export const config = {
	public: awsExports,
	meta: amplifyMetaConfig
};
