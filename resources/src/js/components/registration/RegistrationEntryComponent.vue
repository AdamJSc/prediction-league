<template>
    <div class="registration-form-container">
        <div class="payment-info">
            <div class="alert alert-warning alert-dismissible fade show" role="alert">
                <div>
                    <p><i class="fa fa-exclamation-triangle" aria-hidden="true"></i> <strong>Payment Info</strong></p>
                    <button type="button" class="close" data-dismiss="alert" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                </div>
                <p>The cost to enter the Prediction League is <strong>{{entryFeeData.label}}</strong> which covers the following:</p>
                <ul>
                    <li v-for="item in entryFeeData.breakdown">{{item}}</li>
                </ul>
                <p>On the next screen, you'll be taken to a secure <strong>PayPal</strong> integration in order to make payment.</p>
                <p>If you don't have a PayPal account, you can <strong>pay by Debit or Credit card instead</strong>.</p>
            </div>
        </div>
        <div class="row">
            <div class="col">
                <h2>Step 1/2: Your details</h2>
            </div>
        </div>
        <transition name="fade">
            <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                <ul><li v-for="msg in errorMessages">{{msg}}</li></ul>
            </div>
        </transition>
        <form id="registration-entry-form" class="form-primary">
            <div class="row">
                <div class="col-lg-6 col-md-10">
                    <h3>League details</h3>

                    <div class="form-label-definition">
                        <p class="form-label-definition-name">
                            Nickname
                            <a data-toggle="collapse" href="#nickname-description" role="button" aria-expanded="false" aria-controls="nickname-description">
                                <i class="fa fa-question-circle-o" aria-hidden="true"></i>
                            </a></p>
                        <p class="collapse form-label-definition-description" id="nickname-description">Displayed publicly. This is what you'll be known as on the leaderboard.</p>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.entrant_nickname" type="text" id="inputNickname" class="form-control" placeholder="Nickname" required>
                        <label for="inputNickname">Nickname</label>
                    </div>

                    <div class="form-label-definition">
                        <p class="form-label-definition-name">
                            League PIN
                            <a data-toggle="collapse" href="#league-pin-description" role="button" aria-expanded="false" aria-controls="league-pin-description">
                                <i class="fa fa-question-circle-o" aria-hidden="true"></i>
                            </a></p>
                        <p class="collapse form-label-definition-description" id="league-pin-description">Required. This should have been provided to you by the organiser.</p>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.pin" type="password" id="inputPIN" class="form-control" placeholder="Password" required>
                        <label for="inputPIN">PIN</label>
                    </div>
                </div>
                <div class="col-lg-6 col-md-10">
                    <h3>Contact details</h3>

                    <div class="form-label-definition">
                        <p class="form-label-definition-name">
                            Email
                            <a data-toggle="collapse" href="#email-description" role="button" aria-expanded="false" aria-controls="email-description">
                                <i class="fa fa-question-circle-o" aria-hidden="true"></i>
                            </a></p>
                        <p class="collapse form-label-definition-description" id="email-description">
                            The email address we will use to send you game-related communications such as your Short Code
                            for making a prediction, as well as round-by-round progress updates.
                            We will ONLY contact you about your direct progress within the Prediction League game.</p>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.entrant_email" type="email" id="inputEmail" name="email" class="form-control" placeholder="Email" required>
                        <label for="inputEmail">Email</label>
                    </div>

                    <div class="form-label-definition">
                        <p class="form-label-definition-name">
                            Name
                            <a data-toggle="collapse" href="#name-description" role="button" aria-expanded="false" aria-controls="name-description">
                                <i class="fa fa-question-circle-o" aria-hidden="true"></i>
                            </a></p>
                        <p class="collapse form-label-definition-description" id="name-description">The name we will use in emails we send to you. This will <strong>not</strong> be publicly displayed.</p>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.entrant_name" type="text" id="inputName" name="name" class="form-control" placeholder="Name" required>
                        <label for="inputName">Name</label>
                    </div>
                </div>
                <div class="col-lg-12 col-md-10">
                    <div class="submit-wrapper">
                        <action-button
                                label="Join"
                                @clicked="enterOnClick"
                                :is-disabled="working"
                                :is-working="working"
                                is-primary="true"></action-button>
                    </div>
                </div>
            </div>
        </form>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationForm',
        props: {
            entryFeeData: {
                type: Object
            }
        },
        data: function() {
            return {
                working: false,
                errorMessages: [],
                formData: {}
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            enterOnClick: function() {
                const vm = this
                vm.working = true
                vm.resetErrorMessages()
                axios.request({
                    method: 'post',
                    url: '/api/season/latest/entry',
                    data: this.formData
                })
                    .then(function (response) {
                        vm.$emit('update-entry-data', {
                            id: response.data.data.entry.id,
                            email: vm.formData.entrant_email,
                            shortCode: response.data.data.entry.short_code,
                            needsPayment: response.data.data.entry.needs_payment
                        })
                        vm.$el.querySelector('#registration-entry-form').reset()
                        vm.$emit('workflow-step-change', 'registrationPayment')
                    })
                    .catch(function (error) {
                        let response = error.response
                        switch (response.status) {
                            case 401:
                                vm.errorMessages.push("Please provide the correct entry PIN!")
                                break
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
                        vm.working = false
                    })
            }
        }
    }
</script>
