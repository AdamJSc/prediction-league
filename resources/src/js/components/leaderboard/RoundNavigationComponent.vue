<template>
    <table class="round-navigation-container">
        <tr>
            <td class="round-navigation text-left"><i v-on:click="$parent.$emit('decrement-round')" v-if="!isWorking && (roundNumber > 1)" class="round-navigation fa fa-chevron-left" aria-hidden="true"></i></td>
            <td class="text-center">Round {{roundNumber}}</td>
            <td class="round-navigation text-right"><i v-on:click="$parent.$emit('increment-round')" v-if="!isWorking && (roundNumber < maxRoundNumber)" class="fa fa-chevron-right" aria-hidden="true"></i></td>
        </tr>
    </table>
</template>

<script>
    export default {
        name: 'RoundNavigation',
        props: {
            isWorking: {
                type: Boolean
            },
            maxRoundNumber: {
                type: Number
            },
            roundNumber: {
                type: Number
            },
        }
    }
</script>
