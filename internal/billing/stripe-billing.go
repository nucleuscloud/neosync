package billing

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/stripe/stripe-go/v81"
	stripeapiclient "github.com/stripe/stripe-go/v81/client"
)

type SubscriptionIter interface {
	Subscription() *stripe.Subscription
	Next() bool
	Err() error
}

type Interface interface {
	NewCustomer(req *CustomerRequest) (*stripe.Customer, error)
	NewBillingPortalSession(customerId, accountSlug string) (*stripe.BillingPortalSession, error)
	NewCheckoutSession(
		customerId, accountSlug, userId string,
		logger *slog.Logger,
	) (*stripe.CheckoutSession, error)
	GetSubscriptions(customerId string) SubscriptionIter
	NewMeterEvent(req *MeterEventRequest) (*stripe.BillingMeterEvent, error)
}

type Client struct {
	client *stripeapiclient.API
	cfg    *Config
}

type PriceQuantity map[string]int

var _ Interface = (*Client)(nil)

type Config struct {
	AppBaseUrl   string
	PriceLookups PriceQuantity
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

func (c *Client) GetSubscriptions(customerId string) SubscriptionIter {
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

func (c *Client) NewBillingPortalSession(
	customerId, accountSlug string,
) (*stripe.BillingPortalSession, error) {
	return c.client.BillingPortalSessions.New(&stripe.BillingPortalSessionParams{
		Customer: stripe.String(customerId),
		ReturnURL: stripe.String(
			fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug),
		),
	})
}

func (c *Client) NewCheckoutSession(
	customerId, accountSlug, userId string,
	logger *slog.Logger,
) (*stripe.CheckoutSession, error) {
	priceMap, err := c.getPricesFromLookupKeys()
	if err != nil {
		return nil, err
	}

	lineitems := []*stripe.CheckoutSessionLineItemParams{}
	for lookup, quantity := range c.cfg.PriceLookups {
		price, ok := priceMap[lookup]
		if !ok {
			return nil, fmt.Errorf("unable to find stripe price for lookup key: %s", lookup)
		}
		lineitem := &stripe.CheckoutSessionLineItemParams{
			Price: stripe.String(price.ID),
		}
		if quantity > 0 {
			lineitem.Quantity = stripe.Int64(int64(quantity))
		}
		lineitems = append(lineitems, lineitem)
	}
	logger.Debug("creating stripe checkout session", "numLineItems", len(lineitems))
	return c.client.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		Mode:      stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: lineitems,
		SuccessURL: stripe.String(
			fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug),
		),
		CancelURL: stripe.String(
			fmt.Sprintf("%s/%s/settings/billing", c.cfg.AppBaseUrl, accountSlug),
		),
		Customer: stripe.String(customerId),
		Metadata: map[string]string{"userId": userId},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			BillingCycleAnchor: stripe.Int64(getNextMonthBillingCycleAnchor(time.Now().UTC())),
		},
	})
}

type MeterEventRequest struct {
	EventName  string
	Identifier string
	Timestamp  *int64
	CustomerId string
	Value      string
}

func (c *Client) NewMeterEvent(req *MeterEventRequest) (*stripe.BillingMeterEvent, error) {
	return c.client.BillingMeterEvents.New(&stripe.BillingMeterEventParams{
		EventName:  stripe.String(req.EventName),
		Identifier: stripe.String(req.Identifier),
		Timestamp:  req.Timestamp,
		Payload: map[string]string{
			"stripe_customer_id": req.CustomerId,
			"value":              req.Value,
		},
	})
}

func getNextMonthBillingCycleAnchor(date time.Time) int64 {
	dateUtc := date.UTC()
	// First set to the 1st of the current month
	firstOfMonth := time.Date(dateUtc.Year(), dateUtc.Month(), 1, 0, 0, 0, 0, dateUtc.Location())
	// Then add one month - this avoids day rollover issues
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
	return firstOfNextMonth.Unix()
}

func (c *Client) getPricesFromLookupKeys() (map[string]*stripe.Price, error) {
	output := map[string]*stripe.Price{}
	pricelistParams := &stripe.PriceListParams{
		LookupKeys: stripe.StringSlice(toLookupKeySlice(c.cfg.PriceLookups)),
		Active:     stripe.Bool(true),
	}
	iter := c.client.Prices.List(pricelistParams)
	for iter.Next() {
		p := iter.Price()
		if _, ok := c.cfg.PriceLookups[p.LookupKey]; ok {
			output[p.LookupKey] = p
		}
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	if len(output) != len(c.cfg.PriceLookups) {
		return nil, fmt.Errorf(
			"unable to resolve all stripe price lookups to valid prices. need %d, found %d",
			len(c.cfg.PriceLookups),
			len(output),
		)
	}
	return output, nil
}

func toLookupKeySlice(pc PriceQuantity) []string {
	output := []string{}

	for k := range pc {
		output = append(output, k)
	}
	return output
}
