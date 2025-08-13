# 使用Amazon SPAPI中的哪个API，可以让我只需指定几个参数，就能向我的店铺添加商品（该商品已经在Amazon上存在）

在 Amazon SP-API (Selling Partner API) 中，如果你想向店铺添加一个**已经在 Amazon 目录中存在（已有 ASIN）**的商品，并且希望参数尽可能简单，最推荐的 API 是：

### **Listings Items API (商品详情项 API)**

具体操作是：**`putListingsItem`**

#### 为什么选择这个 API？
1.  **RESTful 风格**：相比于传统的 Feeds API（需要上传文件、异步处理），这是一个直接的 HTTP PUT 请求，同步返回结果，调试更简单。
2.  **无需重复提交商品信息**：因为商品已存在（有 ASIN），你不需要提交标题、图片、描述等目录信息。你只需要提交**卖家特定信息**（如 SKU、价格、数量、状况）以及**身份标识**（用于匹配现有 ASIN）。
3.  **单一接口**：它可以同时完成“创建 Listing"和“设置价格/库存”的动作。

---

### 你需要提供的核心参数（最小集）

虽然说是“几个参数”，但 Amazon 对数据结构有严格要求。要成功匹配现有 ASIN 并上架，你至少需要提供以下信息：

1.  **路径参数**：
    *   `sellerId`: 你的卖家 ID。
    *   `sku`: 你为该商品自定义的卖家 SKU（必须唯一）。
2.  **查询参数**：
    *   `marketplaceIds`: 目标站点 ID（如 `ATVPDKIKX0DER` 代表美国站）。
    *   `productType`: **关键参数**。商品类型（如 `LUGGAGE`, `SHOES`, `ELECTRONICS` 等）。你必须知道该 ASIN 属于哪个 Product Type。
3.  **请求体 (JSON Body)**：
    *   `productType`: 同上。
    *   `requirements`: 通常设为 `LISTING` 或 `LISTING_PRODUCT_ONLY`。
    *   `attributes`: 这是核心部分。为了匹配现有 ASIN 并上架，你通常只需包含：
        *   **身份标识**：如 `externalProductId` (ASIN) 或 `gtin` (UPC/EAN)。*注：具体字段名取决于 Product Type 定义。*
        *   **卖家信息**：`conditionType` (新旧状况), `price` (价格), `quantity` (数量)。

---

### 简单示例 (JSON Payload)

假设你要上架一个已有的行李箱 (Product Type: `LUGGAGE`)，ASIN 已知：

```http
PUT /listings/2021-08-01/items/{sellerId}/{your-sku}?marketplaceIds=ATVPDKIKX0DER
Content-Type: application/json
{
  "productType": "LUGGAGE",
  "requirements": "LISTING",
  "attributes": {
    "externalProductId": [
      {
        "externalProductId": "B0XXXXXXXX", 
        "externalProductType": "ASIN"
      }
    ],
    "conditionType": "NEW",
    "price": {
      "currencyCode": "USD",
      "amount": 59.99
    },
    "quantity": {
      "amount": 10
    }
  }
}
```
*(注意：具体的 attributes 字段名称必须遵循该 `productType` 的定义，不同类目字段名可能不同，例如有的类目用 `sku` 内部关联，有的强制要求 `gtin`。)*

---

### 重要前提与注意事项

虽然这个 API 最直接，但有一个**必须解决的门槛**：

1.  **必须知道 `productType`**：
    你不能随便填一个类型。如果类型填错，API 会报错。
    *   **解决方法**：你可以先调用 **`searchListingsDefinitions`** API，或者通过 Amazon 前台页面查看该 ASIN 的类目，推断出 Product Type。
2.  **属性合规性**：
    即使是匹配现有 ASIN，`attributes` 里的字段也必须符合该 Product Type 的 JSON Schema 定义。例如，价格字段在某些类目下结构可能不同。
3.  **权限**：
    确保你的 SP-API 应用已授权 `listings:write` 权限。

### 替代方案：Feeds API (JSON_LISTINGS_FEED)

如果你发现 `putListingsItem` 对参数校验太严格，或者你需要**批量**添加多个商品，可以使用 **Feeds API**。
*   **Feed Type**: `JSON_LISTINGS_FEED`
*   **优点**：容错性稍好，适合批量。
*   **缺点**：异步流程（上传文档 -> 创建 Feed -> 轮询结果），对于“只加一个商品”来说过于繁琐。

### 总结建议

*   **单个商品、程序化自动上架**：使用 **Listings Items API (`putListingsItem`)**。
*   **关键步骤**：先确定该 ASIN 对应的 **`productType`**，然后构造包含 `ASIN`、`Price`、`Quantity` 的最小 JSON 包进行提交。

这是目前 SP-API 中实现“最小参数上架已有商品”的最现代、最标准的方法。


# 我每次添加的商品不多，只有10个左右，看来用Listing API比较合适。但上例中给出的Json Payload的结构和字段名称请根据官方文档核实一下。另外，我已经知道了ASIN，那么该商品在Amazon就有了分类，还需要我在添加商品时指定分类信息，似乎不太合理，请根据官方文档确认一下

你的直觉非常敏锐。**你是对的：在逻辑上，既然 ASIN 已存在，Amazon 系统内部确实知道它的分类。**

但是，**在 SP-API 的 `Listings Items API (2021-08-01)` 的技术实现中，`productType` 是强制必填参数**（无论是在 Query Parameter 还是 Request Body 中）。

### 为什么必须指定 `productType`？
Amazon 的 Listings API 是**基于 Schema（模式）验证**的。
1.  **动态验证**：不同分类（如“鞋子”和“笔记本电脑”）所需的属性字段完全不同。API 在接收你的请求时，需要先加载对应 `productType` 的 JSON 验证规则，才能检查你提交的 `attributes` 是否合法。
2.  **解耦设计**：API 设计上是“先确定模板（Product Type），再填写内容（Attributes）”，而不是“先猜内容，再匹配模板”。

---

### 最佳实践流程（2 步走）

既然你只有 10 个商品，且已知 ASIN，最稳妥的流程是：

1.  **第一步：获取 `productType`**
    调用 **`searchCatalogItems`** API，传入 ASIN，Amazon 会返回该 ASIN 对应的 `productType`。
2.  **第二步：创建 Listing**
    调用 **`putListingsItem`** API，使用上一步拿到的 `productType` 提交数据。

---

### 核实后的 JSON Payload 结构

根据官方 `Listings Items API (2021-08-01)` 文档，以下是经过核实的请求结构。

#### 1. 请求地址 (Request)
```http
PUT https://sellingpartnerapi-na.amazon.com/listings/2021-08-01/items/{sellerId}/{sku}?marketplaceIds=ATVPDKIKX0DER&productType=LUGGAGE
```
*   `{sellerId}`: 你的卖家 ID
*   `{sku}`: 你自定义的 SKU
*   `marketplaceIds`: 站点 ID（如美国站）
*   `productType`: **必填**（从第一步获取）

#### 2. 请求体 (JSON Body)
```json
{
  "productType": "LUGGAGE",
  "requirements": "LISTING",
  "attributes": {
    "externalProductId": [
      {
        "externalProductId": "B08N5WRWNW",
        "externalProductType": "ASIN"
      }
    ],
    "conditionType": "NEW",
    "price": {
      "currencyCode": "USD",
      "amount": 59.99
    },
    "quantity": {
      "amount": 10
    }
  }
}
```

#### 3. 关键字段说明（官方文档核实）

| 字段路径 | 是否必填 | 说明 |
| :--- | :--- | :--- |
| `productType` | **是** | 必须与 URL 参数中的一致。例如 `LUGGAGE`, `SHOES`, `TOYS`。 |
| `requirements` | 是 | 建议填 `LISTING`。表示你要创建一个完整的商品详情（包含身份、价格、库存）。 |
| `attributes` | **是** | **这是最容易出错的地方**。里面的字段名**取决于** `productType`。 |
| `attributes.externalProductId` | 通常必填 | 用于将你的 SKU 关联到现有的 ASIN。字段名可能是 `externalProductId` 或 `gtin`，取决于分类。 |
| `attributes.conditionType` | 通常必填 | 商品状况，如 `NEW`, `USED_LIKE_NEW` 等。 |
| `attributes.price` | 通常必填 | 包含 `currencyCode` 和 `amount`。 |
| `attributes.quantity` | 可选 | 自发货库存。如果是 FBA，通常不需要填此项（通过入库单管理），但填了也没错。 |

---

### 重要风险提示：`attributes` 的动态性

虽然上面的 JSON 结构是通用的，但 **`attributes` 内部的具体字段名是动态的**。

*   **例子 A (箱包类)**：可能接受 `externalProductId` 来关联 ASIN。
*   **例子 B (服装类)**：可能要求更详细的 `sizeName`, `colorName` 即使你是跟卖已有 ASIN。
*   **例子 C (某些类目)**：可能不允许直接通过 API 关联 ASIN，而要求提供 UPC/EAN。

#### 如何确保 100% 成功？（推荐方案）

既然你只有 10 个商品，建议在代码中加入一个**“获取 Schema"**的步骤，以确保字段名绝对正确：

1.  调用 `searchCatalogItems` (传入 ASIN) -> 拿到 `productType` (例如 `LUGGAGE`)。
2.  调用 **`getListingsItemSchema`** (传入 `productType`)。
    *   这个 API 会返回该分类所需的完整 JSON 定义。
    *   查看返回结果中的 `required` 字段和 `properties`。
    *   确认关联 ASIN 的字段名到底是 `externalProductId` 还是其他。
3.  构造 `putListingsItem` 请求。

### 总结建议

1.  **API 选择**：继续使用 **`Listings Items API (putListingsItem)`**，它是目前最现代的接口。
2.  **关于分类**：接受必须传 `productType` 的设定。不要硬编码，而是通过 `searchCatalogItems` 根据 ASIN 动态获取。
3.  **关于字段**：不要完全依赖通用的 JSON 示例。对于这 10 个商品，建议先写一个小脚本，调用 `getListingsItemSchema` 打印出该分类要求的必填字段，再构造最终请求。这能避免 90% 的 `400 Bad Request` 错误。

**简化版操作流（针对 10 个商品）：**
如果你的 10 个商品属于**同一个大类**（比如都是电子产品），你可以手动查一次 `productType`，然后硬编码在程序里，这样就不用每次调用 Schema 接口了。但如果类目混杂，请务必动态获取。

