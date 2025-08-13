# Amazon SP-API 批量商品信息更新指南

## 概述
要批量更新商品的价格、库存和其他信息，你需要使用以下SP-API：

1. **Listings Items API** - 用于单个商品更新
2. **Feeds API** - 用于批量操作（推荐用于大批量更新）

## 字段映射分析

你需要更新的字段可以分为以下几类：

### 价格相关字段
- `sku` - 商品SKU标识符
- `price` - 标准价格
- `minimum-seller-allowed-price` - 最低允许价格
- `maximum-seller-allowed-price` - 最高允许价格
- `business-price` - 商业客户价格

### 库存相关字段
- `quantity` - 库存数量
- `handling-time` - 处理时间

### 分层价格相关字段
- `quantity-price-type` - 数量价格类型
- `quantity-lower-bound1` - 数量下限1
- `quantity-price1` - 对应价格1
- `quantity-lower-bound2` - 数量下限2
- `quantity-price2` - 对应价格2
- `quantity-lower-bound3` - 数量下限3
- `quantity-price3` - 对应价格3

## 推荐方案：使用Feeds API进行批量更新

### API端点
```
POST /feeds/2021-06-30/feeds
```

### 认证
需要包含以下Headers：
- `Authorization: Bearer {access_token}`
- `x-amz-access-token: {selling_partner_access_token}`
- `Content-Type: application/json`

### Feed类型
使用 `POST_FLAT_FILE_LISTINGS_DATA` 或 `POST_FLAT_FILE_PRICEANDQUANTITYONLY_UPDATE_DATA`

对于你的需求，建议使用：`POST_FLAT_FILE_LISTINGS_DATA`
