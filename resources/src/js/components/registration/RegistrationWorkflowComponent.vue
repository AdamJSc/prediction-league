<template>
    <div class="registration-workflow-container">
        <div class="carousel slide">
            <div class="carousel-inner">
                <div class="carousel-item active">
                    <registration-entry v-show="showRegistrationSteps.registrationForm" v-on:workflow-step-change="changeWorkflowStep" v-on:refresh-entry-data="refreshEntryData"></registration-entry>
                </div>
                <div class="carousel-item">
                    <registration-payment v-show="showRegistrationSteps.registrationPayment" v-on:workflow-step-change="changeWorkflowStep" v-bind:entry-data="entryData"></registration-payment>
                </div>
                <div class="carousel-item">
                    <registration-confirmed v-show="showRegistrationSteps.registrationConfirmed" v-bind:entry-data="entryData"></registration-confirmed>
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
                    registrationPayment: false,
                    registrationConfirmed: false
                },
                entryData: {entryID: "", entryShortCode: ""}
            }
        },
        mounted: function() {
            this.carousel = $(this.$el.querySelector('.carousel'))
            this.carousel.carousel({interval: false})
        },
        methods: {
            refreshEntryData: function(newEntryData) {
                this.entryData = newEntryData
            },
            changeWorkflowStep: function(newWorkflowStep) {
                const vm = this
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