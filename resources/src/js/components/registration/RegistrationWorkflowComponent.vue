<template>
    <div class="registration-workflow-container">
        <countdown label="Entries close in..." v-bind:unix="unix"></countdown>
        <div class="carousel slide">
            <div class="carousel-inner">
                <div class="carousel-item active">
                    <registration-entry
                            v-show="showRegistrationSteps.registrationForm"
                            v-on:workflow-step-change="changeWorkflowStep"
                            v-on:update-entry-data="updateEntryData"
                            v-bind:entry-fee-data="entryFeeData"></registration-entry>
                </div>
                <div class="carousel-item">
                    <registration-payment
                            v-show="showRegistrationSteps.registrationPayment"
                            v-on:workflow-step-change="changeWorkflowStep"
                            v-on:update-payment-data="updatePaymentData"
                            v-bind:entry-data="entryData"
                            v-bind:entry-fee-data="entryFeeData"
                            v-bind:support-email-formatted="supportEmailFormatted"
                            v-bind:support-email-plain-text="supportEmailPlainText"
                            v-bind:realm-name="realmName"></registration-payment>
                </div>
                <div class="carousel-item">
                    <registration-confirmed
                            v-show="showRegistrationSteps.registrationConfirmed"
                            v-bind:entry-data="entryData"
                            v-bind:entry-fee-data="entryFeeData"
                            v-bind:payment-data="paymentData"
                            v-bind:support-email-formatted="supportEmailFormatted"
                            v-bind:support-email-plain-text="supportEmailPlainText"
                            v-bind:realm-name="realmName"></registration-confirmed>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationWorkflowComponent',
        props: {
            unix: {
                type: String
            },
            entryFeeAmount: {
                type: Number
            },
            entryFeeLabel: {
                type: String
            },
            rawEntryFeeBreakdown: {
                type: String
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
                    id: "",
                    email: "",
                    regToken: "",
                    needsPayment: true
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
            updateEntryData: function(newEntryData) {
                this.entryData = newEntryData
            },
            updatePaymentData: function(newPaymentData) {
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