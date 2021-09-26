/*
 Copyright (C) 2016 Apple Inc. All Rights Reserved.
 See LICENSE.txt for this sample’s licensing information

 Abstract:
 A helper function that requests an Apple Pay merchant session using a promise.
 */

function getApplePaySession(url) {
	return new Promise(function(resolve, reject) {
		const xhr = new XMLHttpRequest();
		// xhr.open('POST', '/getApplePaySession');
		xhr.open("POST", "/v1/public/applepay/session/");
		xhr.onload = function() {
			if (this.status >= 200 && this.status < 300) {
				let json = {};
				try {
					json = JSON.parse(xhr.response);
				} catch (e) {
					console.log(xhr.response);
					return reject(e);
				}
				resolve(JSON.parse(xhr.response));
			} else {
				reject({
					status: this.status,
					statusText: xhr.statusText
				});
			}
		};
		xhr.onerror = function() {
			reject({
				status: this.status,
				statusText: xhr.statusText
			});
		};
		xhr.setRequestHeader("Content-Type", "application/json");
		xhr.setRequestHeader(
			"Authorization",
			"Bearer pk_YICkLIG2LPQfZvM3YE00LSDtXIQ5KvYfAPGiYSKkYPNyAPM2Xw9l"
		);
		xhr.send(JSON.stringify({ url }));
	});
}
