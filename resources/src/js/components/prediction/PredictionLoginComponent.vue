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
                <form class="form-primary" method="POST" v-bind:action="loginPageUrl">
                  <div class="tagline">Enter your email to login</div>

                  <div class="form-label-group">
                    <input type="text" id="inputEmail" name="email_addr" class="form-control" placeholder="Email" required>
                    <label for="inputEmail">Email</label>
                  </div>

                  <div class="submit-wrapper">
                    <button type="submit" class="btn btn-primary">Get Link</button>
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
        props: {
            loginPageUrl: {
                type: String,
            }
        },
        data: function() {
            return {
                errorMessages: [],
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
        }
    }
</script>
