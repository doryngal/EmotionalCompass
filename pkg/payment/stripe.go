package payment

import (
	"github.com/stripe/stripe-go/v76/paymentintent"
)

func CreatePaymentIntent(amount int64, currency string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	}
	return paymentintent.New(params)
}
