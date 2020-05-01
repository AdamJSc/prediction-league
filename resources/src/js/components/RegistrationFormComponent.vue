<template>
    <div class="form-container">
        <form id="registration-form">
            <div>
                <div>Name</div>
                <div><input type="text" name="name"/></div>
            </div>
            <div>
                <div>Nickname</div>
                <div><input type="text" name="nickname"/></div>
            </div>
            <div>
                <div>Email</div>
                <div><input type="email" name="email"/></div>
            </div>
            <div>
                <div>League PIN</div>
                <div><input type="password" name="pin"/></div>
            </div>
            <div>
                <div><button type="button" name="enter" v-on:click="submitForm()">Enter!</button></div>
            </div>
        </form>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationForm',
        methods: {
            submitForm: function () {
                console.log("form submission!")
            }
        }
    }
</script>
