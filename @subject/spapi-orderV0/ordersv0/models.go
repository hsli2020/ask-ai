package ordersv0

// Money represents a monetary value.
type Money struct {
	CurrencyCode string `json:"CurrencyCode,omitempty"`
	Amount       string `json:"Amount,omitempty"`
}

// Address represents a shipping address.
type Address struct {
	Name          string `json:"Name"`
	AddressLine1  string `json:"AddressLine1,omitempty"`
	AddressLine2  string `json:"AddressLine2,omitempty"`
	AddressLine3  string `json:"AddressLine3,omitempty"`
	City          string `json:"City,omitempty"`
	County        string `json:"County,omitempty"`
	District      string `json:"District,omitempty"`
	StateOrRegion string `json:"StateOrRegion,omitempty"`
	PostalCode    string `json:"PostalCode,omitempty"`
	CountryCode   string `json:"CountryCode,omitempty"`
	Phone         string `json:"Phone,omitempty"`
	AddressType   string `json:"AddressType,omitempty"`
}

// BuyerTaxInfo represents tax information for a buyer.
type BuyerTaxInfo struct {
	CompanyLegalName  string `json:"CompanyLegalName,omitempty"`
	TaxingRegion      string `json:"TaxingRegion,omitempty"`
	TaxClassifications []struct {
		Name  string `json:"Name,omitempty"`
		Value string `json:"Value,omitempty"`
	} `json:"TaxClassifications,omitempty"`
}

// BuyerInfo represents information about the buyer.
type BuyerInfo struct {
	BuyerEmail        string       `json:"BuyerEmail,omitempty"`
	BuyerName         string       `json:"BuyerName,omitempty"`
	BuyerCounty       string       `json:"BuyerCounty,omitempty"`
	BuyerTaxInfo      *BuyerTaxInfo `json:"BuyerTaxInfo,omitempty"`
	PurchaseOrderNumber string     `json:"PurchaseOrderNumber,omitempty"`
}

// Order represents an order.
type Order struct {
	AmazonOrderID                string       `json:"AmazonOrderId"`
	SellerOrderID                string       `json:"SellerOrderId,omitempty"`
	PurchaseDate                 string       `json:"PurchaseDate"`
	LastUpdateDate               string       `json:"LastUpdateDate"`
	OrderStatus                  string       `json:"OrderStatus"`
	FulfillmentChannel           string       `json:"FulfillmentChannel,omitempty"`
	SalesChannel                 string       `json:"SalesChannel,omitempty"`
	OrderChannel                 string       `json:"OrderChannel,omitempty"`
	ShipServiceLevel             string       `json:"ShipServiceLevel,omitempty"`
	OrderTotal                   *Money       `json:"OrderTotal,omitempty"`
	NumberOfItemsShipped         int          `json:"NumberOfItemsShipped,omitempty"`
	NumberOfItemsUnshipped       int          `json:"NumberOfItemsUnshipped,omitempty"`
	PaymentMethod                string       `json:"PaymentMethod,omitempty"`
	PaymentMethodDetails         []string     `json:"PaymentMethodDetails,omitempty"`
	MarketplaceID                string       `json:"MarketplaceId,omitempty"`
	ShipmentServiceLevelCategory string       `json:"ShipmentServiceLevelCategory,omitempty"`
	EasyShipShipmentStatus       string       `json:"EasyShipShipmentStatus,omitempty"`
	CbaDisplayableShippingLabel  string       `json:"CbaDisplayableShippingLabel,omitempty"`
	OrderType                    string       `json:"OrderType,omitempty"`
	EarliestShipDate             string       `json:"EarliestShipDate,omitempty"`
	LatestShipDate               string       `json:"LatestShipDate,omitempty"`
	EarliestDeliveryDate         string       `json:"EarliestDeliveryDate,omitempty"`
	LatestDeliveryDate           string       `json:"LatestDeliveryDate,omitempty"`
	IsBusinessOrder              bool         `json:"IsBusinessOrder,omitempty"`
	IsPrime                      bool         `json:"IsPrime,omitempty"`
	IsPremiumOrder               bool         `json:"IsPremiumOrder,omitempty"`
	IsGlobalExpressEnabled       bool         `json:"IsGlobalExpressEnabled,omitempty"`
	IsReplacementOrder           bool         `json:"IsReplacementOrder,omitempty"`
	IsSoldByAB                   bool         `json:"IsSoldByAB,omitempty"`
	IsIBA                        bool         `json:"IsIBA,omitempty"`
	IsAccessPointOrder           bool         `json:"IsAccessPointOrder,omitempty"`
	IsISPU                       bool         `json:"IsISPU,omitempty"`
	ShippingAddress              *Address     `json:"ShippingAddress,omitempty"`
	BuyerInfo                    *BuyerInfo   `json:"BuyerInfo,omitempty"`
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ASIN              string `json:"ASIN"`
	SellerSKU         string `json:"SellerSKU,omitempty"`
	OrderItemID       string `json:"OrderItemId"`
	Title             string `json:"Title,omitempty"`
	QuantityOrdered   int    `json:"QuantityOrdered"`
	QuantityShipped   int    `json:"QuantityShipped,omitempty"`
	ItemPrice         *Money `json:"ItemPrice,omitempty"`
	ShippingPrice     *Money `json:"ShippingPrice,omitempty"`
	GiftWrapPrice     *Money `json:"GiftWrapPrice,omitempty"`
	ItemTax           *Money `json:"ItemTax,omitempty"`
	ShippingTax       *Money `json:"ShippingTax,omitempty"`
	GiftWrapTax       *Money `json:"GiftWrapTax,omitempty"`
	ShippingDiscount  *Money `json:"ShippingDiscount,omitempty"`
	PromotionDiscount *Money `json:"PromotionDiscount,omitempty"`
	PromotionIDs      []string `json:"PromotionIds,omitempty"`
	CODFee            *Money `json:"CODFee,omitempty"`
	CODFeeDiscount    *Money `json:"CODFeeDiscount,omitempty"`
	IsGift            bool   `json:"IsGift,omitempty"`
	ConditionNote     string `json:"ConditionNote,omitempty"`
	ConditionID       string `json:"ConditionId,omitempty"`
	ConditionSubtypeID string `json:"ConditionSubtypeId,omitempty"`
	ScheduledDeliveryStartDate string `json:"ScheduledDeliveryStartDate,omitempty"`
	ScheduledDeliveryEndDate string `json:"ScheduledDeliveryEndDate,omitempty"`
	PriceDesignation  string `json:"PriceDesignation,omitempty"`
}

// GetOrdersResponse is the response schema for the getOrders operation.
type GetOrdersResponse struct {
	Payload struct {
		Orders      []Order `json:"Orders"`
		NextToken   string  `json:"NextToken,omitempty"`
		CreatedBefore string `json:"CreatedBefore,omitempty"`
	} `json:"payload"`
	Errors []Error `json:"errors,omitempty"`
}

// GetOrderResponse is the response schema for the getOrder operation.
type GetOrderResponse struct {
	Payload Order   `json:"payload"`
	Errors  []Error `json:"errors,omitempty"`
}

// GetOrderBuyerInfoResponse is the response schema for the getOrderBuyerInfo operation.
type GetOrderBuyerInfoResponse struct {
	Payload BuyerInfo `json:"payload"`
	Errors  []Error   `json:"errors,omitempty"`
}

// GetOrderAddressResponse is the response schema for the getOrderAddress operation.
type GetOrderAddressResponse struct {
	Payload struct {
		AmazonOrderID string   `json:"AmazonOrderId"`
		ShippingAddress Address `json:"ShippingAddress"`
	} `json:"payload"`
	Errors []Error `json:"errors,omitempty"`
}

// GetOrderItemsResponse is the response schema for the getOrderItems operation.
type GetOrderItemsResponse struct {
	Payload struct {
		OrderItems    []OrderItem `json:"OrderItems"`
		NextToken     string      `json:"NextToken,omitempty"`
		AmazonOrderID string      `json:"AmazonOrderId"`
	} `json:"payload"`
	Errors []Error `json:"errors,omitempty"`
}

// GetOrderItemsBuyerInfoResponse is the response for the getOrderItemsBuyerInfo operation.
type GetOrderItemsBuyerInfoResponse struct {
	Payload struct {
		OrderItems    []OrderItemBuyerInfo `json:"OrderItems"`
		NextToken     string               `json:"NextToken,omitempty"`
		AmazonOrderID string               `json:"AmazonOrderId"`
	} `json:"payload"`
	Errors []Error `json:"errors,omitempty"`
}

// OrderItemBuyerInfo contains buyer information for an order item.
type OrderItemBuyerInfo struct {
	OrderItemID       string `json:"OrderItemId"`
	BuyerCustomizedInfo struct {
		CustomizedURL string `json:"CustomizedURL"`
	} `json:"BuyerCustomizedInfo,omitempty"`
	GiftMessageText string `json:"GiftMessageText,omitempty"`
	GiftWrapPrice   *Money `json:"GiftWrapPrice,omitempty"`
	GiftWrapLevel   string `json:"GiftWrapLevel,omitempty"`
	GiftWrapTax     *Money `json:"GiftWrapTax,omitempty"`
}


// UpdateShipmentStatusRequest is the request body for the updateShipmentStatus operation.
type UpdateShipmentStatusRequest struct {
	MarketplaceID  string `json:"marketplaceId"`
	ShipmentStatus string `json:"shipmentStatus"`
	OrderItems     []struct {
		OrderItemID string `json:"orderItemId"`
		Quantity    int    `json:"quantity"`
	} `json:"orderItems,omitempty"`
}

// Error defines the error response structure.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
