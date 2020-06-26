<template>
    <div class="confirmed-step-container">
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <h1>You're In!</h1>
            </div>
        </div>
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <p>Login using the link above to pick your league table</p>
                <ul>
                    <li>Nickname: <strong>{{entryData.entryNickname}}</strong></li>
                    <li>Short Code: <strong>{{entryData.entryShortCode}}</strong></li>
                </ul>
            </div>
        </div>
    </div>
</template>


<script>
    const axios = require('axios').default

    export default {
        name: 'RegistrationConfirmed',
        props: {
            entryData: {
                type: Object
            }
        }
    }
</script>
