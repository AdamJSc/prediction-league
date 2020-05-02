<template>
    <div class="registration-workflow-container">
        <registration-form v-show="displayWorkflowStep=='registration-form'" v-on:workflow-step-change="onWorkflowStepChange" ></registration-form>
        <registration-payment v-show="displayWorkflowStep=='registration-payment'"></registration-payment>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationWorkflowComponent',
        data: function() {
            return {
                displayWorkflowStep: 'registration-form',
            }
        },
        methods: {
            onWorkflowStepChange: function(newWorkflowStep) {
                this.displayWorkflowStep = newWorkflowStep
            }
        }
    }
</script>