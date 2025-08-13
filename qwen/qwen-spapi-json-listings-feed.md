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

根据亚马逊SP-API的最新规范，**JSON_LISTINGS_FEED** 使用 Listings Items API 的 Schema 格式（v2021-08-01），与传统CSV字段的映射关系如下。以下是完整的请求格式示例：

### 一、CSV字段到JSON的映射关系

| CSV字段 | JSON路径 | 说明 |
|--------|----------|------|
| `sku` | 顶层 `sku` 字段 | 商品SKU标识 |
| `price` | `attributes.purchasable_offer[0].our_price` | 标准售价（B2C） |
| `quantity` | `attributes.fulfillment_availability[0].quantity` | 库存数量 |
| `handling-time` | `attributes.fulfillment_availability[0].fulfillment_latency` | **注意**：CSV中的"handling-time"在JSON中对应`fulfillment_latency`（单位：天）[[169]] |
| `business-price` | `attributes.b2b_purchasable_offer[0].our_price` | 企业购价格（B2B） |
| `quantity-price-type` | `attributes.b2b_purchasable_offer[0].quantity_discounts[0].discount_type` | "percent"或"fixed" |
| `quantity-lower-boundX` | `attributes.b2b_purchasable_offer[0].quantity_discounts[X-1].quantity_tier` | 数量阶梯阈值 |
| `quantity-priceX` | `attributes.b2b_purchasable_offer[0].quantity_discounts[X-1].price` | 对应阶梯价格或折扣 |
| `minimum/maximum-seller-allowed-price` | **不直接支持** | 这些是Automate Pricing功能的约束，需通过[Automate Pricing API](https://developer-docs.amazon.com/sp-api/docs/automate-pricing-api-v0)单独设置 [[133]] |

> ⚠️ **重要提示**：  
> 1. `minimum-seller-allowed-price` 和 `maximum-seller-allowed-price` **不能**通过JSON_LISTINGS_FEED直接设置，它们属于Automate Pricing（自动定价）功能的约束条件，需使用专门的Automate Pricing API管理 [[81]]  
> 2. 处理时间字段在JSON中命名为 `fulfillment_latency`（而非handling-time），单位为**自然日**（0=当日发货，1=次日发货）[[169]]

### 二、完整JSON_LISTINGS_FEED请求示例

```json
{
  "header": {
    "sellerId": "A1B2C3D4E5F6G7",
    "version": "2.0",
    "feedType": "JSON_LISTINGS_FEED"
  },
  "messages": [
    {
      "messageId": 1,
      "sku": "YOUR-SKU-001",
      "operationType": "PATCH",  // 使用PATCH仅更新指定字段，无需提供完整商品信息
      "productType": "PRODUCT",  // 通用商品类型，多数情况可用"PRODUCT"；复杂品类需查询Product Type Definitions API获取准确类型
      "attributes": {
        "purchasable_offer": [
          {
            "marketplace_id": "ATVPDKIKX0DER",  // 美国站marketplace ID
            "currency": "USD",
            "our_price": [
              {
                "schedule": [
                  {
                    "value": 29.99  // price字段对应值
                  }
                ]
              }
            ]
          }
        ],
        "fulfillment_availability": [
          {
            "fulfillment_channel_code": "DEFAULT",  // MFN自发货渠道
            "quantity": 150,                        // quantity字段对应值
            "fulfillment_latency": 2                // handling-time字段对应值（2天处理时间）
          }
        ],
        "b2b_purchasable_offer": [
          {
            "marketplace_id": "ATVPDKIKX0DER",
            "currency": "USD",
            "our_price": [
              {
                "schedule": [
                  {
                    "value": 24.99  // business-price字段对应值
                  }
                ]
              }
            ],
            "quantity_discounts": [
              {
                "discount_type": "percent",         // quantity-price-type: "percent"或"fixed"
                "quantity_tier": 10,                // quantity-lower-bound1
                "price": 5.0                        // quantity-price1（5%折扣）
              },
              {
                "discount_type": "percent",
                "quantity_tier": 50,                // quantity-lower-bound2
                "price": 10.0                       // quantity-price2（10%折扣）
              },
              {
                "discount_type": "percent",
                "quantity_tier": 100,               // quantity-lower-bound3
                "price": 15.0                       // quantity-price3（15%折扣）
              }
            ]
          }
        ]
      }
    }
  ]
}
```

### 三、关键注意事项

1. **操作类型选择**：
   - `PATCH`：仅更新指定字段（推荐用于价格/库存更新）[[111]]
   - `UPDATE`：替换整个商品信息（需提供完整必填属性）
   - `DELETE`：删除商品

2. **最小/最大价格限制**：
   - 这些字段属于Automate Pricing功能，**不能**通过JSON_LISTINGS_FEED设置
   - 需使用 [Automate Pricing API](https://developer-docs.amazon.com/sp-api/docs/automate-pricing-api-v0) 的 `putPricingTemplate` 或 `putCompetitivePricing` 接口管理 [[133]]

3. **业务价格（B2B）前提条件**：
   - 必须先在Seller Central激活Amazon Business功能 [[161]]
   - 需通过Product Type Definitions API确认商品类型是否支持`b2b_purchasable_offer`属性

4. **Feed提交流程**：
   ```mermaid
   1. createFeedDocument → 获取加密上传URL
   2. 上传JSON内容到URL
   3. createFeed → 提交Feed（feedType="JSON_LISTINGS_FEED"）
   4. getFeed → 轮询状态
   5. getFeedResult → 下载处理报告
   ```

5. **速率限制**：
   - JSON_LISTINGS_FEED受Listings Items API相同速率限制（通常60次/分钟）[[46]]
   - 批量操作建议单次Feed包含≤1000个SKU

### 四、调试建议

- 使用 [Product Type Definitions API](https://developer-docs.amazon.com/sp-api/docs/product-type-definitions-api-v0) 获取准确的商品类型Schema：
  ```http
  GET /definitions/2020-09-01/productTypes/{productType}
  ```
- 通过 `getListingsItem` 接口查看现有商品的JSON结构，作为PATCH操作的参考 [[90]]
- 处理报告中会返回详细的字段验证错误，重点关注`processingSummary`和`itemProcessingResults` [[71]]

> 💡 **迁移提示**：2025年7月31日后，传统XML/Flat File Feed类型（如POST_PRODUCT_DATA）将被完全弃用，必须迁移到JSON_LISTINGS_FEED [[1]]

