<template>
    <div class="registration-form-container">
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
              <div class="row">

                <div class="col-lg-12">
                  <div class="form-label-definition">
                    <p class="form-label-definition-name">
                      Email
                      <a class="registration-form-toggle" data-toggle="collapse" href="#email-description" role="button" aria-expanded="false" aria-controls="email-description">
                        <i class="fa-solid fa-circle-question"></i>
                      </a></p>
                    <p class="collapse form-label-definition-description" id="email-description">
                      The email address we will use to send you game-related communications.
                      We will only contact you regarding your participation within the game.</p>
                  </div>
                  <div class="form-label-group">
                    <input v-model="formData.entrant_email" type="email" id="inputEmail" name="email" class="form-control" placeholder="Email" required>
                    <label for="inputEmail">Email</label>
                  </div>
                </div>

                <div class="col-lg-12">
                  <div class="form-label-definition">
                    <p class="form-label-definition-name">
                      Name
                      <a class="registration-form-toggle" data-toggle="collapse" href="#name-description" role="button" aria-expanded="false" aria-controls="name-description">
                        <i class="fa-solid fa-circle-question"></i>
                      </a></p>
                    <p class="collapse form-label-definition-description" id="name-description">The name we will use in emails we send to you. This will <strong>not</strong> be publicly displayed.</p>
                  </div>

                  <div class="form-label-group">
                    <input v-model="formData.entrant_name" type="text" id="inputName" name="name" class="form-control" placeholder="Name" required>
                    <label for="inputName">Name</label>
                  </div>
                </div>

                <div class="col-lg-12">
                  <div class="form-label-definition">
                    <p class="form-label-definition-name">
                      Nickname
                      <a class="registration-form-toggle" data-toggle="collapse" href="#nickname-description" role="button" aria-expanded="false" aria-controls="nickname-description">
                        <i class="fa-solid fa-circle-question"></i>
                      </a></p>
                    <p class="collapse form-label-definition-description" id="nickname-description">Displayed publicly. This is what you'll be known as on the leaderboard.</p>
                  </div>
                  <div class="form-label-group">
                    <input v-model="formData.entrant_nickname" type="text" id="inputNickname" class="form-control" placeholder="Nickname" required>
                    <label for="inputNickname">Nickname</label>
                  </div>
                </div>

              </div>

            </div>

            <div class="col-lg-6 col-md-10">
              <div class="row">

                <div class="col-lg-12">
                  <div class="payment-info">
                    <p><strong>Payment Info</strong></p>
                    <p>Entry costs <strong>{{entryFeeData.label}}</strong> and includes:</p>
                    <ul>
                      <li v-for="item in entryFeeData.breakdown">{{item}}</li>
                    </ul>
                    <p>On the next screen, you'll be taken to a secure <strong>PayPal</strong> integration
                      to make payment by debit/credit card.</p>
                  </div>
                </div>

              </div>
            </div>
          </div>

          <div class="row">

            <div class="col-lg-12">
              <div class="submit-wrapper">
                <action-button
                    label="Next"
                    @clicked="enterOnClick"
                    :is-disabled="working"
                    :is-working="working"
                    :is-primary="true"></action-button>
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
            },
            realmPin: {
                type: String
            }
        },
        data: function() {
            return {
                working: false,
                errorMessages: [],
                formData: {
                    pin: this.realmPin
                }
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
                        let body = response.data
                        vm.$emit('update-entry-data', {
                            id: body.data.entry.id,
                            email: vm.formData.entrant_email,
                            regToken: body.data.entry.reg_token,
                            needsPayment: body.data.entry.needs_payment
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
