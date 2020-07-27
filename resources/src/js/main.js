// load components
Vue.component("leaderboard", require("./components/leaderboard/LeaderboardComponent.vue").default)
Vue.component("scored-entry-prediction", require("./components/leaderboard/ScoredEntryPredictionComponent.vue").default)

Vue.component("registration-workflow", require("./components/registration/RegistrationWorkflowComponent.vue").default)
Vue.component("registration-entry", require("./components/registration/RegistrationEntryComponent.vue").default)
Vue.component("registration-payment", require("./components/registration/RegistrationPaymentComponent.vue").default)
Vue.component("registration-confirmed", require("./components/registration/RegistrationConfirmedComponent.vue").default)

Vue.component("open-prediction", require("./components/prediction/OpenPredictionComponent.vue").default)
Vue.component("prediction-countdown", require("./components/prediction/PredictionCountdownComponent.vue").default)
Vue.component("prediction-login", require("./components/prediction/PredictionLoginComponent.vue").default)

Vue.component("action-button", require("./components/ActionButton.vue").default)
Vue.component("countdown", require("./components/CountdownComponent.vue").default)

if (document.getElementById('app') !== null) {
    new Vue({
        el: '#app'
    })
}

const logout = document.getElementById('logout')
if (logout !== null) {
    logout.addEventListener('click', function(){ $('#logoutModal').modal('show'); console.log('yepp!') })
}

const logoutAction = document.getElementById('logout-action')
if (logoutAction !== null) {
    logoutAction.addEventListener('click', function(){
        // reset cookie value
        document.cookie = 'PL_AUTH=;expires=1970-01-01T00:00:00Z;path=/'
        window.location = '/'
    })
}
