package go_printify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	getShopOrdersPath         = "shops/%d/orders.json"
	getShopOrderPath          = "shops/%d/orders/%d.json"
	sendOrderToProductionPath = "shops/%d/orders/%d/send_to_production.json"
	getShippingCostsPath      = "shops/%d/orders/shipping.json"
	cancelOrderPath           = "shops/%d/orders/%d/cancel.json"
)

type Page struct {
	CurrentPage  int         `json:"current_page"`
	Data         interface{} `json:"data"`
	FirstPageUrl string      `json:"first_page_url"`
	From         int         `json:"from"`
	LastPage     int         `json:"last_page"`
	LastPageUrl  string      `json:"last_page_url"`
	Links        []*Link     `json:"links"`
	NextPageUrl  string      `json:"next_page_url"`
	Path         string      `json:"path"`
	PerPage      int         `json:"per_page"`
	PrevPageUrl  string      `json:"prev_page_url"`
	To           int         `json:"to"`
	Total        int         `json:"total"`
}

type Link struct {
	Url    string `json:"url"`
	Label  string `json:"label"`
	Active bool   `json:"active"`
}
type Order struct {
	Id                       *string            `json:"id,omitempty"`
	AddressTo                *map[string]string `json:"address_to,omitempty"`
	LineItems                []*LineItem        `json:"line_items"`
	Metadata                 *OrderMetadata     `json:"metadata,omitempty"`
	TotalPrice               *float32           `json:"total_price,omitempty"`
	TotalShipping            *float32           `json:"total_shipping,omitempty"`
	TotalTax                 *float32           `json:"total_tax,omitempty"`
	Status                   *string            `json:"status,omitempty"`
	ShippingMethod           int                `json:"shipping_method"`
	SendShippingNotification *bool              `json:"send_shipping_notification"`
	Shipments                []*Shipment        `json:"shipments,omitempty"`
	CreatedAt                *time.Time         `json:"created_at,omitempty"`
	SentToProductionAt       *time.Time         `json:"sent_to_production_at,omitempty"`
	FulfilledAt              *time.Time         `json:"fulfilled_at,omitempty"`
}

type LineItem struct {
	Id                 *int               `json:"id,omitempty"`
	VariantId          *int               `json:"variant_id,omitempty"`
	ProductId          *string            `json:"product_id,omitempty"`
	BlueprintId        *int               `json:"blueprint_id,omitempty"`
	Quantity           int                `json:"quantity"`
	PrintProviderId    *int               `json:"print_provider_id,omitempty"`
	PrintAreas         *map[string]string `json:"print_areas,omitempty"`
	PrintDetails       *PrintDetails      `json:"print_details,omitempty"`
	Cost               *float32           `json:"cost,omitempty"`
	Sku                *string            `json:"sku,omitempty"`
	ShippingCost       *float32           `json:"shipping_cost,omitempty"`
	Status             *string            `json:"status,omitempty"`
	Metadata           *LineItemMetadata  `json:"metadata,omitempty"`
	SentToProductionAt *time.Time         `json:"sent_to_production_at,omitempty"`
	FulfilledAt        *time.Time         `json:"fulfilled_at,omitempty"`
}

type OrderMetadata struct {
	OrderType       string    `json:"order_type"`
	ShopOrderId     int       `json:"shop_order_id"`
	ShopOrderLabel  string    `json:"shop_order_label"`
	ShopFulfilledAt time.Time `json:"shop_fulfilled_at"`
}

type LineItemMetadata struct {
	Title        string  `json:"title"`
	Price        float32 `json:"price"`
	VariantLabel string  `json:"variant_label"`
	Sku          string  `json:"sku"`
	Country      string  `json:"country"`
}

type Shipment struct {
	Carrier     string    `json:"carrier"`
	Number      string    `json:"number"`
	Url         string    `json:"url"`
	DeliveredAt time.Time `json:"delivered_at"`
}

type ShippingCost struct {
	Standard float32 `json:"standard"`
	Express  float32 `json:"express"`
}

/*
Retrieve a list of orders
*/
func (c *Client) ListShopOrders(shopId int, page, limit *int, statusFilter *string) ([]*Order, error) {
	path := fmt.Sprintf(getShopOrdersPath, shopId)
	if page != nil || limit != nil || statusFilter != nil {
		path = fmt.Sprintf("%s?", path)
	}
	if page != nil {
		path = fmt.Sprintf("%spage=%d", path, *page)
	}

	if limit != nil {
		path = fmt.Sprintf("%s&limit=%d", path, *limit)
	}

	if statusFilter != nil {
		path = fmt.Sprintf("%s&status=%s", path, *statusFilter)
	}

	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	orderList := make([]*Order, 0)
	_, err = c.do(req, &orderList)
	return orderList, err
}

/*
Get order details by ID
*/
func (c *Client) GetOrderDetails(shopId, orderId int) (*Order, error) {
	path := fmt.Sprintf(getShopOrderPath, shopId, orderId)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	order := &Order{}
	_, err = c.do(req, order)
	return order, err
}

/*
SubmitOrder to printify API
returns Printify order ID
*/
func (c *Client) SubmitOrder(shopId int, order *Order) (*string, error) {
	path := fmt.Sprintf(getShopOrdersPath, shopId)
	req, err := c.newRequest(http.MethodPost, path, order)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req, order)

	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		// {
		// 	"id": "5a96f649b2439217d070f507"
		// }

		var responseData OrderSubmitResponse
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			return nil, err
		}
		return responseData.ID, nil

	}
	return nil, err
}

type OrderSubmitResponse struct {
	ID *string `json:"id"`
}

/*
Send an existing order to production
*/
func (c *Client) SendOrderToProduction(shopId, orderId int) (*Order, error) {
	path := fmt.Sprintf(sendOrderToProductionPath, shopId, orderId)
	req, err := c.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	order := &Order{}
	_, err = c.do(req, order)
	return order, err
}

/*
Calculate the shipping cost of an order
*/
func (c *Client) CalculateShippingCosts(shopId int, order *Order) (*ShippingCost, error) {
	path := fmt.Sprintf(getShippingCostsPath, shopId)
	req, err := c.newRequest(http.MethodPost, path, order)
	if err != nil {
		return nil, err
	}
	shippingCost := &ShippingCost{}
	_, err = c.do(req, shippingCost)
	return shippingCost, err
}

/*
Cancel an order
*/
func (c *Client) CancelOrder(shopId, orderId int) (*Order, error) {
	path := fmt.Sprintf(cancelOrderPath, shopId, orderId)
	req, err := c.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	order := &Order{}
	_, err = c.do(req, order)
	return order, err
}
