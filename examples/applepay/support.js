/*
 Copyright (C) 2016 Apple Inc. All Rights Reserved.
 See LICENSE.txt for this sample’s licensing information

 Abstract:
 The main client-side JS. Handles displaying the Apple Pay button and requesting a payment.
 */

/**
 * This method is called when the page is loaded.
 * We use it to show the Apple Pay button as appropriate.
 * Here we're using the ApplePaySession.canMakePayments() method,
 * which performs a basic hardware check.
 *
 * If we wanted more fine-grained control, we could use
 * ApplePaySession.canMakePaymentsWithActiveCards() instead.
 */
document.addEventListener("DOMContentLoaded", function() {
	if (window.ApplePaySession) {
		if (ApplePaySession.canMakePayments) {
			showApplePayButton();
		}
	}
});

function showApplePayButton() {
	HTMLCollection.prototype[Symbol.iterator] =
		Array.prototype[Symbol.iterator];
	// const buttons = document.getElementsByClassName("apple-pay-button");
	// for (let button of buttons) {
	// 	button.className += " visible";
	// 	button.onclick = function(e) {
	// 		console.log(e);
	// 	};
	// }
	const button = document.getElementById("apple-pay-button");
	button.className += " visible";
	button.onclick = function(e) {
		console.log(e);
		applePayButtonClicked();
	};
}

/**
 * Apple Pay Logic
 * Our entry point for Apple Pay interactions.
 * Triggered when the Apple Pay button is pressed
 */
function applePayButtonClicked() {
	const totalAmount = 5500;
	const totalAmountFloat = parseFloat(totalAmount / 100);
	const paymentRequest = {
		countryCode: "AU",
		currencyCode: "AUD",
		shippingMethods: [
			{
				label: "Free Shipping",
				amount: "0.00",
				identifier: "free",
				detail: "Delivers in five business days"
			},
			{
				label: "Express Shipping",
				amount: "0.49",
				identifier: "express",
				detail: "Delivers in two business days"
			}
		],

		lineItems: [
			{
				label: "Shipping",
				amount: "0.00"
			}
		],

		total: {
			label: "Apple Pay Example",
			amount: `${totalAmountFloat}`
		},

		supportedNetworks: ["amex", "discover", "masterCard", "visa"],
		merchantCapabilities: ["supports3DS"],

		requiredShippingContactFields: ["email"]
	};

	const session = new ApplePaySession(1, paymentRequest);

	/**
	 * Merchant Validation
	 * We call our merchant session endpoint, passing the URL to use
	 */
	session.onvalidatemerchant = event => {
		console.log("Validate merchant");
		const { validationURL } = event;
		console.log(event);
		getApplePaySession(event.validationURL).then(function(response) {
			console.log(response);
			session.completeMerchantValidation(response);
		});
	};

	/**
	 * Shipping Method Selection
	 * If the user changes their chosen shipping method we need to recalculate
	 * the total price. We can use the shipping method identifier to determine
	 * which method was selected.
	 */
	session.onshippingmethodselected = event => {
		const shippingCost =
			event.shippingMethod.identifier === "free" ? "0.00" : "0.49";
		const totalCost =
			event.shippingMethod.identifier === "free"
				? `${totalAmountFloat}`
				: `${totalAmountFloat + 0.5}`;

		const lineItems = [
			{
				label: "Shipping",
				amount: shippingCost
			}
		];

		const total = {
			label: "Apple Pay Example",
			amount: totalCost
		};

		session.completeShippingMethodSelection(
			ApplePaySession.STATUS_SUCCESS,
			total,
			lineItems
		);
	};

	/**
	 * Payment Authorization
	 * Here you receive the encrypted payment data. You would then send it
	 * on to your payment provider for processing, and return an appropriate
	 * status in session.completePayment()
	 */
	session.onpaymentauthorized = event => {
		// Send payment for processing...
		const payment = {
			result: event.payment,
			checkoutId: "a1aeb05b-3f02-4dd5-b9a9-941377fb9c15",
			paymentMethodId: "11cfbaf1-8094-498d-af51-c38d7cf4bc0d",
			shippingMethodId: "18c13641-c8a4-4e5d-a237-c6a7803477b3",
			currency: "AUD",
			country: "AU",
			amount: totalAmount,
			rawPaymentRequest: paymentRequest
		};
		console.log(event);
		const r = new XMLHttpRequest();
		r.open("POST", "/v1/public/applepay/process/");
		r.onreadystatechange = function() {
			if (r.readyState != 4) {
				return;
			}
			if (r.status != 200) {
				session.completePayment(ApplePaySession.STATUS_FAILURE);
			}
			session.completePayment(ApplePaySession.STATUS_SUCCESS);
			console.log(r);
		};
		r.setRequestHeader("Content-Type", "application/json");
		r.setRequestHeader(
			"Authorization",
			"Bearer pk_YICkLIG2LPQfZvM3YE00LSDtXIQ5KvYfAPGiYSKkYPNyAPM2Xw9l"
		);
		r.send(JSON.stringify(payment));
	};

	// All our handlers are setup - start the Apple Pay payment
	session.begin();
}
