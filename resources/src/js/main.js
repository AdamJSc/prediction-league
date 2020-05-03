import Vue from "vue/dist/vue.esm.js"

Vue.component("hello-world", require("./components/HelloWorldComponent.vue").default)

Vue.component("registration-workflow", require("./components/registration/RegistrationWorkflowComponent.vue").default)
Vue.component("registration-form", require("./components/registration/RegistrationFormComponent.vue").default)
Vue.component("registration-payment", require("./components/registration/RegistrationPaymentComponent.vue").default)

new Vue({
    el: '#app'
})

console.log("compiled!")
