/* eslint-disable no-prototype-builtins, import/no-unresolved, import/extensions */

/**
 * Files are based in ES6 for compatibility with auto-generated Amplify files.
 */

import * as mutations from "../../amplify/dev/src/graphql/mutations";
import * as queries from "../../amplify/dev/src/graphql/queries";
import * as subscriptions from "../../amplify/dev/src/graphql/subscriptions";

import awsExports from "../../amplify/dev/src/aws-exports";
import amplifyMetaConfig from "../../amplify/dev/src/amplify-meta.json";

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
