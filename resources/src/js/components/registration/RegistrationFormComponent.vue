<template>
    <div class="registration-form-container">
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <transition name="fade">
                    <div v-if="errorMessages.length > 0" class="alert alert-block alert-danger">
                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                        <ul><li v-for="msg in errorMessages">{{ msg }}</li></ul>
                    </div>
                </transition>
                <form id="registration-form" class="form-primary">
                    <div class="form-label-group">
                        <input v-model="formData.entrant_name" type="text" id="inputName" name="name" class="form-control" placeholder="Name" required autofocus>
                        <label for="inputName">Name</label>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.entrant_email" type="email" id="inputEmail" name="email" class="form-control" placeholder="Email" required>
                        <label for="inputEmail">Email</label>
                    </div>

                    <hr>

                    <div class="form-label-group">
                        <input v-model="formData.entrant_nickname" type="text" id="inputNickname" class="form-control" placeholder="Nickname" required>
                        <label for="inputNickname">Nickname</label>
                    </div>

                    <div class="form-label-group">
                        <input v-model="formData.pin" type="password" id="inputPIN" class="form-control" placeholder="Password" required>
                        <label for="inputPIN">PIN</label>
                    </div>

                    <button class="btn btn-lg btn-primary btn-block" v-on:click="submitRegistration">Enter</button>
                </form>
            </div>
        </div>
    </div>
</template>

<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationForm',
        data: function() {
            return {
                errorMessages: [],
                formData: {}
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            submitRegistration: function(e) {
                e.preventDefault()
                const vm = this
                vm.resetErrorMessages()
                axios.request({
                    method: 'post',
                    url: '/api/season/latest/entry',
                    data: this.formData
                })
                    .then(function (response) {
                        vm.$emit('refresh-entry-data', {
                            entryID: response.data.data.entry.id,
                            entryShortCode: response.data.data.entry.short_code,
                            entryNickname: response.data.data.entry.nickname,
                        })
                        vm.$el.querySelector('#registration-form').reset()
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
                    })
            }
        }
    }
</script>
