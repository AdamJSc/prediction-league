<template>
    <div class="registration-workflow-container">
        <div class="carousel slide">
            <div class="carousel-inner">
                <div class="carousel-item active">
                    <registration-form v-show="showRegistrationForm" v-on="onListeners"></registration-form>
                </div>
                <div class="carousel-item">
                    <registration-payment v-show="showRegistrationPayment" v-bind:short-code="shortCode"></registration-payment>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationWorkflowComponent',
        data: function() {
            return {
                carousel: {},
                showRegistrationSteps: {
                    registrationForm: true,
                    registrationPayment: false
                },
                shortCode: ""
            }
        },
        mounted: function() {
            this.carousel = $(this.$el.querySelector('.carousel'))
            this.carousel.carousel({interval: false})
        },
        computed: {
            onListeners: function() {
                let vm = this
                return Object.assign({},
                    this.$listeners,
                    {
                        "workflow-step-change": function(newWorkflowStep) {
                            vm.showAllRegistrationSteps()
                            vm.carousel.on('slid.bs.carousel', function () {
                                vm.showOnlyRegistrationStep(newWorkflowStep)
                            })
                            vm.carousel.carousel('next')
                        },
                        "refresh-short-code": function(newShortCode) {
                            vm.shortCode = newShortCode
                        }
                    }
                )
            },
            showRegistrationForm: function() {
                return this.showRegistrationSteps.registrationForm
            },
            showRegistrationPayment: function() {
                return this.showRegistrationSteps.registrationPayment
            }
        },
        methods: {
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