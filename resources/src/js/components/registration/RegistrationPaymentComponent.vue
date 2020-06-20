<template>
    <div class="payment-step-container">
        <div class="row">
            <div class="col-md-8 offset-md-2"><h2>Make payment</h2></div>
        </div>
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <transition name="fade">
                    <div v-if="errorMessages.length > 0" class="alert alert-block alert-danger">
                        <button type="button" class="close" v-on:click="resetErrorMessages">&times;</button>
                        <ul><li v-for="msg in errorMessages">{{ msg }}</li></ul>
                    </div>
                </transition>
                <form id="registration-payment-form" class="form-primary">
                    <action-button
                            label="Pay Later"
                            @clicked="submitPayment"
                            :is-disabled="working"
                            :is-working="working"></action-button>
                </form>
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
                working: false,
                errorMessages: []
            }
        },
        props: {
            entryData: {
                type: Object
            }
        },
        computed: {
            formData: function() {
                return {
                    payment_method: "other",
                    payment_ref: "bank_transfer",
                    entry_id: this.entryData.entryID
                }
            }
        },
        methods: {
            resetErrorMessages: function() {
                this.errorMessages = []
            },
            submitPayment: function(e) {
                const vm = this
                vm.working = true
                vm.resetErrorMessages()
                let url = `/api/entry/${vm.entryData.entryID}/payment`
                axios.request({
                    method: 'patch',
                    url: url,
                    data: this.formData
                })
                    .then(function (response) {
                        // redirect to team selection
                        window.location = `${location.origin}/${encodeURI(vm.entryData.entryNickname)}/${vm.entryData.entryShortCode}/teams`
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
                        vm.working = false
                    })
            }
        }
    }
</script>
