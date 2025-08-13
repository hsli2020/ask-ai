# Amazon SP-API 完整的批量更新方案

## 核心API调用

根据你需要更新的字段，你需要调用以下SP-API：

### 1. 主要API：Feeds API (推荐用于批量操作)

**API端点：** `POST /feeds/2021-06-30/feeds`

**Feed类型：** `POST_FLAT_FILE_LISTINGS_DATA`

**适用场景：** 批量更新价格、库存、商品信息

### 2. 备选API：Listings Items API (适用于单个商品更新)

**API端点：** `PATCH /listings/2021-08-01/items/{sellerId}/{sku}`

**适用场景：** 单个商品的精确更新

## 字段映射和API使用

### 使用Feeds API的完整流程：

1. **创建Feed文档**
   ```
   POST /feeds/2021-06-30/documents
   ```

2. **上传数据到S3**
   ```
   PUT {upload_url}
   ```

3. **创建Feed**
   ```
   POST /feeds/2021-06-30/feeds
   ```

4. **监控Feed状态**
   ```
   GET /feeds/2021-06-30/feeds/{feedId}
   ```

### JSON Feed格式（用于Feeds API）

现在使用JSON格式的Feed，支持更灵活的数据结构和所有字段类型。

### 分层价格说明

Amazon SP-API中的分层价格（Quantity Discounts）工作方式：
- `quantity-price-type`: "percent" 或 "fixed"
- `quantity-lower-bound1`: 第一个价格阶梯的最低数量
- `quantity-price1`: 对应的价格
- 以此类推...

## 完整的请求示例

### 1. 创建Feed的请求体
```json
{
  "feedType": "JSON_LISTINGS_FEED",
  "marketplaceIds": ["ATVPDKIKX0DER"],
  "inputFeedDocument": "amzn1.tortuga.3.xxxxxxxxxxxxxxxxxxxx"
}
```

### 2. JSON Feed数据格式
现在使用JSON格式而不是TSV格式，具体结构见下面的示例文件。

## 重要注意事项

1. **认证**：需要有效的SP-API访问令牌
2. **权限**：确保有相应的角色权限
3. **市场ID**：美国站 = "ATVPDKIKX0DER"
4. **处理时间**：以天为单位
5. **价格**：使用市场对应货币
6. **Feed处理**：通常需要几分钟到几小时

## 错误处理

常见错误及解决方案：
- **8541**: 价格超出允许范围 → 检查最低/最高价格设置
- **8542**: SKU不存在 → 确认SKU在Amazon系统中存在
- **18016**: 库存数量无效 → 检查数量格式

## 批处理建议

- 每个Feed建议不超过10,000个商品
- 大批量数据建议分批处理
- 实现重试机制处理临时错误
- 监控Feed处理结果