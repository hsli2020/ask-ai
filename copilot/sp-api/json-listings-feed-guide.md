# Amazon SP-API 批量更新指南 (JSON_LISTINGS_FEED)

## 核心信息

### ✅ 正确的Feed类型
使用：`JSON_LISTINGS_FEED` （新版本）
~~废弃：`POST_FLAT_FILE_LISTINGS_DATA`~~

### 🎯 适用的API端点
1. **POST /feeds/2021-06-30/documents** - 创建Feed文档
2. **PUT {upload_url}** - 上传JSON数据
3. **POST /feeds/2021-06-30/feeds** - 创建Feed任务
4. **GET /feeds/2021-06-30/feeds/{feedId}** - 监控状态

## 字段对应关系

### 你的字段 → JSON_LISTINGS_FEED 属性映射

| 你的字段 | JSON路径 | 说明 |
|---------|----------|------|
| `sku` | `messages[].sku` | 商品SKU |
| `price` | `attributes.purchasable_offer[].our_price[].schedule[].value_with_tax` | 标准价格 |
| `minimum-seller-allowed-price` | `attributes.purchasable_offer[].minimum_seller_allowed_price[].schedule[].value_with_tax` | 最低价格 |
| `maximum-seller-allowed-price` | `attributes.purchasable_offer[].maximum_seller_allowed_price[].schedule[].value_with_tax` | 最高价格 |
| `quantity` | `attributes.fulfillment_availability[].quantity` | 库存数量 |
| `handling-time` | `attributes.fulfillment_availability[].lead_time_to_ship_max_days` | 处理时间(天) |
| `business-price` | `attributes.purchasable_offer[].business_price[].schedule[].value_with_tax` | 商业价格 |
| `quantity-price-type` | `attributes.purchasable_offer[].quantity_discount_type[].value` | 折扣类型 |
| `quantity-lower-bound1/2/3` | `attributes.purchasable_offer[].quantity_discount[].quantity_lower_bound` | 数量阶梯 |
| `quantity-price1/2/3` | 通过计算转换为折扣金额或百分比 | 阶梯价格 |

## 完整的JSON请求示例

### 1. 创建Feed请求
```json
{
  "feedType": "JSON_LISTINGS_FEED",
  "marketplaceIds": ["ATVPDKIKX0DER"],
  "inputFeedDocument": "amzn1.tortuga.3.xxx..."
}
```

### 2. JSON Feed数据结构
```json
{
  "header": {
    "sellerId": "YOUR_SELLER_ID",
    "version": "2.0",
    "issueLocale": "en_US"
  },
  "messages": [
    {
      "messageId": 1,
      "sku": "YOUR-SKU-001",
      "operationType": "UPDATE",
      "productType": "PRODUCT", 
      "requirements": "LISTING",
      "attributes": {
        "condition_type": [{"value": "new_new"}],
        "purchasable_offer": [{
          "currency": "USD",
          "our_price": [{"schedule": [{"value_with_tax": 29.99}]}],
          "minimum_seller_allowed_price": [{"schedule": [{"value_with_tax": 25.00}]}],
          "maximum_seller_allowed_price": [{"schedule": [{"value_with_tax": 35.00}]}],
          "business_price": [{"schedule": [{"value_with_tax": 27.99}]}],
          "quantity_discount_type": [{"value": "FIXED_AMOUNT"}],
          "quantity_discount": [
            {
              "quantity_tier": 1,
              "quantity_discount_type": "FIXED_AMOUNT",
              "discount_amount": 1.00,
              "quantity_lower_bound": 10
            }
          ]
        }],
        "fulfillment_availability": [{
          "fulfillment_channel_code": "DEFAULT",
          "quantity": 100,
          "lead_time_to_ship_max_days": 2
        }]
      }
    }
  ]
}
```

## 分层价格处理

### 数量折扣类型 (quantity_discount_type)
- `FIXED_AMOUNT`: 固定金额折扣
- `PERCENT_OFF`: 百分比折扣

### 转换逻辑
- **quantity-price-type = "fixed"** → `FIXED_AMOUNT`
- **quantity-price-type = "percent"** → `PERCENT_OFF`

### 价格转换示例
如果你的数据是：
- 基础价格：29.99
- quantity-lower-bound1：10
- quantity-price1：28.99

转换为JSON：
```json
{
  "quantity_tier": 1,
  "quantity_discount_type": "FIXED_AMOUNT",
  "discount_amount": 1.00,  // 29.99 - 28.99 = 1.00
  "quantity_lower_bound": 10
}
```

## 重要注意事项

### ✅ 优势
- JSON格式更灵活、可读性强
- 支持复杂的嵌套结构
- 更好的错误处理和验证
- 官方推荐的新标准

### ⚠️ 注意点
1. **Seller ID必需**：需要在header中提供
2. **Content-Type**：上传时使用 `application/json`
3. **版本号**：使用 "2.0"
4. **操作类型**：通常使用 "UPDATE"
5. **产品类型**：通常使用 "PRODUCT"

### 📋 市场ID参考
- 美国：`ATVPDKIKX0DER`
- 加拿大：`A2EUQ1WTGCTBG2`
- 英国：`A1F83G8C2ARO7P`
- 德国：`A1PA6795UKMFR9`
- 法国：`A13V1IB3VIYZZH`

## 错误处理

常见错误码：
- **99001**: JSON格式错误
- **99002**: 必需字段缺失
- **8541**: 价格超出范围
- **8542**: SKU不存在

## 最佳实践

1. **批量大小**：建议每批不超过1000个商品
2. **重试机制**：实现指数退避重试
3. **状态监控**：定期检查Feed处理状态
4. **错误日志**：记录详细的错误信息
5. **测试环境**：先在沙箱环境测试