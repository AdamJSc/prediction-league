// load components
Vue.component("registration-workflow", require("./components/registration/RegistrationWorkflowComponent.vue").default)
Vue.component("registration-entry", require("./components/registration/RegistrationEntryComponent.vue").default)
Vue.component("registration-payment", require("./components/registration/RegistrationPaymentComponent.vue").default)
Vue.component("registration-confirmed", require("./components/registration/RegistrationConfirmedComponent.vue").default)

Vue.component("open-prediction", require("./components/prediction/OpenPredictionComponent.vue").default)
Vue.component("prediction-countdown", require("./components/prediction/PredictionCountdownComponent.vue").default)
Vue.component("prediction-login", require("./components/prediction/PredictionLoginComponent.vue").default)

Vue.component("action-button", require("./components/ActionButton.vue").default)

if (document.getElementById('app') !== null) {
    new Vue({
        el: '#app'
    })
}
