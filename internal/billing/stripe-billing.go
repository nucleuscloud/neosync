package billing

import (
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v79"
	stripeapiclient "github.com/stripe/stripe-go/v79/client"
	"github.com/stripe/stripe-go/v79/subscription"
)

type Interface interface {
	NewCustomer(req *CustomerRequest) (*stripe.Customer, error)
	NewBillingPortalSession(customerId, accountSlug string) (*stripe.BillingPortalSession, error)
	NewCheckoutSession(customerId, accountSlug, userId string) (*stripe.CheckoutSession, error)
	GetSubscriptions(customerId string) *subscription.Iter
}

type Client struct {
	client *stripeapiclient.API
	cfg    *Config
}

var _ Interface = (*Client)(nil)

type Config struct {
	AppBaseUrl     string
	PriceLookupKey string
}

func New(
	client *stripeapiclient.API,
	cfg *Config,
) *Client {
	return &Client{client: client, cfg: cfg}
}

type CustomerRequest struct {
	Email     string
	Name      string
	AccountId string
	UserId    string
}

func (c *Client) GetSubscriptions(customerId string) *subscription.Iter {
	return c.client.Subscriptions.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerId),
	})
}

func (c *Client) NewCustomer(req *CustomerRequest) (*stripe.Customer, error) {
	return c.client.Customers.New(&stripe.CustomerParams{
		Email: stripe.String(req.Email),
		Name:  stripe.String(req.Name),
		Metadata: map[string]string{
			"accountId":   req.AccountId,
			"createdById": req.UserId,
		},
	})
}

func (c *Client) NewBillingPortalSession(customerId, accountSlug string) (*stripe.BillingPortalSession, error) {
	return c.client.BillingPortalSessions.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerId),
		ReturnURL: stripe.String(fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug)),
	})
}

func (c *Client) NewCheckoutSession(customerId, accountSlug, userId string) (*stripe.CheckoutSession, error) {
	price, err := c.getPriceFromLookupKey()
	if err != nil {
		return nil, err
	}

	return c.client.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price.ID),
				Quantity: stripe.Int64(1), // todo: remove this once we set up metering
			},
		},
		SuccessURL: stripe.String(fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug)),
		Customer:   stripe.String(customerId),
		Metadata:   map[string]string{"userId": userId},
	})
}

func (c *Client) getPriceFromLookupKey() (*stripe.Price, error) {
	pricelistParams := &stripe.PriceListParams{
		LookupKeys: stripe.StringSlice([]string{c.cfg.PriceLookupKey}),
		Active:     stripe.Bool(true),
	}
	iter := c.client.Prices.List(pricelistParams)
	var price *stripe.Price
	for iter.Next() {
		p := iter.Price()
		price = p
		break
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	if price == nil {
		return nil, errors.New("unable to find price during checkout session lookup")
	}
	return price, nil
}
