<template>
    <div class="payment-step-container">
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <h1>Make Payment</h1>
            </div>
        </div>
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <transition name="fade">
                    <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                        <ul><li v-for="msg in errorMessages">{{msg}}</li></ul>
                    </div>
                </transition>
                <div id="paypal-button-container"></div>
            </div>
        </div>
    </div>
</template>


<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationPayment',
        data: function() {
            return {
                errorMessages: []
            }
        },
        props: {
            entryData: {
                type: Object
            }
        },
        mounted: function() {
            const vm = this
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
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            paypalOrderCreate: function(data, actions) {
                return actions.order.create({
                    purchase_units: [{
                        amount: {
                            value: '0.01' // TODO - pass this in as property
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

                // TODO - emit amount paid + bank statement descriptor and store on workflow component
                this.updateEntryPayment(paymentResult.paymentReference)
            },
            getPaymentResultFromPayPalDetailsPayload: function(details) {
                if (details.status !== 'COMPLETED') {
                    throw `payment status: ${details.status}`
                }
                if (details.purchase_units.length === 0) {
                    throw 'no purchase units'
                }

                const parsePurchaseUnit = function(purchaseUnit) {
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
                        return parsePurchaseUnit(details.purchase_units[i])
                    } catch (e) {
                        continue
                    }
                }

                throw 'no succeeded purchase units found'
            },
            updateEntryPayment: function(paymentReference) {
                const vm = this
                vm.resetErrorMessages()
                let url = `/api/entry/${vm.entryData.entryID}/payment`
                axios.request({
                    method: 'patch',
                    url: url,
                    data: {
                        payment_method: "paypal",
                        payment_ref: paymentReference,
                        entry_id: vm.entryData.entryID
                    }
                })
                    .then(function (response) {
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
