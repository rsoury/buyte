const colors = require("colors");
const AWS = require("aws-sdk");
const _get = require("lodash.get");
const argv = require("minimist")(process.argv.slice(2));

if (typeof serverless === "undefined") {
	serverless = {};
}

const region = _get(serverless, "service.provider.region", argv.region);
const service = _get(serverless, "service.service", argv.service);
const stage = _get(serverless, "service.provider.stage", argv.stage);
const name = _get(
	serverless,
	"service.custom.apiGatewayName",
	`${service}-${stage}`
);
console.log(`Removing IAM roles with prefix: ${name}`);
const iam = new AWS.IAM();
const getRoles = () =>
	new Promise((resolve, reject) => {
		iam.listRoles({}, function(err, data) {
			if (err) {
				return reject(err);
			}
			return resolve(
				data.Roles.filter(role => role.RoleName.startsWith(name))
			);
		});
	});
const removeRolePolicy = RoleName =>
	new Promise((resolve, reject) => {
		iam.deleteRolePolicy(
			{ RoleName, PolicyName: `${stage}-${service}-lambda` },
			function(err, data) {
				if (err) {
					return reject(err);
				}
				return resolve(data);
			}
		);
	});
const removeRole = RoleName =>
	new Promise((resolve, reject) => {
		iam.deleteRole({ RoleName }, function(err, data) {
			if (err) {
				return reject(err);
			}
			return resolve(data);
		});
	});

(async () => {
	const roles = await getRoles();
	try {
		await Promise.all(
			roles.map(async role => {
				await removeRolePolicy(role.RoleName);
				await removeRole(role.RoleName);
				console.log(colors.cyan(`Removed role: ${role.RoleName}`));
			})
		);
	} catch (e) {
		console.error(e);
	}
	console.log(colors.green(`Successfully removed zombie roles.`));
})();
