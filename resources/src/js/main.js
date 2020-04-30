import Vue from "vue/dist/vue.esm.js"

Vue.component("hello-world", require("./components/HelloWorldComponent.vue").default)

new Vue({
    el: '#app'
})

console.log("compiled!")
