// load components
Vue.component("registration-workflow", require("./components/registration/RegistrationWorkflowComponent.vue").default)
Vue.component("registration-entry", require("./components/registration/RegistrationEntryComponent.vue").default)
Vue.component("registration-payment", require("./components/registration/RegistrationPaymentComponent.vue").default)
Vue.component("registration-confirmed", require("./components/registration/RegistrationConfirmedComponent.vue").default)

Vue.component("open-selection", require("./components/selection/OpenSelectionComponent.vue").default)
Vue.component("selection-countdown", require("./components/selection/SelectionCountdownComponent.vue").default)
Vue.component("selection-login", require("./components/selection/SelectionLoginComponent.vue").default)

Vue.component("action-button", require("./components/ActionButton.vue").default)

if (document.getElementById('app') !== null) {
    new Vue({
        el: '#app'
    })
}
