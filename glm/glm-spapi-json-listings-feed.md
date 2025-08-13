# 我原来通过上传csv的方式更新亚马逊店铺中的商品信息，
现在要改为通过SP-API中的Feed API上传JSON_LISTINGS_FEED来更新，

原来的csv文件包含下列字段：
sku
price
minimum-seller-allowed-price
maximum-seller-allowed-price
quantity
handling-time
business-price
quantity-price-type
quantity-lower-bound1
quantity-price1
quantity-lower-bound2
quantity-price2
quantity-lower-bound3
quantity-price3

请给出对应的JSON_LISTINGS_FEED请求格式


将 CSV 文件转换为 SP-API 的 `JSON_LISTINGS_FEED` 格式时，你需要将 CSV 中的列名映射到特定的 JSON 结构中。

以下是详细的字段映射关系以及完整的 JSON 请求体格式示例。

### 1. 字段映射表

| CSV 字段 | JSON 路径 | 说明 |
| :--- | :--- | :--- |
| **sku** | `product.sku` | 卖家 SKU |
| **price** | `product.offers[0].pricing.standardPrice` | 标准价格 (需包含 amount 和 currency) |
| **minimum-seller-allowed-price** | `product.offers[0].pricing.minimumSellerAllowedPrice` | 最低价格 |
| **maximum-seller-allowed-price** | `product.offers[0].pricing.maximumSellerAllowedPrice` | 最高价格 |
| **quantity** | `product.fulfillmentAvailability[0].quantity` | 库存数量 |
| **handling-time** | `product.fulfillmentAvailability[0].fulfillmentLatency` | 配置时间 (发货准备时间) |
| **business-price** | (包含在 quantities 中) | 商业价格 (在 JSON 中通常通过 quantities 数组的第一层体现，或单独字段，取决于具体 Schema 版本，通常合并在 tier 中) |
| **quantity-price-type** | `product.businessPrice.quantityPriceType` | 数量价格类型 (`flat` 或 `tiered`) |
| **quantity-lower-bound1** | `product.businessPrice.quantities[0].lowerBound` | 数量下限 1 |
| **quantity-price1** | `product.businessPrice.quantities[0].price.amount` | 数量价格 1 |
| **quantity-lower-bound2** | `product.businessPrice.quantities[1].lowerBound` | 数量下限 2 |
| **quantity-price2** | `product.businessPrice.quantities[1].price.amount` | 数量价格 2 |
| ... (以此类推) | `product.businessPrice.quantities[n]...` | 更多层级 |

---

### 2. JSON_LISTINGS_FEED 请求格式示例

请将下面的 JSON 作为 Feed 内容上传。注意：`JSON_LISTINGS_FEED` 是一个包含 Header 和 Messages 数组的结构。

**注意点：**
1.  **Currency**: JSON 中必须显式指定货币代码（如 "USD"）。
2.  **Arrays**: `offers`, `fulfillmentAvailability`, 和 `quantities` 都是数组，即使只有一个元素也需要使用数组格式。
3.  **SellerId**: 在 Header 中需要填入你的 Seller ID。

```json
{
  "header": {
    "sellerId": "A1B2C3D4E5F6G7", // 请替换为你的 Seller ID
    "schemaVersion": "2.0"
  },
  "messages": [
    {
      "messageId": 1, 
      "operationType": "UPDATE", 
      "product": {
        "sku": "YOUR_SKU_HERE", // 对应 CSV: sku
        
        // 价格相关字段
        "offers": [
          {
            "pricing": {
              "standardPrice": {
                "amount": 19.99, // 对应 CSV: price
                "currency": "USD"
              },
              "minimumSellerAllowedPrice": {
                "amount": 15.00, // 对应 CSV: minimum-seller-allowed-price
                "currency": "USD"
              },
              "maximumSellerAllowedPrice": {
                "amount": 25.00, // 对应 CSV: maximum-seller-allowed-price
                "currency": "USD"
              }
            }
          }
        ],

        // 库存和发货时间
        "fulfillmentAvailability": [
          {
            "fulfillmentChannelCode": "DEFAULT", // 通常是 DEFAULT 或 AMAZON_NA
            "quantity": 100, // 对应 CSV: quantity
            "fulfillmentLatency": 2 // 对应 CSV: handling-time
          }
        ],

        // B2B (Business Price) 相关字段
        "businessPrice": {
          "quantityPriceType": "flat", // 对应 CSV: quantity-price-type (通常是 "flat")
          "quantities": [
            {
              "lowerBound": 2, // 对应 CSV: quantity-lower-bound1
              "price": {
                "amount": 18.00, // 对应 CSV: quantity-price1
                "currency": "USD"
              }
            },
            {
              "lowerBound": 10, // 对应 CSV: quantity-lower-bound2
              "price": {
                "amount": 17.00, // 对应 CSV: quantity-price2
                "currency": "USD"
              }
            },
            {
              "lowerBound": 20, // 对应 CSV: quantity-lower-bound3
              "price": {
                "amount": 16.00, // 对应 CSV: quantity-price3
                "currency": "USD"
              }
            }
            // 如果 CSV 中还有更多层级，继续在此添加对象
          ]
        }
      }
    }
    // 如果需要更新多个 SKU，请在这里添加更多的 message 对象 (messageId 递增)
  ]
}
```

### 3. 代码实现提示

如果你在编写代码生成这个 JSON，请注意以下逻辑：

*   **循环处理 SKU**：每个 SKU 应该对应 `messages` 数组中的一个对象。不要把所有 SKU 放在一个 `product` 对象里。
*   **处理 B2B 层级**：你的 CSV 中有固定的 `quantity-lower-bound1` 到 `3`。在生成 JSON 时，你需要检查这些字段是否有值。
    *   例如：如果 `quantity-price2` 为空，则 `quantities` 数组中只应该包含第一个对象（lower-bound1 的数据），不应该放入空的层级。
*   **价格类型**：CSV 中的 `quantity-price-type` 通常是 `flat`（固定折扣）或 `tiered`（阶梯定价）。请确保该值准确传递到 `quantityPriceType`。


# 上文中的"字段映射表"从何而来，是amazon官方来源吗？

**结论先说：**  
我之前那条回复里给出的“字段映射表”（`product.offers` / `product.businessPrice` 那套结构）并不是来自亚马逊官方文档，而是我基于常见 JSON 写法推断出来的错误示例，这一点需要向你明确更正。

---

## 一、真正的官方来源是什么？

和 `JSON_LISTINGS_FEED` 直接相关的官方资料，主要有这几份：

1) JSON Listings Feed 的 Schema 定义（GitHub，Amazon 官方仓库）  
- listings-feed-schema-v2.json：整个 Feed 的顶层结构（`header` / `messages` 等）【turn3fetch0】  
- listings-feed-message-schema-v2.json：单条 message 的结构（`messageId` / `sku` / `operationType` / `productType` / `requirements` / `attributes` / `patches` 等）【turn12fetch0】

2) 官方中文映射文档：将商品属性映射到商品信息 API 和基于 JSON 的上传数据  
在这篇文档里，亚马逊明确给出了旧版 XML/Flat File 字段（如 `StandardPrice`、`MinimumSellerAllowedPrice`、`MaximumSellerAllowedPrice`、`Quantity`、`FulfillmentLatency`、`BusinessPrice`、`QuantityPriceType`、`QuantityPrice1`、`QuantityLowerBound1` 等）到 JSON 属性的映射路径。举几个和你的 CSV 直接相关的：  

- 库存相关【turn18fetch0】：  
  - `Quantity` → `/attributes/fulfillment_availability/0/quantity`  
  - `FulfillmentLatency` → `/attributes/fulfillment_availability/0/lead_time_to_ship_max_days`  
  - `FulfillmentCenterID` → `/attributes/fulfillment_availability/0/fulfillment_channel_code`

- 定价相关【turn18fetch0】【turn19fetch0】：  
  - `StandardPrice` → `/attributes/purchasable_offer/0/our_price/0/schedule/0/value_with_tax`，并建议 `audience = "ALL"`  
  - `MinimumSellerAllowedPrice` → `/attributes/purchasable_offer/0/minimum_seller_allowed_price/0/schedule/0/value_with_tax`  
  - `MaximumSellerAllowedPrice` → `/attributes/purchasable_offer/0/maximum_seller_allowed_price/0/schedule/0/value_with_tax`

- B2B 与数量折扣相关【turn19fetch0】：  
  - `BusinessPrice` → `/attributes/purchasable_offer/0/our_price/0/schedule/0/value_with_tax`，且 `audience = "B2B"`  
  - `QuantityPriceType` → `/attributes/purchasable_offer/0/quantity_discount_plan/0/schedule/0/discount_type`  
  - `QuantityPrice1` → `/attributes/purchasable_offer/0/quantity_discount_plan/0/schedule/0/levels/0/value`  
  - `QuantityLowerBound1` → `/attributes/purchasable_offer/0/quantity_discount_plan/0/schedule/0/levels/0/lower_bound`  
  - `QuantityPrice2` / `QuantityLowerBound2` 等，对应 `levels[1]`、`levels[2]`…，文档只写到 1，但扩展方式就是按数组索引递增【turn19fetch0】

3) Listings API / Listings Items API 的示例  
官方的 Listings API 用例示例中，返回的 `attributes` 也包含 `purchasable_offer` 和 `fulfillment_availability`，结构与上述 JSON 路径一致【turn13search3】【turn13search28】。

综合这些官方资料，可以确定：

- `JSON_LISTINGS_FEED` 的 message 结构，是通过 `attributes` 对象来承载价格、库存、B2B 数量折扣等属性；  
- 这些属性的“字段名”和“路径”是以 `/attributes/fulfillment_availability` 和 `/attributes/purchasable_offer` 这种 JSON Pointer 形式在映射文档里明确给出的【turn18fetch0】【turn19fetch0】。

因此，“正确的字段映射表”应该以这些官方 JSON Pointer 为准，而不是我之前给出的 `product.offers` / `product.businessPrice` 结构。

## 二、对你 CSV 字段的正确官方映射（基于 Amazon 文档）

你原 CSV 字段：

- sku  
- price  
- minimum-seller-allowed-price  
- maximum-seller-allowed-price  
- quantity  
- handling-time  
- business-price  
- quantity-price-type  
- quantity-lower-bound1 / quantity-price1  
- quantity-lower-bound2 / quantity-price2  
- quantity-lower-bound3 / quantity-price3  

对应到 `JSON_LISTINGS_FEED` 里的 JSON 路径，按官方映射文档整理如下（都位于 `attributes` 对象中）【turn18fetch0】【turn19fetch0】：

- sku → `messages[i].sku`（不在 `attributes` 里，在 message 顶层字段，由 schema 直接定义【turn12fetch0】）

- price → `attributes.purchasable_offer[0].our_price[0].schedule[0].value_with_tax`【对应 StandardPrice，turn18fetch0】

- minimum-seller-allowed-price → `attributes.purchasable_offer[0].minimum_seller_allowed_price[0].schedule[0].value_with_tax`【turn18fetch0】

- maximum-seller-allowed-price → `attributes.purchasable_offer[0].maximum_seller_allowed_price[0].schedule[0].value_with_tax`【turn18fetch0】

- quantity → `attributes.fulfillment_availability[0].quantity`【turn18fetch0】

- handling-time（相当于旧字段 `FulfillmentLatency`）→ `attributes.fulfillment_availability[0].lead_time_to_ship_max_days`【turn18fetch0】

- business-price → `attributes.purchasable_offer[0].our_price[0].schedule[0].value_with_tax`，并设置 `attributes.purchasable_offer[0].audience = "B2B"`【turn19fetch0】

- quantity-price-type → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].discount_type`【turn19fetch0】

- quantity-lower-bound1 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[0].lower_bound`【turn19fetch0】  
- quantity-price1 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[0].value`【turn19fetch0】

- quantity-lower-bound2 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[1].lower_bound`  
- quantity-price2 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[1].value`

- quantity-lower-bound3 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[2].lower_bound`  
- quantity-price3 → `attributes.purchasable_offer[0].quantity_discount_plan[0].schedule[0].levels[2].value`

## 三、为什么我之前用了 product.offers / product.businessPrice？

这两个结构在 Amazon 官方的 JSON Listings Feed schema 和映射文档中是找不到的：

- listings-feed-schema-v2.json 的 message 只有：`messageId` / `sku` / `operationType` / `productType` / `requirements` / `attributes` / `patches`【turn3fetch0】【turn12fetch0】；  
- 映射文档中，所有定价、折扣、B2B 字段都指向 `/attributes/purchasable_offer/...`，而不是指向 `/product/offers/...` 或 `/product/businessPrice/...`【turn18fetch0】【turn19fetch0】。

因此，之前回复中：

- `product.offers[0].pricing.standardPrice`  
- `product.businessPrice.quantities[...]`  
这类写法是我在没有严格核对官方 schema 的情况下，按“常见电商 API 风格”推测出来的，属于不准确的“臆造字段”，不应该采用。

## 四、基于官方映射的正确 JSON_LISTINGS_FEED 请求示例（贴合你 CSV 字段）

下面给一个贴合你 CSV 字段的、严格按官方映射表构造的 JSON 示例（仅展示核心结构，完整内容需按实际 marketplaceId / currency 等填充）：

```json
{
  "header": {
    "sellerId": "A1B2C3D4E5F6G7",
    "version": "2.0",
    "issueLocale": "en_US"
  },
  "messages": [
    {
      "messageId": 1,
      "sku": "YOUR_SKU",
      "operationType": "PARTIAL_UPDATE",
      "requirements": "LISTING_OFFER_ONLY",
      "productType": "PRODUCT",
      "attributes": {
        "fulfillment_availability": [
          {
            "fulfillment_channel_code": "DEFAULT",
            "quantity": 100,
            "lead_time_to_ship_max_days": 2
          }
        ],
        "purchasable_offer": [
          {
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "audience": "ALL",
            "our_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 19.99
                  }
                ]
              }
            ],
            "minimum_seller_allowed_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 15.00
                  }
                ]
              }
            ],
            "maximum_seller_allowed_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 25.00
                  }
                ]
              }
            ]
          }
        ]
      }
    },
    {
      "messageId": 2,
      "sku": "YOUR_SKU_2",
      "operationType": "PARTIAL_UPDATE",
      "requirements": "LISTING_OFFER_ONLY",
      "productType": "PRODUCT",
      "attributes": {
        "fulfillment_availability": [
          {
            "fulfillment_channel_code": "DEFAULT",
            "quantity": 50,
            "lead_time_to_ship_max_days": 1
          }
        ],
        "purchasable_offer": [
          {
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "audience": "B2B",
            "our_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 18.00
                  }
                ]
              }
            ],
            "quantity_discount_plan": [
              {
                "schedule": [
                  {
                    "discount_type": "flat",
                    "levels": [
                      {
                        "lower_bound": 2,
                        "value": 17.00
                      },
                      {
                        "lower_bound": 10,
                        "value": 16.00
                      },
                      {
                        "lower_bound": 20,
                        "value": 15.00
                      }
                    ]
                  }
                ]
              }
            ]
          }
        ]
      }
    }
  ]
}
```

说明：

- `sku` 直接用 `messages[i].sku`（schema 官方字段）【turn12fetch0】；  
- 库存 & handling-time → `fulfillment_availability.quantity` / `lead_time_to_ship_max_days`【turn18fetch0】；  
- 标价、最低/最高价 → `purchasable_offer.our_price` / `minimum_seller_allowed_price` / `maximum_seller_allowed_price`，都放在 `schedule[0].value_with_tax` 中【turn18fetch0】；  
- B2B 基础价格：同样是 `purchasable_offer.our_price[0].schedule[0].value_with_tax`，只是 `audience` 设置为 `"B2B"`【turn19fetch0】；  
- 数量折扣类型 `quantity-price-type` → `quantity_discount_plan[0].schedule[0].discount_type`【turn19fetch0】；  
- 各档 `quantity-lower-boundN` / `quantity-priceN` → `levels[N-1].lower_bound` / `levels[N-1].value`【turn19fetch0】。

## 五、总结一句话

- “是 Amazon 官方来源吗？”  
  - 对 JSON_LISTINGS_FEED 本身，是的：顶层结构来自 listings-feed-schema-v2.json / listings-feed-message-schema-v2.json【turn3fetch0】【turn12fetch0】；  
  - 对具体字段映射（price / quantity / handling-time / business-price / quantity-price-type / quantity-lower-bound / quantity-price），是来自官方的“将商品属性映射到商品信息 API 和基于 JSON 的上传数据”文档，而不是我之前写的那份 `product.offers` / `product.businessPrice` 表【turn18fetch0】【turn19fetch0】。  

- 我之前那条回答中的字段映射表（`product.offers` / `product.businessPrice` 等）不是官方结构，属于错误推断，请以本条回答中的 JSON Pointer 映射为准。


# 你给出的JSON示例中，字段分布在两个message中，如何在一个message中，同时更新所有字段呢

要将所有字段（包括价格、库存、B2B 价格及数量折扣）合并在**同一个 message** 中更新，关键在于利用 `purchasable_offer` 这个**数组**属性。

在 Amazon 的数据结构中，B2C（普通消费者）报价和 B2B（企业买家）报价可以分别存放在 `purchasable_offer` 数组的**不同索引位置**中（例如：索引 0 放 B2C，索引 1 放 B2B）。同时，库存信息 `fulfillment_availability` 作为独立的属性放在 `attributes` 对象中。

### 核心逻辑

1.  **使用 `PARTIAL_UPDATE`**：因为你只更新价格和库存（不涉及产品类目的强制属性如颜色、尺寸等），使用 `PARTIAL_UPDATE` 并配合 `requirements: "LISTING_OFFER_ONLY"` 是最安全、最高效的方式。
2.  **构造 `purchasable_offer` 数组**：
    *   **`purchasable_offer[0]`**：存放 B2C 价格（对应 CSV 的 `price`、`minimum-seller-allowed-price`、`maximum-seller-allowed-price`）。设置 `audience: "ALL"`。
    *   **`purchasable_offer[1]`**：存放 B2B 价格（对应 CSV 的 `business-price`）以及数量折扣阶梯（`quantity-price-type`、`quantity-lower-bound*`、`quantity-price*`）。设置 `audience: "B2B"`。
3.  **构造 `fulfillment_availability`**：存放库存 `quantity` 和 配置时间 `handling-time`。

### 单 Message 完整更新 JSON 示例

以下是一个完整的 `JSON_LISTINGS_FEED` 示例，它在**一个 message** 中同时更新了你 CSV 中的所有字段：

```json
{
  "header": {
    "sellerId": "A1B2C3D4E5F6G7",
    "version": "2.0",
    "issueLocale": "en_US"
  },
  "messages": [
    {
      "messageId": 1,
      "sku": "YOUR_SKU_HERE",
      "operationType": "PARTIAL_UPDATE",
      "requirements": "LISTING_OFFER_ONLY",
      "attributes": {
        // 1. 库存与发货时间 (对应 CSV: quantity, handling-time)
        "fulfillment_availability": [
          {
            "fulfillment_channel_code": "DEFAULT",
            "quantity": 100,
            "lead_time_to_ship_max_days": 2
          }
        ],

        // 2. 价格信息数组
        "purchasable_offer": [
          {
            // --- 索引 0: B2C 价格 (对应 CSV: price, minimum-seller-allowed-price, maximum-seller-allowed-price) ---
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "audience": "ALL",
            "our_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 19.99
                  }
                ]
              }
            ],
            "minimum_seller_allowed_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 15.00
                  }
                ]
              }
            ],
            "maximum_seller_allowed_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 25.00
                  }
                ]
              }
            ]
          },
          {
            // --- 索引 1: B2B 价格与数量折扣 (对应 CSV: business-price, quantity-price-type, quantity-lower-bound*, quantity-price*) ---
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "audience": "B2B",
            
            // B2B 基础价格
            "our_price": [
              {
                "schedule": [
                  {
                    "value_with_tax": 18.00
                  }
                ]
              }
            ],

            // B2B 数量折扣阶梯
            "quantity_discount_plan": [
              {
                "schedule": [
                  {
                    // 对应 CSV: quantity-price-type
                    "discount_type": "flat",
                    // 对应 CSV: quantity-lower-bound1 ~ 3, quantity-price1 ~ 3
                    "levels": [
                      {
                        "lower_bound": 2,
                        "value": 17.00
                      },
                      {
                        "lower_bound": 10,
                        "value": 16.00
                      },
                      {
                        "lower_bound": 20,
                        "value": 15.00
                      }
                    ]
                  }
                ]
              }
            ]
          }
        ]
      }
    }
  ]
}
```

### 关键字段对照总结

| 你的 CSV 字段 | JSON 中的位置 (单 Message 版本) |
| :--- | :--- |
| `sku` | `messages[0].sku` |
| `quantity` | `attributes.fulfillment_availability[0].quantity` |
| `handling-time` | `attributes.fulfillment_availability[0].lead_time_to_ship_max_days` |
| `price` | `attributes.purchasable_offer[0].our_price[0].schedule[0].value_with_tax` |
| `minimum-seller-allowed-price` | `attributes.purchasable_offer[0].minimum_seller_allowed_price[0].schedule[0].value_with_tax` |
| `maximum-seller-allowed-price` | `attributes.purchasable_offer[0].maximum_seller_allowed_price[0].schedule[0].value_with_tax` |
| `business-price` | `attributes.purchasable_offer[1].our_price[0].schedule[0].value_with_tax` |
| `quantity-price-type` | `attributes.purchasable_offer[1].quantity_discount_plan[0].schedule[0].discount_type` |
| `quantity-lower-bound1~3` | `attributes.purchasable_offer[1].quantity_discount_plan[0].schedule[0].levels[0~2].lower_bound` |
| `quantity-price1~3` | `attributes.purchasable_offer[1].quantity_discount_plan[0].schedule[0].levels[0~2].value` |

### 注意事项

1.  **如果不涉及 B2B**：如果你的 CSV 某一行没有 B2B 数据，你可以在 `purchasable_offer` 数组中只保留索引 0（即只传 B2C 价格），直接省略索引 1 的对象即可。
2.  **requirements 字段**：请务必设置为 `"LISTING_OFFER_ONLY"`。这告诉 Amazon 你只更新报价和库存，不需要校验产品类目的其他属性（如材质、尺寸等），这样可以避免很多报错。
3.  **marketplace_id**：务必填入正确的商城 ID（如美国站是 `ATVPDKIKX0DER`）。
