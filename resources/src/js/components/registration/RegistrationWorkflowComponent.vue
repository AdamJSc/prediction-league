<template>
    <div class="registration-workflow-container">
        <div class="carousel slide">
            <div class="carousel-inner">
                <div class="carousel-item active">
                    <registration-entry
                            v-show="showRegistrationSteps.registrationForm"
                            v-on:workflow-step-change="changeWorkflowStep"
                            v-on:refresh-entry-data="refreshEntryData"
                            v-bind:entry-fee-data="entryFeeData"></registration-entry>
                </div>
                <div class="carousel-item">
                    <registration-payment
                            v-show="showRegistrationSteps.registrationPayment"
                            v-on:workflow-step-change="changeWorkflowStep"
                            v-on:refresh-payment-data="refreshPaymentData"
                            v-bind:entry-data="entryData"
                            v-bind:payment-amount="entryFeeData.amount"
                            v-bind:support-email-formatted="supportEmailFormatted"
                            v-bind:support-email-plain-text="supportEmailPlainText"
                            v-bind:realm-name="realmName"></registration-payment>
                </div>
                <div class="carousel-item">
                    <registration-confirmed
                            v-show="showRegistrationSteps.registrationConfirmed"
                            v-bind:entry-data="entryData"
                            v-bind:payment-data="paymentData"></registration-confirmed>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationWorkflowComponent',
        props: {
            entryFeeAmount: {
                type: Number
            },
            entryFeeLabel: {
                type: String
            },
            rawEntryFeeBreakdown: {
                type: Array
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
                carousel: {},
                showRegistrationSteps: {
                    registrationForm: true,
                    registrationPayment: false,
                    registrationConfirmed: false
                },
                entryFeeData: {
                    amount: this.entryFeeAmount,
                    label: this.entryFeeLabel,
                    breakdown: JSON.parse(this.rawEntryFeeBreakdown)
                },
                entryData: {
                    entryID: "",
                    entryShortCode: ""
                },
                paymentData: {
                    paymentReference: "",
                    bankStatementDescriptor: ""
                }
            }
        },
        mounted: function() {
            this.carousel = $(this.$el.querySelector('.carousel'))
            this.carousel.carousel({interval: false})
        },
        methods: {
            refreshEntryData: function(newEntryData) {
                this.entryData = newEntryData
            },
            refreshPaymentData: function(newPaymentData) {
                this.paymentData = newPaymentData
            },
            changeWorkflowStep: function(newWorkflowStep) {
                const vm = this
                vm.showAllRegistrationSteps()
                vm.carousel.on('slid.bs.carousel', function () {
                    vm.showOnlyRegistrationStep(newWorkflowStep)
                })
                vm.carousel.carousel('next')
            },
            showAllRegistrationSteps: function() {
                let keys = Object.keys(this.showRegistrationSteps)
                for (let i = 0; i < keys.length; i++) {
                    this.showRegistrationSteps[keys[i]] = true
                }
            },
            showOnlyRegistrationStep: function(step) {
                let keys = Object.keys(this.showRegistrationSteps)
                for (let i = 0; i < keys.length; i++) {
                    this.showRegistrationSteps[keys[i]] = keys[i] == step
                }
            }
        }
    }
</script>