<template>
    <button v-on:click="buttonClick" v-bind:class="{'btn': true, 'btn-primary': isPrimary, 'btn-secondary': '!isPrimary'}" v-bind:disabled="isDisabled">
        <span v-if="isWorking"><img alt="loader" src="/assets/img/loader-dark-bg.svg" /></span>
        <span v-else>{{label}}</span>
    </button>
</template>

<script>
    export default {
        name: 'ActionButton',
        props: {
            label: {
                type: String
            },
            isDisabled: {
                type: Boolean
            },
            isWorking: {
                type: Boolean
            },
            isPrimary: {
                type: Boolean
            }
        },
        methods: {
            buttonClick: function(e) {
                e.preventDefault()
                this.$emit('clicked')
            }
        }
    }
</script>
