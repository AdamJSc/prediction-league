<template>
    <div class="form-container">
        <form class="form-register">
            <div class="form-label-group">
                <input type="text" id="inputName" name="name" class="form-control" placeholder="Name" required autofocus>
                <label for="inputName">Name</label>
            </div>

            <div class="form-label-group">
                <input type="email" id="inputEmail" name="email" class="form-control" placeholder="Email" required>
                <label for="inputEmail">Email</label>
            </div>

            <hr>

            <div class="form-label-group">
                <input type="text" id="inputNickname" class="form-control" placeholder="Nickname" required>
                <label for="inputNickname">Nickname</label>
            </div>

            <div class="form-label-group">
                <input type="password" id="inputPIN" class="form-control" placeholder="Password" required>
                <label for="inputPIN">PIN</label>
            </div>

            <button class="btn btn-lg btn-primary btn-block" type="submit" v-on:click="submitRegistration">Enter</button>
        </form>
    </div>
</template>

<script>
    export default {
        name: 'RegistrationForm',
        methods: {
            submitRegistration: function (e) {
                e.preventDefault()
                console.log(e)
                console.log("form submission!")
            }
        }
    }
</script>
