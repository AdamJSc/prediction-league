import Vue from "vue/dist/vue.esm.js"

Vue.component("hello-world", require("./components/HelloWorldComponent.vue").default)
Vue.component("registration-form", require("./components/RegistrationFormComponent.vue").default)

new Vue({
    el: '#app'
})

console.log("compiled!")
