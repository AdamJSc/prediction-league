<template>
    <div class="payment-step-container">
        <div class="row">
            <div class="col">
                <h2>Step 2/2: Payment</h2>
            </div>
        </div>
        <transition name="fade">
            <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                <ul><li v-for="msg in errorMessages">{{msg}}</li></ul>
            </div>
        </transition>
        <div v-if="entryData.needsPayment">
            <div class="row">
                <div class="col">
                    <div class="payment-summary">
                        <p>Use the options below to make a payment via your <strong>PayPal account</strong> or a <strong>Debit or Credit card</strong>.</p>
                        <p>All payments, including card payments, are processed securely via <a href="https://paypal.com" target="_blank">PayPal</a> so you don't need an existing PayPal account in order to pay.</p>
                        <p>If you encounter any technical issues, please email <a v-bind:href="mailToSupportEmail">{{supportEmailPlainText}}</a>
                        <div class="payment-amount">
                            Payment due: {{entryFeeData.label}}
                        </div>
                    </div>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <div id="paypal-button-container"></div>
                </div>
            </div>
        </div>
        <div v-else>
            <div class="row">
                <div class="col">
                    <p>It looks like you're running in dev mode without PayPal configured...</p>
                    <p>Go ahead and make an empty payment!</p>
                    <button class="btn btn-primary btn-lg" v-on:click="updateEntryPayment('other', 'PAYMENTREF12345', '*LOCALHOST BANK')">Make Payment</button>
                </div>
            </div>
            <div class="row">
                <div class="col">
                    <div id="paypal-button-container"></div>
                </div>
            </div>
        </div>
    </div>
</template>


<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationPayment',
        props: {
            entryData: {
                type: Object
            },
            entryFeeData: {
                type: Object
            },
            supportEmailFormatted: {
                type: String
            },
            supportEmailPlainText: {
                type: String
            },
            realmName: {
                type: String
            }
        },
        data: function() {
            return {
                errorMessages: [],
                paymentData: {
                    paymentReference: "",
                    bankStatementDescriptor: ""
                },
                mailToSupportEmail: 'mailto:' + this.supportEmailFormatted + '?subject=Please%20help%20me%20with%20my%20payment%20issue%20at%20' + this.realmName
            }
        },
        mounted: function() {
            const vm = this
            setTimeout(function(){
                // arbitrary delay rendering paypal form so that remainder of DOM refresh isn't blocked
                if (typeof paypal === 'undefined') {
                    return
                }
                paypal.Buttons({
                    style: {
                        shape: 'rect',
                        color: 'gold',
                        layout: 'vertical',
                        label: 'pay',
                    },
                    createOrder: vm.paypalOrderCreate,
                    onApprove: vm.paypalOrderApproved
                }).render('#paypal-button-container')
            }, 2000)
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            paypalOrderCreate: function(data, actions) {
                const vm = this
                return actions.order.create({
                    purchase_units: [{
                        amount: {
                            value: vm.entryFeeData.amount
                        }
                    }]
                });
            },
            paypalOrderApproved: function(data, actions) {
                const vm = this
                return actions.order.capture().then(vm.paypalOrderSucceeded)
            },
            paypalOrderSucceeded: function(details) {
                let paymentResult
                try {
                    paymentResult = this.getPaymentResultFromPayPalDetailsPayload(details)
                } catch (e) {
                    this.errorMessages = [
                        'Something went wrong :(',
                        e.toString(),
                        'Please try again'
                    ]
                }

                this.updateEntryPayment("paypal", paymentResult.paymentReference, paymentResult.bankStatementDescriptor)
            },
            getPaymentResultFromPayPalDetailsPayload: function(details) {
                if (details.status !== 'COMPLETED') {
                    throw `payment status: ${details.status}`
                }
                if (details.purchase_units.length === 0) {
                    throw 'no purchase units'
                }

                const processPurchaseUnit = function(purchaseUnit) {
                    let bankStatementDescriptor = purchaseUnit.soft_descriptor

                    for (let i in purchaseUnit.payments.captures) {
                        let capture = purchaseUnit.payments.captures[i]
                        let amountPaid = `${capture.amount.value} ${capture.amount.currency_code}`
                        let paymentReference = capture.id

                        if (capture.status === 'COMPLETED') {
                            return {
                                amountPaid: amountPaid,
                                paymentReference: paymentReference,
                                bankStatementDescriptor: bankStatementDescriptor
                            }
                        }
                    }

                    throw 'no purchase unit captures have COMPLETED'
                }

                for (let i in details.purchase_units) {
                    try {
                        return processPurchaseUnit(details.purchase_units[i])
                    } catch (e) {
                        continue
                    }
                }

                throw 'no succeeded purchase units found'
            },
            updateEntryPayment: function(paymentMethod, paymentReference, merchantName) {
                const vm = this
                vm.resetErrorMessages()
                vm.paymentData = {
                    paymentReference: paymentReference,
                    bankStatementDescriptor: merchantName
                }
                let url = `/api/entry/${vm.entryData.id}/payment`
                axios.request({
                    method: 'patch',
                    url: url,
                    data: {
                        payment_method: paymentMethod,
                        payment_ref: paymentReference,
                        payment_amount: vm.entryFeeData.label,
                        merchant_name: merchantName,
                        reg_token: vm.entryData.regToken
                    }
                })
                    .then(function (response) {
                        vm.$emit('update-payment-data', vm.paymentData)
                        vm.$emit('workflow-step-change', 'registrationConfirmed')
                    })
                    .catch(function (error) {
                        let response = error.response
                        switch (response.status) {
                            case 409:
                                vm.errorMessages.push(response.data.data.error)
                                break
                            case 422:
                                vm.errorMessages = response.data.data.error.reasons
                                break
                            default:
                                vm.errorMessages.push("Something went wrong :(")
                                break
                        }
                    })
            }
        }
    }
</script>
