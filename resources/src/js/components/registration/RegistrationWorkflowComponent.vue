<template>
    <div class="registration-workflow-container">
        <div class="carousel slide">
            <div class="carousel-inner">
                <div class="carousel-item active">
                    <registration-form v-show="showRegistrationForm" v-on:workflow-step-change="onWorkflowStepChange"></registration-form>
                </div>
                <div class="carousel-item">
                    <registration-payment v-show="showRegistrationPayment"></registration-payment>
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
                }
            }
        },
        mounted: function() {
            this.carousel = $(this.$el.querySelector('.carousel'))
            this.carousel.carousel({interval: false})
        },
        computed: {
            showRegistrationForm: function() {
                return this.showRegistrationSteps.registrationForm
            },
            showRegistrationPayment: function() {
                return this.showRegistrationSteps.registrationPayment
            }
        },
        methods: {
            onWorkflowStepChange: function(newWorkflowStep) {
                let vm = this
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