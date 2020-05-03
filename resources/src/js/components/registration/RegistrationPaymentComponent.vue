<template>
    <div class="payment-step-container">
        <div class="row">
            <div class="col-md-8 offset-md-2"><h2>Make payment</h2></div>
        </div>
        <div class="row">
            <div class="col-md-8 offset-md-2">
                Your short code is {{shortCode}}
            </div>
        </div>
    </div>
</template>

<script>
    export default {
        name: 'PaymentStep',
        props: {
            shortCode: {
                type: String
            }
        }
    }
</script>
