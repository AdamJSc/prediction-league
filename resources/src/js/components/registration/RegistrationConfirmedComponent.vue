<template>
    <div class="confirmed-step-container">
        <div class="row">
            <div class="col">
                <div class="registration-success-summary">
                    <div class="success-highlight">
                        <div class="big-tick">
                            <i class="fa fa-check" aria-hidden="true"></i>
                        </div>
                        <h2>You're In!</h2>
                    </div>

                    <p>Your payment of <strong>{{entryFeeData.label}}</strong> will show on your bank statement
                        as <strong>{{paymentData.bankStatementDescriptor}}</strong></p>
                    <p>Your PayPal reference is <strong>{{paymentData.paymentReference}}</strong></p>
                    <p>We've just sent a confirmation email
                        to<br /><span class="text-highlight">{{entryData.email}}</span></p>
                    <p>This will provide instructions on how to submit your first prediction!</p>
                    <p>If you don't receive your email, please contact us
                        at <a v-bind:href="mailToSupportEmail">{{supportEmailPlainText}}</a></p>
                </div>
            </div>
        </div>
    </div>
</template>


<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationConfirmed',
        props: {
            entryData: {
                type: Object
            },
            entryFeeData: {
                type: Object
            },
            paymentData: {
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
                mailToSupportEmail: 'mailto:' + this.supportEmailFormatted + '?subject=Please%20help%20me%20with%20my%20confirmation%20email%20at%20' + this.realmName
            }
        }
    }
</script>
