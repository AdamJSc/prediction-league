<template>
    <div class="selection-login-container">
        <transition name="fade">
            <div v-if="errorMessages.length > 0" class="alert alert-block alert-danger">
                <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                <ul><li v-for="msg in errorMessages">{{ msg }}</li></ul>
            </div>
        </transition>
        <form id="selection-login-form" class="form-primary">
            <div class="form-label-group">
                <input v-model="formData.email_nickname" type="email" id="inputEmailNickname" name="email_nickname" class="form-control" placeholder="Email or Nickname" required autofocus>
                <label for="inputEmailNickname">Email or Nickname</label>
            </div>

            <div class="form-label-group">
                <input v-model="formData.short_code" type="text" id="inputShortCode" name="short_code" class="form-control" placeholder="Short Code" required>
                <label for="inputShortCode">Short Code</label>
            </div>

            <div class="submit-wrapper">
                <action-button
                        label="Login"
                        @clicked="loginOnClick"
                        :is-disabled="working"
                        :is-working="working"></action-button>
            </div>
        </form>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'SelectionLogin',
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
            loginOnClick: function() {
                console.log("log in now...")
            }
        }
    }
</script>
