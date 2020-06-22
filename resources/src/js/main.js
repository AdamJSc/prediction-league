import Vue from "vue/dist/vue.esm.js"

// load components
Vue.component("registration-workflow", require("./components/registration/RegistrationWorkflowComponent.vue").default)
Vue.component("registration-entry", require("./components/registration/RegistrationEntryComponent.vue").default)
Vue.component("registration-payment", require("./components/registration/RegistrationPaymentComponent.vue").default)
Vue.component("registration-confirmed", require("./components/registration/RegistrationConfirmedComponent.vue").default)

Vue.component("no-more-selections", require("./components/selection/NoMoreSelectionsComponent.vue").default)
Vue.component("next-selection", require("./components/selection/NextSelectionComponent.vue").default)
Vue.component("open-selection", require("./components/selection/OpenSelectionComponent.vue").default)

Vue.component("action-button", require("./components/ActionButton.vue").default)

if (document.getElementById('app') !== null) {
    new Vue({
        el: '#app'
    })
}

console.log("compiled!")
