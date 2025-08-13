# 我需要调用Amazon的SP-API对大批量商品的价格等信息进行更新，每个商品要更新的字段如下：
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

请告诉我需要调用哪个或哪几个API才能完成这个工作，request的格式是什么，最好给出一个request的示例

要批量更新您列出的这些商品信息（包括价格、数量、处理时间、业务价格、数量折扣等），您需要使用 **Amazon Selling Partner API (SP-API) 的 Feeds API**，具体是提交一个 **`POST_INVENTORY_PRICING_DATA`** 类型的 Feed。

---

### ✅ 为什么使用 Feeds API？

- 您要更新的字段（如价格、数量、处理时间、业务价格、数量折扣等）**无法通过 Listings Items API 一次性完整支持**（尤其数量折扣、最小/最大允许价格等）。
- `POST_INVENTORY_PRICING_DATA` 是 Amazon 官方推荐用于批量更新库存和定价信息的 Feed 类型，支持您列出的所有字段。
- Feeds API 支持大批量操作（单个 Feed 文件可包含数千个 SKU）。

---

### 🔧 所需 API 步骤

1. **创建 Feed 文档**（`createFeedDocument`）  
   → 获取上传 URL 和 `feedDocumentId`

2. **将您的数据按指定格式写入文件（通常是 XML 或 Tab-delimited TXT）并上传到该 URL**

3. **创建 Feed**（`createFeed`）  
   → 指定 `feedType = "POST_INVENTORY_PRICING_DATA"`，关联 `feedDocumentId`

4. **轮询 Feed 处理状态**（`getFeed`）  
   → 确认是否成功处理

---

### 📄 Feed 格式：推荐使用 **Tab-delimited (.txt)**

Amazon 推荐使用制表符分隔的文本格式（更简单、不易出错）。以下是字段对应关系：

| 列名（Header） | 说明 |
|----------------|------|
| `sku` | 商品 SKU |
| `price` | 销售价格（单位：货币，如 USD） |
| `minimum-seller-allowed-price` | 最低允许售价（可选） |
| `maximum-seller-allowed-price` | 最高允许售价（可选） |
| `quantity` | 可售数量 |
| `handling-time` | 处理时间（发货准备天数，整数） |
| `business-price` | 企业买家价格（可选） |
| `quantity-price-type` | 数量折扣类型（通常为 `PERCENTAGE` 或 `AMOUNT`，但 Amazon 通常只接受 `AMOUNT` 表示固定金额）|
| `quantity-lower-bound1` | 数量折扣下限1（如 10）|
| `quantity-price1` | 对应数量的单价1 |
| `quantity-lower-bound2` | 数量折扣下限2 |
| `quantity-price2` | 对应单价2 |
| `quantity-lower-bound3` | 数量折扣下限3 |
| `quantity-price3` | 对应单价3 |

> ⚠️ 注意：
> - 所有价格字段必须使用 **数字格式**（如 `19.99`），不要带货币符号。
> - 如果某个字段不更新，可留空或省略（但建议保留列，填空值）。
> - `quantity-price-type` 在实际使用中通常设为 `AMOUNT`（表示绝对价格），Amazon 不支持百分比折扣。

---

### 📝 示例：Tab-delimited Feed 内容（.txt 文件）

```txt
sku	price	minimum-seller-allowed-price	maximum-seller-allowed-price	quantity	handling-time	business-price	quantity-price-type	quantity-lower-bound1	quantity-price1	quantity-lower-bound2	quantity-price2	quantity-lower-bound3	quantity-price3
ABC123	29.99	25.00	35.00	100	2	27.99	AMOUNT	10	28.99	50	26.99	100	24.99
XYZ789	15.50		20.00	200	1	14.00	AMOUNT	20	14.50	100	13.00		
```

> 说明：
> - 第二行：SKU `ABC123`，正常价 $29.99，企业价 $27.99，数量 ≥10 时 $28.99，≥50 时 $26.99，≥100 时 $24.99
> - 第三行：`minimum-seller-allowed-price` 留空（表示不更新），`quantity-lower-bound3` 和 `quantity-price3` 也留空

---

### 📤 API 调用流程示例（伪代码）

#### Step 1: 创建 Feed 文档
```http
POST https://sellingpartnerapi-na.amazon.com/feeds/2021-06-30/documents
Content-Type: application/json

{
  "contentType": "text/tab-separated-values; charset=UTF-8"
}
```

响应：
```json
{
  "payload": {
    "feedDocumentId": "amzn1.tortuga.3.1234567890",
    "url": "https://tortuga-prod-na.s3-external-1.amazonaws.com/..."
  }
}
```

#### Step 2: 上传数据到 `url`
使用 PUT 请求上传上述 `.txt` 内容（带 header）

#### Step 3: 创建 Feed
```http
POST https://sellingpartnerapi-na.amazon.com/feeds/2021-06-30/feeds
Content-Type: application/json

{
  "feedType": "POST_INVENTORY_PRICING_DATA",
  "marketplaceIds": ["ATVPDKIKX0DER"], // US marketplace ID
  "inputFeedDocumentId": "amzn1.tortuga.3.1234567890"
}
```

#### Step 4: 轮询状态
```http
GET https://sellingpartnerapi-na.amazon.com/feeds/2021-06-30/feeds/{feedId}
```

---

### 🔗 官方文档参考

- Feeds API: https://developer-docs.amazon.com/sp-api/docs/feeds-api-v2021-06-30-use-case-guide
- Feed Type List: https://developer-docs.amazon.com/sp-api/docs/feed-type-values
- `POST_INVENTORY_PRICING_DATA` 模板：在 Seller Central > Catalog > Add Products via Upload > Inventory Files > Download Template

---

### 💡 提示

- 建议先在 **Sandbox 环境**测试
- 每次 Feed 处理可能需要几分钟到几小时（取决于队列）
- 可通过 `getFeedResultDocument` 下载处理报告，查看错误 SKU

如有需要，我也可以提供 Python 或 Node.js 的完整上传示例代码。


# POST_INVENTORY_PRICING_DATA 这种Feed类型已经被废弃了，请参考amazon的sp-api的最新文档JSON_LISTINGS_FEED，提供相应的解决方案

您说得完全正确！感谢您的指正。

截至 **2023 年底至 2024 年**，Amazon 已逐步弃用传统的 XML/Tab-delimited Feed 类型（包括 `POST_INVENTORY_PRICING_DATA`），并全面推广 **JSON 格式的 Listings Feed**，即 **`JSON_LISTINGS_FEED`**。这是 Amazon SP-API 推荐的现代化、结构化方式，用于创建或更新商品信息，包括价格、库存、配送设置、业务价格和数量折扣等。

---

## ✅ 正确解决方案：使用 `JSON_LISTINGS_FEED`

### 🔧 支持的字段（覆盖您所有需求）
通过 `JSON_LISTINGS_FEED`，您可以更新以下字段（对应您列出的所有项）：

| 您的需求字段 | JSON Listings 中的对应路径 |
|--------------|---------------------------|
| `sku` | `sku`（顶层字段） |
| `price` | `product_type` → `offer` → `price` |
| `minimum-seller-allowed-price` | `product_type` → `offer` → `minimum_seller_allowed_price` |
| `maximum-seller-allowed-price` | `product_type` → `offer` → `maximum_seller_allowed_price` |
| `quantity` | `product_type` → `offer` → `fulfillment_availability` → `quantity_in_stock` |
| `handling-time` | `product_type` → `offer` → `fulfillment_availability` → `lead_time_to_ship`（单位：天） |
| `business-price` | `product_type` → `offer` → `business_price` |
| `quantity-price-type` / `quantity-lower-bound1` / `quantity-price1` 等 | `product_type` → `offer` → `quantity_discounts`（数组）|

> ✅ 所有字段均可通过 `JSON_LISTINGS_FEED` 一次性更新。

---

## 📄 JSON_LISTINGS_FEED 请求格式

### 1. Feed 内容格式（JSON）
每个商品是一个独立的 JSON 对象。多个商品可放在一个数组中（但 Amazon 要求 **每个 Feed 文件只包含一个 JSON 对象**，即每次提交一个 SKU；若需批量，需循环提交多个 Feed 或使用多 SKU 数组——但官方推荐 **一个 Feed 一个 SKU** 以简化错误处理）。

> ⚠️ 注意：虽然技术上可在一个 JSON 数组中包含多个 SKU，但 Amazon **强烈建议每个 Feed 只包含一个 SKU**，否则任一 SKU 失败会导致整个 Feed 失败。

---

### 2. 示例：更新一个 SKU 的完整 JSON

```json
{
  "header": {
    "sellerId": "A1B2C3D4E5F6G7",
    "version": "2.0",
    "feedId": "123e4567-e89b-12d3-a456-426614174000"
  },
  "issues": [],
  "messages": [
    {
      "sku": "ABC123",
      "operationType": "UPDATE",
      "productType": "PRODUCT",
      "requirements": "LISTING_OFFER_ONLY",
      "attributes": {
        "offer": [
          {
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "price": [
              {
                "amount": 29.99,
                "currency": "USD"
              }
            ],
            "minimum_seller_allowed_price": [
              {
                "amount": 25.0,
                "currency": "USD"
              }
            ],
            "maximum_seller_allowed_price": [
              {
                "amount": 35.0,
                "currency": "USD"
              }
            ],
            "business_price": [
              {
                "amount": 27.99,
                "currency": "USD"
              }
            ],
            "quantity_discounts": [
              {
                "type": "AMOUNT",
                "tiers": [
                  {
                    "quantity_lower_bound": 10,
                    "price": {
                      "amount": 28.99,
                      "currency": "USD"
                    }
                  },
                  {
                    "quantity_lower_bound": 50,
                    "price": {
                      "amount": 26.99,
                      "currency": "USD"
                    }
                  },
                  {
                    "quantity_lower_bound": 100,
                    "price": {
                      "amount": 24.99,
                      "currency": "USD"
                    }
                  }
                ]
              }
            ],
            "fulfillment_availability": [
              {
                "quantity_in_stock": 100,
                "lead_time_to_ship": 2
              }
            ]
          }
        ]
      }
    }
  ]
}
```

---

### 🔑 关键说明

| 字段 | 说明 |
|------|------|
| `operationType` | 必须为 `"UPDATE"`（仅更新 offer 信息） |
| `requirements` | 使用 `"LISTING_OFFER_ONLY"` 表示只更新价格/库存等 offer 信息，不修改商品详情（如标题、描述） |
| `productType` | 对大多数标准商品，使用 `"PRODUCT"`；特殊类目需查 Product Type Definitions |
| `quantity_discounts.type` | 仅支持 `"AMOUNT"`（固定金额），不支持百分比 |
| `lead_time_to_ship` | 即 handling-time，单位为 **天（整数）** |
| `currency` | 必须与 marketplace 匹配（如 US 用 USD） |

---

## 📤 API 调用流程（与 Feeds API 一致）

1. **调用 `createFeedDocument`**  
   ```http
   POST /feeds/2021-06-30/documents
   {
     "contentType": "application/json; charset=UTF-8"
   }
   ```

2. **上传上述 JSON 到返回的 S3 URL（PUT 请求）**

3. **调用 `createFeed`**  
   ```http
   POST /feeds/2021-06-30/feeds
   {
     "feedType": "JSON_LISTINGS_FEED",
     "marketplaceIds": ["ATVPDKIKX0DER"],
     "inputFeedDocumentId": "amzn1.tortuga.3.xxxxx"
   }
   ```

4. **轮询 `getFeed` 查看状态，成功后可选下载处理报告**

---

## 📚 官方文档参考

- JSON Listings Feed Guide:  
  https://developer-docs.amazon.com/sp-api/docs/json-listings-feed-schema-v2
- Offer Attributes (含 price, quantity, discounts):  
  https://developer-docs.amazon.com/sp-api/docs/listings-items-api-v2021-08-01-use-case-guide#offer-attributes
- Product Type Definitions (获取类目支持字段):  
  https://developer-docs.amazon.com/sp-api/docs/product-type-definitions-api

---

## 💡 批量处理建议

- 虽然每个 Feed 推荐一个 SKU，但可通过 **并发提交多个 Feed** 实现大批量更新（注意 API 速率限制）。
- 使用 `Listings Items API`（`putListingsItem`）也可实现单 SKU 更新，但 **Feeds API 更适合大批量**，且支持异步处理。

---

如需，我可以提供：
- Python 脚本示例（上传 JSON Feed）
- 如何动态生成 `quantity_discounts` 数组
- 如何查询特定 marketplace 的 product type schema

请告诉我您的具体需求！