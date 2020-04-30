<template>
    <div>
        <h1>This is a message for {{ greetingName }}</h1>
    </div>
</template>

<script>
    export default {
        name: 'HelloWorld',
        props: {
            greetingName: {
                type: String
            }
        },
        mounted: function() {
            console.log(this.sayHello())
        },
        methods: {
            sayHello: function () {
                return "Hello " + this.greetingName + "!"
            }
        }
    }
</script>
