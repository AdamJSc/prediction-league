<template>
    <div class="prediction-login-container">
        <div class="row">
            <div class="col-lg-6 offset-lg-3 col-md-10 offset-md-1">
                <transition name="fade">
                    <div v-if="errorMessages.length > 0" class="error-messages alert alert-block alert-danger">
                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                        <ul><li v-for="msg in errorMessages">{{msg}}</li></ul>
                    </div>
                </transition>
                <!-- TODO - feat: remove form -->
                <form v-if="showForgotShortCodeForm" class="form-primary" method="POST" action="/login/magic">
                    <div class="form-label-group">
                        <input type="text" id="inputForgotShortCodeEmailNickname" name="email_nickname" class="form-control" placeholder="Email or Nickname" required>
                        <label for="inputForgotShortCodeEmailNickname">Email or Nickname</label>
                    </div>

                    <div class="submit-wrapper">
                        <button type="submit" class="btn btn-primary">Reset my Short Code</button>
                    </div>
                </form>
                <!-- TODO - feat: amend login form -->
                <form v-else id="prediction-login-form" class="form-primary">
                    <div class="form-label-group">
                        <input v-model="formData.email_nickname" type="text" id="inputEmailNickname" name="email_nickname" class="form-control" placeholder="Email or Nickname" required>
                        <label for="inputEmailNickname">Email or Nickname</label>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.short_code" type="text" id="inputShortCode" name="short_code" class="form-control" placeholder="Short Code" required>
                        <label for="inputShortCode">Short Code</label>
                    </div>

                    <div class="submit-wrapper">
                        <action-button
                                label="Go!"
                                @clicked="loginOnClick"
                                :is-disabled="working"
                                :is-working="working"
                                :is-primary="true"></action-button>
                        <action-button
                                label="Forgot my Short Code"
                                @clicked="forgotShortCodeOnClick"
                                :is-disabled="working"
                                :is-working="working"></action-button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'PredictionLogin',
        data: function() {
            return {
                working: false,
                errorMessages: [],
                showForgotShortCodeForm: false,
                formData: {}
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            loginOnClick: function() {
                const vm = this
                vm.working = true
                vm.resetErrorMessages()
                axios.request({
                    method: 'post',
                    url: '/api/prediction/login',
                    data: this.formData
                })
                    .then(function (response) {
                        // response was successful, so auth cookie should have been set
                        // refresh the page so that this will render with our new auth token
                        window.location.reload()
                    })
                    .catch(function (error) {
                        let response = error.response
                        switch (response.status) {
                            case 401:
                                vm.errorMessages.push("Those details don't match our records. Please try again.")
                                break
                            default:
                                vm.errorMessages.push("Something went wrong :(")
                                break
                        }
                        vm.working = false
                    })
            },
            forgotShortCodeOnClick: function() {
                this.showForgotShortCodeForm = true
            }
        }
    }
</script>
