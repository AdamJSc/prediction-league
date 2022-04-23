<template>
    <table class="round-navigation-container">
        <tr>
            <td class="round-navigation text-left"><i v-on:click="$emit('decrement-round')" v-if="!working && (roundNumber > 1)" class="round-navigation fa fa-chevron-left" aria-hidden="true"></i></td>
            <td class="text-center">Round {{roundNumber}}</td>
            <td class="round-navigation text-right"><i v-on:click="$emit('increment-round')" v-if="!working && (roundNumber < maxRoundNumber)" class="fa fa-chevron-right" aria-hidden="true"></i></td>
        </tr>
    </table>
</template>

<script>
    export default {
        name: 'RoundNavigation',
        props: {
            maxRoundNumber: {
                type: Number
            },
            roundNumber: {
                type: Number
            },
            working: {
                type: Boolean
            }
        }
    }
</script>
