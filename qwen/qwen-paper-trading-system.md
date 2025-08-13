# 我想编写一个模拟盘交易系统，给交易初学者提供一个练习的机会，提供真实的行情，但资金都是虚拟的，初学者可以通过这样的系统练习对行情走势的判断能力，请帮我设计这样的系统，需要考虑哪些事情，要点和难点，列出可供参考的开源系统

这是一个非常有价值的项目，既能帮助初学者低成本试错，又能作为你展示全栈开发能力的作品。设计一个模拟盘交易系统（Paper Trading System）看似简单，但要做到“逼真”且“有用”，需要在数据、逻辑、体验和合规上下功夫。

以下是为您设计的系统方案，涵盖架构、核心要点、难点及开源参考。

---

### 一、系统架构设计

建议采用 **前后端分离** 的架构，核心在于**低延迟的数据处理**和**准确的撮合逻辑**。

```mermaid
graph TD
    User[用户终端 (Web/App)] --> API[API 网关]
    API --> Auth[认证服务]
    API --> Trade[交易核心服务]
    API --> Market[行情服务]
    
    subgraph 数据层
    Feed[外部行情源 (交易所/数据商)] --> DataParser[数据解析/清洗]
    DataParser --> Redis[实时缓存 (Redis)]
    DataParser --> DB[历史数据库 (TimescaleDB/InfluxDB)]
    end
    
    subgraph 核心逻辑
    Trade --> Match[模拟撮合引擎]
    Match --> Account[账户/资金管理]
    Match --> Risk[风控/限额检查]
    end
    
    Match --> DB
    Account --> DB
```

#### 核心模块划分：
1.  **行情接入模块 (Market Data Feed)**：负责从外部获取实时行情（WebSocket 优先）。
2.  **模拟撮合引擎 (Simulation Matching Engine)**：核心中的核心，根据行情判断用户的挂单是否成交。
3.  **账户管理系统 (Account & Ledger)**：管理虚拟资金、持仓、冻结资金、流水记录。
4.  **用户交互端 (UI/UX)**：K 线图、下单面板、持仓列表、盈亏分析。
5.  **教学与反馈模块 (Education)**：交易复盘、错误提示、知识点推送。

---

### 二、需要考虑的关键事项 (Key Considerations)

#### 1. 行情数据的真实性与成本
*   **数据源**：
    *   **加密货币**：数据公开免费（如 Binance, OKX 的 WebSocket API），最适合起步。
    *   **股票 (A 股/美股)**：实时 Level-2 行情通常收费昂贵。初学者系统可使用**延时 15 分钟**的免费数据，或购买合规的聚合数据接口（如聚宽、TuShare 等）。
*   **数据清洗**：外部数据可能有噪点（如瞬间插针），系统需要过滤异常值，避免导致模拟盘盈亏失真。

#### 2. 撮合逻辑的逼真度 (最关键)
如果撮合太简单，用户会养成坏习惯。
*   **市价单 (Market Order)**：应模拟**滑点 (Slippage)**。不能简单按当前最新价成交，应根据订单大小模拟对盘口的冲击。
*   **限价单 (Limit Order)**：
    *   *简单模式*：当行情最新价触及挂单价即成交。
    *   *进阶模式*：需要维护一个**虚拟订单簿 (Virtual Order Book)**。只有当买一价 >= 卖一价，且你的挂单在队列中时才能成交。
*   **交易费用**：必须模拟手续费、印花税、资金费率（如果是合约），否则用户无法理解真实成本。

#### 3. 用户体验与心理模拟
*   **初始资金**：建议提供不同档位（如 1 万、10 万、100 万），让用户感受仓位管理的区别。
*   **强制平仓**：如果是杠杆/合约模拟，必须严格执行强平逻辑，让用户体验爆仓的风险。
*   **交易日记**：强制或鼓励用户记录开仓理由，系统后期可生成“交易心理报告”。

#### 4. 安全性与防作弊
*   虽然是虚拟资金，但要防止用户通过篡改本地时间、重放请求等方式刷排行榜。
*   所有逻辑必须在**服务端**完成，前端只负责展示。
*   接口需做频率限制 (Rate Limiting)。

---

### 三、技术难点 (Difficulties)

1.  **高并发下的数据一致性**
    *   **难点**：行情剧烈波动时，大量用户同时下单。
    *   **解决**：使用消息队列（Kafka/RabbitMQ）削峰；资金扣减使用数据库乐观锁或 Redis Lua 脚本保证原子性。
2.  **撮合引擎的性能**
    *   **难点**：如果采用虚拟订单簿撮合，内存消耗大，匹配速度慢。
    *   **解决**：对于初学者系统，可采用“最新价 + 随机滑点”的简化模型；若追求专业，需使用内存数据库或 Go/C++ 编写撮合核心。
3.  **状态同步**
    *   **难点**：前端 K 线图、持仓盈亏、账户余额需要实时同步。
    *   **解决**：大量使用 WebSocket 推送，而非 HTTP 轮询。
4.  **合规与法律风险 (中国大陆特别重要)**
    *   **难点**：在中国，提供证券/期货行情和交易模拟可能涉及**金融信息服务牌照**。
    *   **解决**：
        *   明确标注“模拟交易，不构成投资建议”。
        *   避免涉及真实法币出入金通道（纯虚拟币）。
        *   若面向公众运营，务必咨询法律顾问，建议初期仅作为个人技术作品或小范围内部工具。

---

### 四、技术栈推荐

*   **后端**：
    *   语言：**Go** (高并发撮合) 或 **Python** (生态丰富，vn.py 基于 Python)。
    *   数据库：**PostgreSQL** (用户/订单数据) + **Redis** (缓存/会话) + **TimescaleDB/InfluxDB** (K 线时序数据)。
    *   通信：**WebSocket** (行情推送) + **gRPC/REST** (业务接口)。
*   **前端**：
    *   框架：**React** 或 **Vue3**。
    *   图表：**TradingView Lightweight Charts** (免费、性能好、专业) 或 **ECharts**。
*   **部署**：Docker + Kubernetes (便于扩展)。

---

### 五、可供参考的开源系统

完全符合你需求的“开箱即用”开源模拟盘较少，大多是量化框架或交易所核心，你可以基于它们进行改造：

#### 1. vn.py (Python)
*   **简介**：中国最流行的开源量化交易框架。
*   **参考价值**：它内置了 `Simulator` 模块，支持基于历史数据或实时数据的模拟交易。你可以参考它的**事件驱动引擎**设计和**撮合逻辑**。
*   **地址**：`github.com/vnpy/vnpy`

#### 2. Hummingbot (Python/Cython)
*   **简介**：主打加密货币的高频做市机器人。
*   **参考价值**：它有一个 `paper_trade` 模式，可以在不连接真实交易所的情况下模拟订单簿和成交。其**订单簿管理逻辑**非常值得学习。
*   **地址**：`github.com/hummingbot/hummingbot`

#### 3. QuantConnect / Lean Engine (C#)
*   **简介**：国外著名的量化平台，其引擎 Lean 是开源的。
*   **参考价值**：企业级的回测和实盘引擎，数据管理和撮合非常严谨，但架构较重，适合学习设计理念。
*   **地址**：`github.com/QuantConnect/Lean`

#### 4. Superalgos (Node.js)
*   **简介**：免费的加密资产自动化交易和回测平台。
*   **参考价值**：提供完整的前端界面和后端逻辑，包含模拟交易功能，架构比较现代。
*   **地址**：`github.com/superalgos/superalgos`

#### 5. 轻量级图表库：TradingView Lightweight Charts
*   **简介**：不是交易系统，但它是构建交易界面的标准。
*   **参考价值**：如何绘制专业的 K 线图、成交量、指标。
*   **地址**：`github.com/tradingview/lightweight-charts`

---

### 六、开发路线图建议 (MVP 版本)

不要试图一开始就做一个完美的系统，建议分步走：

1.  **阶段一：数据与图表 (Data & Chart)**
    *   接入一个免费的 Crypto WebSocket 行情源。
    *   前端实现 K 线图实时跳动。
    *   *目标*：看到真实的价格在动。
2.  **阶段二：基础交易逻辑 (Basic Trading)**
    *   实现用户注册、虚拟账户初始化。
    *   实现“市价买入/卖出”，按当前最新价成交，扣除手续费。
    *   实现持仓显示和简单盈亏计算。
    *   *目标*：能买能卖，资金对得上。
3.  **阶段三：进阶撮合与风控 (Advanced Matching)**
    *   加入限价单（Limit Order）逻辑。
    *   加入滑点模拟。
    *   加入每日结算单、交易历史导出。
    *   *目标*：体验接近真实交易。
4.  **阶段四：社区与教学 (Community)**
    *   增加排行榜（注意防作弊）。
    *   增加“交易笔记”功能。
    *   *目标*：增加用户粘性，实现教育目的。

### 七、给初学者的特别提示

在系统设计中，你可以加入一些**“教育性干预”**，这是区别于普通模拟盘的地方：
*   **风险警示**：当用户仓位过重（如单笔超过 50% 资金）时，弹出提示“建议分散投资”。
*   **止损提醒**：当用户开仓未设止损且亏损扩大时，提示“设置止损是保护本金的关键”。
*   **复盘工具**：允许用户回放自己的交易记录，并在 K 图上标记出当时的买卖点，分析当时的心态。

### 总结
开发模拟盘系统的技术门槛中等，但**业务逻辑的严谨性**要求很高。建议从**加密货币**入手（数据免费、API 友好、合规风险相对低），使用 **Python (vn.py 架构思路) + Vue/React + TradingView 图表** 进行快速构建。务必在显著位置标注“模拟训练，不涉及真实资金”，以规避法律风险。

# 如何设计数据库系统表结构，存储各种数据

设计模拟盘交易系统的数据库是核心环节，需要兼顾**数据一致性**（资金不能出错）、**查询性能**（实时看盘、查历史）和**扩展性**（支持多品种、多用户）。

建议采用 **关系型数据库 (PostgreSQL/MySQL)** 存储业务数据，配合 **时序数据库 (TimescaleDB/InfluxDB)** 或 **宽表** 存储行情数据。

以下是详细的表结构设计方案：

---

### 一、核心设计理念

1.  **精度控制**：所有金额、价格字段必须使用 `DECIMAL` 类型（如 `DECIMAL(20, 8)`），严禁使用 `FLOAT` 或 `DOUBLE`，避免浮点数计算误差导致资金对不上。
2.  **双记账原则**：资金变动必须有对应的流水记录（Ledger），确保 `账户余额 = 初始资金 + 所有流入 - 所有流出`。
3.  **状态机管理**：订单状态（待成交、部分成交、已成交、已撤单、已取消）必须清晰定义，防止状态混乱。
4.  **读写分离**：高频的行情写入和低频的用户查询分开处理。

---

### 二、详细表结构设计 (ER 模型)

#### 1. 用户与账户模块 (User & Account)

**`users` (用户基础信息表)**
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT (PK) | 用户 ID |
| `username` | VARCHAR | 用户名 |
| `password_hash` | VARCHAR | 加密密码 |
| `email` | VARCHAR | 邮箱 |
| `created_at` | TIMESTAMP | 注册时间 |
| `status` | TINYINT | 状态 (1:正常，0:冻结) |

**`accounts` (资金账户表)**
*每个用户可能有一个主账户，或者按币种分账户。这里简化为单币种或多币种混合模式。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT (PK) | 账户 ID |
| `user_id` | BIGINT (FK) | 关联用户 ID |
| `currency` | VARCHAR | 币种 (e.g., USDT, CNY) |
| `balance_total` | DECIMAL(20,8) | **总资产** (可用 + 冻结) |
| `balance_available`| DECIMAL(20,8) | **可用余额** (可下单的资金) |
| `balance_frozen` | DECIMAL(20,8) | **冻结余额** (挂单未成交占用的资金) |
| `updated_at` | TIMESTAMP | 最后更新时间 |
| *索引* | `(user_id, currency)` | 唯一索引 |

**`account_ledgers` (资金流水表 - 核心审计表)**
*任何余额变动都必须在此插入一条记录，不可修改，只能追加。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT (PK) | 流水 ID |
| `account_id` | BIGINT (FK) | 关联账户 ID |
| `type` | VARCHAR | 类型 (DEPOSIT:充值，WITHDRAW:提现，TRADE_PROFIT:交易盈亏，FEE:手续费，FREEZE:冻结，UNFREEZE:解冻) |
| `amount` | DECIMAL(20,8) | 变动金额 (+/-) |
| `balance_after` | DECIMAL(20,8) | 变动后的余额快照 (用于快速核对) |
| `related_order_id` | BIGINT | 关联订单 ID (如果是交易产生) |
| `description` | VARCHAR | 备注描述 |
| `created_at` | TIMESTAMP | 发生时间 |

---

#### 2. 市场与标的模块 (Market & Instruments)

**`instruments` (交易标的表)**
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `symbol` | VARCHAR (PK) | 交易对代码 (e.g., BTCUSDT, 600519.SH) |
| `name` | VARCHAR | 名称 (e.g., 比特币，贵州茅台) |
| `base_currency` | VARCHAR | 基础币种 (e.g., BTC) |
| `quote_currency` | VARCHAR | 计价币种 (e.g., USDT) |
| `price_tick` | DECIMAL(20,8) | 最小价格变动单位 |
| `volume_tick` | DECIMAL(20,8) | 最小交易量单位 |
| `min_order_size` | DECIMAL(20,8) | 最小下单数量 |
| `maker_fee_rate` | DECIMAL(10,6) | 挂单手续费率 (0.001 表示 0.1%) |
| `taker_fee_rate` | DECIMAL(10,6) | 吃单手续费率 |
| `is_active` | BOOLEAN | 是否停牌/交易中 |

---

#### 3. 交易核心模块 (Trading Core)

**`orders` (订单表)**
*记录用户的每一笔委托。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `order_id` | VARCHAR/BIGINT (PK)| 订单号 (建议使用雪花算法生成字符串) |
| `user_id` | BIGINT (FK) | 用户 ID |
| `symbol` | VARCHAR (FK) | 交易对 |
| `side` | VARCHAR | 方向 (BUY, SELL) |
| `type` | VARCHAR | 类型 (LIMIT:限价，MARKET:市价) |
| `price` | DECIMAL(20,8) | 委托价格 (市价单可为 NULL 或 0) |
| `quantity` | DECIMAL(20,8) | 委托数量 |
| `filled_quantity` | DECIMAL(20,8) | 已成交数量 |
| `remaining_quantity`| DECIMAL(20,8) | 剩余数量 (冗余字段，方便计算) |
| `status` | VARCHAR | 状态 (PENDING, PARTIAL_FILLED, FILLED, CANCELLED, REJECTED) |
| `time_in_force` | VARCHAR | 有效期策略 (GTC:有效直到取消，IOC:立即成交否则取消) |
| `created_at` | TIMESTAMP | 下单时间 |
| `updated_at` | TIMESTAMP | 最后更新时间 |
| *索引* | `(user_id, status, created_at)` | 常用查询组合 |
| *索引* | `(symbol, status)` | 撮合引擎可能需要 |

**`trades` (成交明细表)**
*订单撮合成功后产生的具体成交记录。一个订单可能对应多条成交记录（分批成交）。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `trade_id` | BIGINT (PK) | 成交 ID |
| `order_id` | VARCHAR (FK) | 关联订单 ID |
| `symbol` | VARCHAR | 交易对 |
| `user_id` | BIGINT (FK) | 用户 ID |
| `counter_party_id`| BIGINT | 对手方 ID (模拟盘中通常是系统做市商，可填固定值或 NULL) |
| `price` | DECIMAL(20,8) | **成交价格** |
| `quantity` | DECIMAL(20,8) | 成交数量 |
| `fee` | DECIMAL(20,8) | 本次成交产生的手续费 |
| `fee_currency` | VARCHAR | 手续费币种 |
| `role` | VARCHAR | 角色 (MAKER:挂单方，TAKER:吃单方) |
| `created_at` | TIMESTAMP | 成交时间 |
| *索引* | `(order_id)` | 查订单的成交明细 |
| *索引* | `(user_id, created_at)` | 查用户的交易历史 |

**`positions` (持仓表)**
*记录用户当前的持仓情况。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT (PK) | 主键 |
| `user_id` | BIGINT (FK) | 用户 ID |
| `symbol` | VARCHAR | 交易对 |
| `direction` | VARCHAR | 方向 (LONG:多头，SHORT:空头 - 若支持双向) |
| `quantity` | DECIMAL(20,8) | 持仓数量 |
| `avg_entry_price` | DECIMAL(20,8) | **平均开仓成本** (加权平均) |
| `unrealized_pnl` | DECIMAL(20,8) | 浮动盈亏 (根据最新价实时计算，可不存库或定期更新) |
| `updated_at` | TIMESTAMP | 最后更新时间 |
| *唯一索引* | `(user_id, symbol, direction)` | 确保同一用户对同一标的只有一个持仓记录 |

---

#### 4. 行情数据存储 (Market Data)

*注意：高频行情数据量大，不建议直接存入 MySQL 的主表中，推荐使用 TimescaleDB (基于 PG 的插件) 或 InfluxDB。如果必须用 MySQL，需按月分表。*

**`klines_1m` (1 分钟 K 线表 - 示例)**
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `symbol` | VARCHAR | 交易对 |
| `timestamp` | BIGINT/TIMESTAMP | 时间戳 (分区键) |
| `open` | DECIMAL(20,8) | 开盘价 |
| `high` | DECIMAL(20,8) | 最高价 |
| `low` | DECIMAL(20,8) | 最低价 |
| `close` | DECIMAL(20,8) | 收盘价 |
| `volume` | DECIMAL(20,8) | 成交量 |
| *索引* | `(symbol, timestamp)` | 复合索引 |

**`tick_data` (Tick 级数据表 - 可选，数据量极大)**
*仅当需要极高精度的回测或复盘时才存储，通常只保留最近 N 天。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `symbol` | VARCHAR | ... |
| `timestamp` | BIGINT | ... |
| `price` | DECIMAL | ... |
| `volume` | DECIMAL | ... |
| `bid_price` | DECIMAL | 买一价 |
| `ask_price` | DECIMAL | 卖一价 |

---

#### 5. 教学与辅助模块 (Education & Misc)

**`user_notes` (交易日记)**
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | BIGINT | ... |
| `user_id` | BIGINT | ... |
| `order_id` | VARCHAR | 关联的订单 (可选) |
| `content` | TEXT | 日记内容 (为什么买？心情如何？) |
| `tags` | JSON | 标签 (e.g., ["追高", "止损"]) |
| `created_at` | TIMESTAMP | ... |

**`system_config` (系统配置)**
*用于动态调整模拟盘参数，如初始资金、手续费打折活动等。*
| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| `key` | VARCHAR (PK) | 配置键 (e.g., default_initial_balance) |
| `value` | VARCHAR/JSON | 配置值 |
| `description` | VARCHAR | 描述 |

---

### 三、关键业务流程的 SQL 逻辑示例

#### 1. 下单并冻结资金 (原子操作)
当用户发起买入限价单时，需要在事务中完成：
1. 检查可用余额是否充足。
2. 扣除可用余额，增加冻结余额。
3. 插入订单记录 (Status: PENDING)。
4. 插入资金流水 (Type: FREEZE)。

```sql
BEGIN;

-- 1. 更新账户余额 (使用乐观锁 version 或 FOR UPDATE)
UPDATE accounts 
SET balance_available = balance_available - :total_cost,
    balance_frozen = balance_frozen + :total_cost,
    updated_at = NOW()
WHERE user_id = :uid AND currency = 'USDT' 
  AND balance_available >= :total_cost;

-- 检查受影响行数，若为 0 则抛出异常（余额不足）

-- 2. 插入订单
INSERT INTO orders (order_id, user_id, symbol, side, type, price, quantity, status, created_at)
VALUES (:oid, :uid, :sym, 'BUY', 'LIMIT', :price, :qty, 'PENDING', NOW());

-- 3. 记录流水
INSERT INTO account_ledgers (account_id, type, amount, balance_after, related_order_id, created_at)
SELECT id, 'FREEZE', -(:total_cost), balance_available - :total_cost, :oid, NOW()
FROM accounts WHERE user_id = :uid AND currency = 'USDT';

COMMIT;
```

#### 2. 成交处理 (撮合成功后)
当模拟引擎判定订单成交时：
1. 将冻结资金转为实际持仓成本（或扣除手续费后剩余退回可用）。
2. 更新订单状态和部分成交数量。
3. 插入成交记录 (`trades`)。
4. 更新或插入持仓表 (`positions`)，重新计算平均成本。
5. 记录手续费流水。

---

### 四、优化与建议

1.  **分库分表策略**：
    *   `trades` 和 `account_ledgers` 表增长最快。当数据量超过千万级时，建议按 `user_id` 哈希分表，或按 `created_at` 月份进行范围分区。
2.  **缓存层 (Redis)**：
    *   **账户余额**：高频读取，建议缓存在 Redis 中，异步持久化到 DB，但涉及资金扣减必须走 DB 事务保证一致性。
    *   **最新行情**：所有 K 线和 Tick 数据先写 Redis，前端通过 WebSocket 从 Redis 订阅，DB 仅做归档。
    *   **订单簿**：模拟撮合引擎的内存订单簿可以定期快照到 Redis。
3.  **数据归档**：
    *   对于初学者，可能不需要保存几年的 Tick 数据。设置定时任务，将 3 个月前的明细数据迁移到冷存储（如 AWS S3 或专门的归档库），保持主库轻量。
4.  **软删除**：
    *   用户注销或数据清理时，尽量使用 `is_deleted` 标记，而不是物理删除，以便审计和恢复。

这个 schema 设计既满足了 ACID 特性以保证资金安全，又通过合理的索引和分离设计支撑了高并发查询，非常适合作为模拟盘系统的基石。


# 用golang编写撮合交易系统

用 Go 语言编写撮合引擎（Matching Engine）是极佳的选择，因为 Go 拥有强大的并发模型（Goroutines/Channels）和高性能，非常适合处理高吞吐的交易请求。

下面我将为你提供一个**核心撮合引擎的最小可行产品 (MVP)** 代码实现。这个系统包含：
1.  **订单簿 (OrderBook)**：维护买单和卖单。
2.  **限价单撮合逻辑**：价格优先、时间优先。
3.  **并发安全**：使用 Mutex 保护共享状态。
4.  **事件驱动**：成交后生成 `Trade` 事件。

### 1. 核心数据结构设计

首先定义订单、订单簿和交易结果的结构体。

```go
package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// Side 代表交易方向
type Side int

const (
	Buy Side = iota + 1
	Sell
)

func (s Side) String() string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}

// Order 代表一个订单
type Order struct {
	ID        string
	UserID    int64
	Symbol    string
	Side      Side
	Price     float64 // 使用 float64 演示，生产环境建议用 decimal 库或整数 (分/聪)
	Quantity  float64
	Timestamp time.Time
}

// Trade 代表一笔成交记录
type Trade struct {
	Price       float64
	Quantity    float64
	BuyerOrderID  string
	SellerOrderID string
	BuyerUserID   int64
	SellerUserID  int64
	Timestamp     time.Time
}

// OrderBook 订单簿结构
type OrderBook struct {
	Symbol string
	// 买单列表：按价格降序排列 (高价优先)，同价格按时间升序
	bids *list.List 
	// 卖单列表：按价格升序排列 (低价优先)，同价格按时间升序
	asks *list.List
	
	// 为了快速查找特定价格的订单队列，实际生产中通常用 map[price]*list.List
	// 这里为了简化 MVP，我们直接在列表中遍历或使用更简单的逻辑
	// 注意：下面的实现为了代码简洁，采用了线性搜索匹配，高性能版本需优化索引
	
	mu sync.Mutex
}

// NewOrderBook 创建新的订单簿
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		bids:   list.New(),
		asks:   list.New(),
	}
}
```

### 2. 撮合引擎核心逻辑

这是最关键的部分：`MatchOrder` 函数。它接收新订单，尝试与对手盘匹配，剩余部分进入订单簿。

```go
// MatchOrder 处理新订单，返回成交列表和剩余订单（如果未完全成交）
func (ob *OrderBook) MatchOrder(order *Order) ([]Trade, *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []Trade
	remainingQty := order.Quantity

	// 根据订单方向选择对手盘列表
	var opponentList *list.List
	var matchCondition func(o *Order) bool

	if order.Side == Buy {
		// 买单：去卖单池 (asks) 找价格 <= 买单价格的订单
		opponentList = ob.asks
		matchCondition = func(o *Order) bool {
			return o.Price <= order.Price
		}
	} else {
		// 卖单：去买单池 (bids) 找价格 >= 卖单价格的订单
		opponentList = ob.bids
		matchCondition = func(o *Order) bool {
			return o.Price >= order.Price
		}
	}

	// 遍历对手盘进行撮合
	// 注意：list 遍历中删除元素需要小心处理指针
	var nextElement *list.Element
	for e := opponentList.Front(); e != nil && remainingQty > 1e-8; e = nextElement { // 浮点数精度容错
		nextElement = e.Next()
		opponentOrder := e.Value.(*Order)

		if !matchCondition(opponentOrder) {
			// 价格不满足，停止撮合 (因为列表已排序，后面的更不可能满足)
			break
		}

		// 计算成交数量
		fillQty := min(remainingQty, opponentOrder.Quantity)

		// 生成成交记录
		trade := Trade{
			Price:         opponentOrder.Price, // 成交价通常为对手盘挂单价格 (Price Time Priority)
			Quantity:      fillQty,
			BuyerOrderID:  getOrderID(order, opponentOrder, Buy),
			SellerOrderID: getOrderID(order, opponentOrder, Sell),
			BuyerUserID:   getUserID(order, opponentOrder, Buy),
			SellerUserID:  getUserID(order, opponentOrder, Sell),
			Timestamp:     time.Now(),
		}
		trades = append(trades, trade)

		// 更新剩余数量
		remainingQty -= fillQty
		opponentOrder.Quantity -= fillQty

		// 如果对手单完全成交，从列表中移除
		if opponentOrder.Quantity < 1e-8 {
			opponentList.Remove(e)
			fmt.Printf("[MATCH] 完全成交：订单 %s vs %s, 价格 %.2f, 数量 %.2f\n", 
				trade.BuyerOrderID, trade.SellerOrderID, trade.Price, trade.Quantity)
		} else {
			fmt.Printf("[MATCH] 部分成交：订单 %s vs %s, 价格 %.2f, 数量 %.2f, 对手单剩余 %.2f\n", 
				trade.BuyerOrderID, trade.SellerOrderID, trade.Price, trade.Quantity, opponentOrder.Quantity)
		}
	}

	// 如果还有剩余数量，将订单加入订单簿
	var remainingOrder *Order
	if remainingQty > 1e-8 {
		order.Quantity = remainingQty
		order.Timestamp = time.Now() // 更新时间为当前入池时间
		
		if order.Side == Buy {
			insertSorted(ob.bids, order, func(a, b *Order) bool {
				// 买单：价格高优先；价格相同，时间早优先
				if a.Price != b.Price {
					return a.Price > b.Price
				}
				return a.Timestamp.Before(b.Timestamp)
			})
		} else {
			insertSorted(ob.asks, order, func(a, b *Order) bool {
				// 卖单：价格低优先；价格相同，时间早优先
				if a.Price != b.Price {
					return a.Price < b.Price
				}
				return a.Timestamp.Before(b.Timestamp)
			})
		}
		remainingOrder = order
		fmt.Printf("[ORDERBOOK] 订单 %s 挂单成功，剩余数量 %.2f, 价格 %.2f\n", order.ID, order.Quantity, order.Price)
	} else {
		fmt.Printf("[FULLY FILLED] 订单 %s 全部成交\n", order.ID)
	}

	return trades, remainingOrder
}

// 辅助函数：获取买方 ID
func getOrderID(newOrder, oppOrder *Order, side Side) string {
	if side == Buy {
		if newOrder.Side == Buy {
			return newOrder.ID // 这种情况极少，通常是吃单
		}
		return oppOrder.ID
	}
	if newOrder.Side == Sell {
		return newOrder.ID
	}
	return oppOrder.ID
}
// 修正上面的逻辑，简化为：
func getBuyerID(newOrder, oppOrder *Order) int64 {
	if newOrder.Side == Buy {
		return newOrder.UserID
	}
	return oppOrder.UserID
}
func getSellerID(newOrder, oppOrder *Order) int64 {
	if newOrder.Side == Sell {
		return newOrder.UserID
	}
	return oppOrder.UserID
}
// 重写 Trade 生成部分的 ID 获取逻辑以匹配上面
// (在实际代码中会直接替换上面 MatchOrder 中的对应行)

// 辅助函数：向有序链表插入节点
func insertSorted(l *list.List, order *Order, less func(a, b *Order) bool) {
	for e := l.Front(); e != nil; e = e.Next() {
		existing := e.Value.(*Order)
		if less(order, existing) {
			l.InsertBefore(order, e)
			return
		}
	}
	l.PushBack(order)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
```

*注意：上面的 `getOrderID` 逻辑有点混乱，我在下面的完整可运行代码中会修正它。*

### 3. 完整可运行示例 (`main.go`)

将上述逻辑整合，并修复 ID 逻辑，添加模拟测试。

```go
package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// --- 数据类型定义 ---

type Side int

const (
	Buy Side = iota + 1
	Sell
)

func (s Side) String() string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}

type Order struct {
	ID        string
	UserID    int64
	Symbol    string
	Side      Side
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

type Trade struct {
	Price       float64
	Quantity    float64
	BuyerOrderID  string
	SellerOrderID string
	BuyerUserID   int64
	SellerUserID  int64
	Timestamp     time.Time
}

// --- 订单簿实现 ---

type OrderBook struct {
	Symbol string
	bids   *list.List // 买单：高价在前
	asks   *list.List // 卖单：低价在前
	mu     sync.Mutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		bids:   list.New(),
		asks:   list.New(),
	}
}

// PlaceOrder 外部调用入口
func (ob *OrderBook) PlaceOrder(order *Order) []Trade {
	trades, _ := ob.MatchOrder(order)
	return trades
}

func (ob *OrderBook) MatchOrder(order *Order) ([]Trade, *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []Trade
	remainingQty := order.Quantity
	order.Timestamp = time.Now() // 确保有时间戳用于排序

	// 确定对手盘列表和匹配条件
	var opponentList *list.List
	var isMatch func(opponent *Order) bool

	if order.Side == Buy {
		opponentList = ob.asks
		isMatch = func(o *Order) bool { return o.Price <= order.Price }
	} else {
		opponentList = ob.bids
		isMatch = func(o *Order) bool { return o.Price >= order.Price }
	}

	// 撮合循环
	var next *list.Element
	for e := opponentList.Front(); e != nil && remainingQty > 0.00000001; e = next {
		next = e.Next()
		opponent := e.Value.(*Order)

		if !isMatch(opponent) {
			break // 价格不优，停止
		}

		// 计算成交量
		fillQty := opponent.Quantity
		if remainingQty < fillQty {
			fillQty = remainingQty
		}

		// 构造成交记录
		var buyerID, sellerID int64
		var buyerOID, sellerOID string

		if order.Side == Buy {
			buyerID, buyerOID = order.UserID, order.ID
			sellerID, sellerOID = opponent.UserID, opponent.ID
		} else {
			buyerID, buyerOID = opponent.UserID, opponent.ID
			sellerID, sellerOID = order.UserID, order.ID
		}

		trade := Trade{
			Price:         opponent.Price, // 以挂单方价格成交
			Quantity:      fillQty,
			BuyerOrderID:  buyerOID,
			SellerOrderID: sellerOID,
			BuyerUserID:   buyerID,
			SellerUserID:  sellerID,
			Timestamp:     time.Now(),
		}
		trades = append(trades, trade)

		// 更新数量
		remainingQty -= fillQty
		opponent.Quantity -= fillQty

		fmt.Printf("✅ 成交：[%s] %s vs %s | 价格：%.2f | 数量：%.2f\n", 
			ob.Symbol, buyerOID, sellerOID, trade.Price, trade.Quantity)

		// 如果对手单被吃完，移除
		if opponent.Quantity < 0.00000001 {
			opponentList.Remove(e)
		}
	}

	// 处理剩余订单
	if remainingQty > 0.00000001 {
		order.Quantity = remainingQty
		// 插入订单簿
		if order.Side == Buy {
			insertSorted(ob.bids, order, func(a, b *Order) bool {
				if a.Price != b.Price {
					return a.Price > b.Price // 买：价高优先
				}
				return a.Timestamp.Before(b.Timestamp) // 同价：时早优先
			})
		} else {
			insertSorted(ob.asks, order, func(a, b *Order) bool {
				if a.Price != b.Price {
					return a.Price < b.Price // 卖：价低优先
				}
				return a.Timestamp.Before(b.Timestamp)
			})
		}
		fmt.Printf("📝 挂单：[%s] 订单 %s (%s) 进入簿，价格 %.2f, 剩余 %.2f\n", 
			ob.Symbol, order.ID, order.Side, order.Price, order.Quantity)
		return trades, order
	}

	fmt.Printf("🎉 完单：订单 %s 全部成交\n", order.ID)
	return trades, nil
}

// 有序插入辅助函数
func insertSorted(l *list.List, order *Order, less func(a, b *Order) bool) {
	for e := l.Front(); e != nil; e = e.Next() {
		curr := e.Value.(*Order)
		if less(order, curr) {
			l.InsertBefore(order, e)
			return
		}
	}
	l.PushBack(order)
}

// --- 主函数测试 ---

func main() {
	fmt.Println("🚀 启动 Go 模拟撮合引擎...")
	ob := NewOrderBook("BTC/USDT")

	// 场景 1: 用户 A 挂出卖单 (Maker)
	// 卖出 1 BTC @ 50000
	orderA := &Order{ID: "ORD-A-001", UserID: 101, Symbol: "BTC/USDT", Side: Sell, Price: 50000, Quantity: 1.0}
	ob.PlaceOrder(orderA)

	// 场景 2: 用户 B 挂出卖单 (Maker)
	// 卖出 0.5 BTC @ 50100
	orderB := &Order{ID: "ORD-B-002", UserID: 102, Symbol: "BTC/USDT", Side: Sell, Price: 50100, Quantity: 0.5}
	ob.PlaceOrder(orderB)

	fmt.Println("\n--- 此时订单簿状态 ---")
	fmt.Println("卖单 (Asks): 50000 (1.0), 50100 (0.5)")
	fmt.Println("买单 (Bids): 空")
	fmt.Println("--------------------\n")

	// 场景 3: 用户 C 发起买入 (Taker)
	// 买入 1.2 BTC，最高愿意出 50200
	// 预期：
	// 1. 吃掉 ORD-A (1.0 @ 50000)
	// 2. 吃掉 ORD-B (0.2 @ 50100)，ORD-B 剩余 0.3
	orderC := &Order{ID: "ORD-C-003", UserID: 103, Symbol: "BTC/USDT", Side: Buy, Price: 50200, Quantity: 1.2}
	trades := ob.PlaceOrder(orderC)

	fmt.Printf("\n📊 本次共产生 %d 笔成交\n", len(trades))
	
	// 场景 4: 用户 D 低价买入，无法成交，进入挂单
	orderD := &Order{ID: "ORD-D-004", UserID: 104, Symbol: "BTC/USDT", Side: Buy, Price: 49000, Quantity: 2.0}
	ob.PlaceOrder(orderD)
}
```

### 4. 代码关键点解析

1.  **线程安全 (`sync.Mutex`)**:
    *   在 `MatchOrder` 开始时加锁，结束时解锁。这确保了在同一时刻只有一个订单在处理，避免了“超卖”或资金数据竞争。
    *   *进阶优化*：如果是超高并发，可以对不同的 `Symbol` 使用不同的锁（Sharding），或者使用无锁队列 + 单线程处理模型（类似 Disruptor 模式）。

2.  **价格优先与时间优先**:
    *   通过 `insertSorted` 函数实现。
    *   **买单**：价格高的排在前面（`Front`），价格相同时时间早的排在前面。
    *   **卖单**：价格低的排在前面，价格相同时时间早的排在前面。
    *   撮合时直接从 `Front()` 开始遍历，天然符合优先级规则。

3.  **浮点数精度警告**:
    *   代码中使用了 `float64` 和 `0.00000001` 作为 epsilon 比较。
    *   **生产环境必须修改**：金融系统中严禁直接使用 `float64` 进行金额计算。
    *   **解决方案**：
        *   方案 A：使用 `github.com/shopspring/decimal` 库。
        *   方案 B：将所有价格乘以 $10^8$ 转为 `int64` (整数) 存储和计算（例如比特币用“聪”，法币用“分”）。

4.  **数据结构选择**:
    *   这里使用了 `container/list`。对于初学者或小规模模拟盘足够。
    *   **大规模优化**：`list` 的插入和查找是 $O(N)$。专业交易所通常使用：
        *   `map[price]*list.List`：快速定位价格档位。
        *   `Red-Black Tree` 或 `Skip List`：维护价格排序。
        *   这样可以将匹配复杂度降低到 $O(\log N)$ 或 $O(1)$。

### 5. 如何集成到你的系统中？

1.  **API 层**：创建一个 HTTP/WebSocket 接口接收用户的下单请求。
2.  **转换**：将 JSON 请求转换为 `Order` 结构体。
3.  **调用引擎**：调用 `orderBook.PlaceOrder(order)`。
4.  **处理结果**：
    *   获取返回的 `[]Trade`。
    *   **异步写入数据库**：将成交记录写入 `trades` 表。
    *   **更新账户**：根据成交结果，扣减买方资金，增加卖方资金/持仓（这一步需要在数据库事务中完成）。
    *   **推送消息**：通过 WebSocket 将成交结果实时推送给前端。

### 6. 下一步扩展建议

*   **市价单支持**：市价单不需要检查价格条件，直接一直匹配直到数量用完或对手盘为空。
*   **订单取消**：实现 `CancelOrder(orderID)` 方法，遍历链表找到订单并移除，同时解冻资金。
*   **持久化快照**：定期将内存中的 OrderBook 状态保存到 Redis 或 DB，防止程序重启数据丢失。
*   **压力测试**：使用 Go 的 `testing` 包或工具（如 k6）模拟每秒 10,000+ 的下单请求，观察锁竞争情况。

这个实现是一个坚实的起点，既展示了 Go 的并发安全性，又实现了核心的金融逻辑。你可以直接复制上面的代码运行查看效果。


# 加上”市价单支持"

在原有的限价单（Limit Order）基础上增加**市价单（Market Order）**支持，核心区别在于：
1.  **价格逻辑**：市价单不指定价格，而是以“当前对手盘的最优价格”立即成交。
2.  **资金/数量校验**：
    *   **买入市价单**：用户指定的是**总金额**（例如：我要买 10000 USDT 的比特币），而不是数量。系统需要计算能买多少币。
    *   **卖出市价单**：用户指定的是**总数量**（例如：我要卖出 1 个比特币），系统直接按市场价卖掉。
3.  **滑点风险**：如果对手盘深度不足，市价单会连续吃掉多个档位的订单，导致平均成交价偏离当前最新价。

以下是修改后的完整代码，增加了市价单逻辑、总金额字段以及更完善的精度处理（引入 `decimal` 库的概念，但为了保持单文件可运行，这里暂时用 `float64` 并加强注释，生产环境请替换为 `shopspring/decimal`）。

### 核心改动点
1.  **Order 结构体**：增加 `QuoteQuantity` 字段（用于买入市价单的总金额）。
2.  **MatchOrder 逻辑**：
    *   识别 `Type == MARKET`。
    *   如果是**买入市价单**：循环匹配直到 `剩余金额` 用完。
    *   如果是**卖出市价单**：循环匹配直到 `剩余数量` 用完。
3.  **防穿仓保护**：如果对手盘被吃光了还没满足需求，市价单会停止（不会挂入订单簿），并返回实际成交结果。

### 完整代码实现 (`main.go`)

```go
package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// --- 常量与类型定义 ---

type Side int
type OrderType int

const (
	Buy Side = iota + 1
	Sell
)

const (
	LIMIT OrderType = iota + 1
	MARKET
)

func (s Side) String() string {
	if s == Buy {
		return "BUY"
	}
	return "SELL"
}

func (t OrderType) String() string {
	if t == LIMIT {
		return "LIMIT"
	}
	return "MARKET"
}

// Order 订单结构
type Order struct {
	ID             string
	UserID         int64
	Symbol         string
	Side           Side
	Type           OrderType
	Price          float64 // 限价单价格
	Quantity       float64 // 基础资产数量 (如 BTC 数量)
	QuoteQuantity  float64 // 计价资产数量 (如 USDT 金额，仅用于 BUY MARKET)
	Timestamp      time.Time
}

// Trade 成交记录
type Trade struct {
	Price         float64
	Quantity      float64
	BuyerOrderID  string
	SellerOrderID string
	BuyerUserID   int64
	SellerUserID  int64
	Timestamp     time.Time
}

// OrderBook 订单簿
type OrderBook struct {
	Symbol string
	bids   *list.List // 买单：高价在前
	asks   *list.List // 卖单：低价在前
	mu     sync.Mutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		bids:   list.New(),
		asks:   list.New(),
	}
}

// PlaceOrder 外部入口
func (ob *OrderBook) PlaceOrder(order *Order) []Trade {
	trades, _ := ob.MatchOrder(order)
	return trades
}

// MatchOrder 核心撮合逻辑 (支持限价和市价)
func (ob *OrderBook) MatchOrder(order *Order) ([]Trade, *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []Trade
	order.Timestamp = time.Now()

	// 确定对手盘列表
	var opponentList *list.List
	if order.Side == Buy {
		opponentList = ob.asks
	} else {
		opponentList = ob.bids
	}

	// --- 处理市价单 (MARKET) ---
	if order.Type == MARKET {
		return ob.matchMarketOrder(order, opponentList, &trades)
	}

	// --- 处理限价单 (LIMIT) ---
	return ob.matchLimitOrder(order, opponentList, &trades)
}

// 撮合限价单逻辑
func (ob *OrderBook) matchLimitOrder(order *Order, opponentList *list.List, trades *[]Trade) ([]Trade, *Order) {
	remainingQty := order.Quantity
	
	// 定义匹配条件
	isMatch := func(o *Order) bool {
		if order.Side == Buy {
			return o.Price <= order.Price // 买：对手卖价 <= 我的限价
		}
		return o.Price >= order.Price // 卖：对手买价 >= 我的限价
	}

	var next *list.Element
	for e := opponentList.Front(); e != nil && remainingQty > 1e-8; e = next {
		next = e.Next()
		opponent := e.Value.(*Order)

		if !isMatch(opponent) {
			break // 价格不满足，停止
		}

		fillQty := min(remainingQty, opponent.Quantity)
		ob.executeTrade(order, opponent, fillQty, trades)

		remainingQty -= fillQty
		opponent.Quantity -= fillQty

		if opponent.Quantity < 1e-8 {
			opponentList.Remove(e)
		}
	}

	// 如果有剩余，挂入订单簿
	if remainingQty > 1e-8 {
		order.Quantity = remainingQty
		ob.insertIntoBook(order)
		fmt.Printf("📝 [LIMIT] 挂单成功：%s (%s) @ %.2f, 剩余 %.4f\n", 
			order.ID, order.Side, order.Price, order.Quantity)
		return *trades, order
	}

	fmt.Printf("🎉 [LIMIT] 订单 %s 全部成交\n", order.ID)
	return *trades, nil
}

// 撮合市价单逻辑
func (ob *OrderBook) matchMarketOrder(order *Order, opponentList *list.List, trades *[]Trade) ([]Trade, *Order) {
	
	// 市价单不会进入订单簿，必须立即成交或取消 (IOC)
	
	if order.Side == Buy {
		// === 买入市价单 ===
		// 用户指定的是 QuoteQuantity (总金额，如 10000 USDT)
		remainingQuoteQty := order.QuoteQuantity
		
		if remainingQuoteQty < 0.01 { // 最小金额限制
			fmt.Printf("❌ [MARKET BUY] 订单 %s 金额过小，拒绝\n", order.ID)
			return *trades, nil
		}

		var next *list.Element
		for e := opponentList.Front(); e != nil && remainingQuoteQty > 0.01; e = next {
			next = e.Next()
			opponent := e.Value.(*Order) // 对手是卖单

			// 计算这一档能买多少币：min(剩余金额/对手价格，对手数量)
			maxBuyQtyByMoney := remainingQuoteQty / opponent.Price
			fillQty := min(maxBuyQtyByMoney, opponent.Quantity)
			
			if fillQty < 1e-8 {
				break
			}

			cost := fillQty * opponent.Price // 实际花费
			
			ob.executeTrade(order, opponent, fillQty, trades)

			remainingQuoteQty -= cost
			opponent.Quantity -= fillQty

			fmt.Printf("💰 [MARKET BUY] 消耗 %.2f USDT, 买入 %.4f BTC @ %.2f\n", 
				cost, fillQty, opponent.Price)

			if opponent.Quantity < 1e-8 {
				opponentList.Remove(e)
			}
		}

		if remainingQuoteQty > 1.0 { // 如果还剩很多钱没花出去，说明深度不足
			fmt.Printf("⚠️ [MARKET BUY] 订单 %s 未完全成交。请求：%.2f, 剩余未成交金额：%.2f (深度不足)\n", 
				order.ID, order.QuoteQuantity, remainingQuoteQty)
			// 注意：市价单剩余部分直接取消，不挂单
		} else if remainingQuoteQty > 0.01 {
			// 剩下一点点零头，通常退回给用户或在最后一笔成交中微调，这里简化处理视为成交结束
			fmt.Printf("✅ [MARKET BUY] 订单 %s 基本成交，剩余零头 %.4f 已忽略/退回\n", order.ID, remainingQuoteQty)
		} else {
			fmt.Printf("🎉 [MARKET BUY] 订单 %s 全部成交\n", order.ID)
		}

	} else {
		// === 卖出市价单 ===
		// 用户指定的是 Quantity (总数量，如 5 BTC)
		remainingQty := order.Quantity

		var next *list.Element
		for e := opponentList.Front(); e != nil && remainingQty > 1e-8; e = next {
			next = e.Next()
			opponent := e.Value.(*Order) // 对手是买单

			fillQty := min(remainingQty, opponent.Quantity)
			
			ob.executeTrade(order, opponent, fillQty, trades)

			remainingQty -= fillQty
			opponent.Quantity -= fillQty

			fmt.Printf("💰 [MARKET SELL] 卖出 %.4f BTC @ %.2f, 获得 %.2f USDT\n", 
				fillQty, opponent.Price, fillQty * opponent.Price)

			if opponent.Quantity < 1e-8 {
				opponentList.Remove(e)
			}
		}

		if remainingQty > 1e-8 {
			fmt.Printf("⚠️ [MARKET SELL] 订单 %s 未完全成交。请求：%.4f, 剩余未成交数量：%.4f (深度不足)\n", 
				order.ID, order.Quantity, remainingQty)
		} else {
			fmt.Printf("🎉 [MARKET SELL] 订单 %s 全部成交\n", order.ID)
		}
	}

	return *trades, nil
}

// 执行具体的成交动作
func (ob *OrderBook) executeTrade(taker, maker *Order, qty float64, trades *[]Trade) {
	var buyerID, sellerID int64
	var buyerOID, sellerOID string

	if taker.Side == Buy {
		buyerID, buyerOID = taker.UserID, taker.ID
		sellerID, sellerOID = maker.UserID, maker.ID
	} else {
		buyerID, buyerOID = maker.UserID, maker.ID
		sellerID, sellerOID = taker.UserID, taker.ID
	}

	trade := Trade{
		Price:         maker.Price, // 以挂单方 (Maker) 的价格成交
		Quantity:      qty,
		BuyerOrderID:  buyerOID,
		SellerOrderID: sellerOID,
		BuyerUserID:   buyerID,
		SellerUserID:  sellerID,
		Timestamp:     time.Now(),
	}
	*trades = append(*trades, trade)
}

// 将限价单插入订单簿
func (ob *OrderBook) insertIntoBook(order *Order) {
	if order.Side == Buy {
		insertSorted(ob.bids, order, func(a, b *Order) bool {
			if a.Price != b.Price {
				return a.Price > b.Price
			}
			return a.Timestamp.Before(b.Timestamp)
		})
	} else {
		insertSorted(ob.asks, order, func(a, b *Order) bool {
			if a.Price != b.Price {
				return a.Price < b.Price
			}
			return a.Timestamp.Before(b.Timestamp)
		})
	}
}

func insertSorted(l *list.List, order *Order, less func(a, b *Order) bool) {
	for e := l.Front(); e != nil; e = e.Next() {
		curr := e.Value.(*Order)
		if less(order, curr) {
			l.InsertBefore(order, e)
			return
		}
	}
	l.PushBack(order)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// --- 测试场景 ---

func main() {
	fmt.Println("🚀 启动支持市价单的撮合引擎...")
	ob := NewOrderBook("BTC/USDT")

	// 1. 构建深度：卖家挂单
	// 卖家 A: 1 BTC @ 50,000
	ob.PlaceOrder(&Order{ID: "S-A", UserID: 101, Side: Sell, Type: LIMIT, Price: 50000, Quantity: 1.0})
	// 卖家 B: 2 BTC @ 50,100
	ob.PlaceOrder(&Order{ID: "S-B", UserID: 102, Side: Sell, Type: LIMIT, Price: 50100, Quantity: 2.0})
	// 卖家 C: 5 BTC @ 50,500
	ob.PlaceOrder(&Order{ID: "S-C", UserID: 103, Side: Sell, Type: LIMIT, Price: 50500, Quantity: 5.0})

	fmt.Println("\n--- 当前卖盘深度 ---")
	fmt.Println("1.0 BTC @ 50,000")
	fmt.Println("2.0 BTC @ 50,100")
	fmt.Println("5.0 BTC @ 50,500")
	fmt.Println("------------------\n")

	// 2. 测试买入市价单 (Market Buy)
	// 用户想花 150,000 USDT 买入 BTC
	// 预期：
	// - 吃掉 S-A: 1.0 * 50000 = 50,000 (剩 100,000)
	// - 吃掉 S-B: 2.0 * 50100 = 100,200 -> 钱不够了！
	// - 只能买 100,000 / 50100 ≈ 1.996 BTC
	fmt.Println("👤 用户发起市价买入：总额 150,000 USDT")
	ob.PlaceOrder(&Order{
		ID:            "M-BUY-001",
		UserID:        201,
		Side:          Buy,
		Type:          MARKET,
		QuoteQuantity: 150000, // 关键：市价买单看的是钱
	})

	fmt.Println("\n------------------\n")

	// 3. 重建深度：买家挂单
	ob.PlaceOrder(&Order{ID: "B-X", UserID: 301, Side: Buy, Type: LIMIT, Price: 49000, Quantity: 3.0})
	ob.PlaceOrder(&Order{ID: "B-Y", UserID: 302, Side: Buy, Type: LIMIT, Price: 48000, Quantity: 2.0})

	fmt.Println("--- 当前买盘深度 ---")
	fmt.Println("3.0 BTC @ 49,000")
	fmt.Println("2.0 BTC @ 48,000")
	fmt.Println("------------------\n")

	// 4. 测试卖出市价单 (Market Sell)
	// 用户想卖出 4.0 BTC
	// 预期：
	// - 吃掉 B-X: 3.0 @ 49000
	// - 吃掉 B-Y: 1.0 @ 48000 (剩余 1.0 无法成交，因为买盘没了)
	fmt.Println("👤 用户发起市价卖出：数量 4.0 BTC")
	ob.PlaceOrder(&Order{
		ID:       "M-SELL-001",
		UserID:   401,
		Side:     Sell,
		Type:     MARKET,
		Quantity: 4.0, // 关键：市价卖单看的是币的数量
	})
}
```

### 关键逻辑解析

#### 1. 买入市价单 (`Side == Buy`, `Type == MARKET`)
*   **输入**：`QuoteQuantity` (例如 150,000 USDT)。
*   **逻辑**：
    *   遍历卖单列表（从最低价开始）。
    *   计算当前档位能买多少：`canBuy = remainingMoney / opponentPrice`。
    *   实际成交数量 `fillQty = min(canBuy, opponent.Quantity)`。
    *   扣除花费：`remainingMoney -= fillQty * opponentPrice`。
    *   **终止条件**：钱花完了，或者卖单被吃光了。
*   **特点**：如果卖单深度不足，用户可能买不到预期的数量，剩下的钱会退回（模拟系统中表现为订单结束，未成交部分不挂单）。

#### 2. 卖出市价单 (`Side == Sell`, `Type == MARKET`)
*   **输入**：`Quantity` (例如 4.0 BTC)。
*   **逻辑**：
    *   遍历买单列表（从最高价开始）。
    *   直接匹配数量：`fillQty = min(remainingQty, opponent.Quantity)`。
    *   **终止条件**：货卖完了，或者买单被吃光了。
*   **特点**：如果买盘深度不足，用户可能只能卖掉一部分，剩下的币保留在账户中（订单未完全成交部分自动取消）。

#### 3. 滑点演示
在上面的测试代码中：
*   用户想买 150,000 USDT。
*   第一档 50,000 很便宜，全吃了。
*   第二档 50,100 稍贵，把钱花光了。
*   **结果**：用户的平均成交成本会高于 50,000，这就是**滑点**。系统日志清晰地展示了这个过程。

### 生产环境注意事项

1.  **精度问题**：
    *   代码中的 `float64` 和 `1e-8` 只是演示。
    *   **必须**使用 `github.com/shopspring/decimal`。
    *   对于市价单计算 `fillQty = money / price`，需要设定明确的舍入规则（通常是向下取整到最小交易单位，避免产生小数点后无数位的碎股）。

2.  **资金预冻结**：
    *   在市价单进入撮合引擎**之前**，账户系统必须先冻结用户的资金（买入时冻结 USDT，卖出时冻结 BTC）。
    *   撮合结束后，根据实际成交金额进行多退少补（将未成交的资金解冻退回可用余额）。

3.  **最大滑点保护**：
    *   真实交易中，用户可能会担心市价单在极端行情下把价格打到地板/天花板。
    *   通常会增加一个参数 `SlippageLimit` (例如 2%)。如果撮合过程中，加权平均价格偏离当前最新价超过 2%，则停止撮合并取消剩余订单。

这个版本已经具备了模拟盘所需的核心市价单功能，能够很好地展示“以市场最优价立即成交”的行为特征。

# 加上“订单取消”功能

