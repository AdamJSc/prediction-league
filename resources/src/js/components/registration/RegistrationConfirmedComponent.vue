<template>
    <div class="confirmed-step-container">
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <h1>You're In!</h1>
            </div>
        </div>
        <div class="row">
            <div class="col-md-8 offset-md-2">
                <p>Use the <em>My Prediction</em> link in the menu to submit your league table prediction!</p>
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
