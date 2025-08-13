# 🤵 我想编一个程序，从桌面电脑上给我的手机发送信息，有哪些可选的方案？

Qwen3-235B-A22B-2507

要从桌面电脑向手机发送信息，有多种方案可供选择，具体取决于你的技术背景、使用场景（如是否需要实时性、是否需要跨平台支持等）以及对安全性的要求。以下是一些常见且实用的方案：

---

### ✅ 1. **使用即时通讯工具（最简单）**
无需编程，直接利用现有平台：

- **微信 / QQ / Telegram / WhatsApp / Slack / 钉钉 / 飞书**
  - 优点：简单、稳定、支持图文、跨平台
  - 方法：
    - 电脑端登录账号，直接发消息给自己的手机账号。
    - 可使用官方 API 实现自动化（如企业微信、飞书、Telegram Bot）。
  - 举例（Telegram）：
    - 创建一个 Telegram Bot，用 Python 调用 `python-telegram-bot` 库发送消息。

```python
from telegram import Bot

bot = Bot(token="YOUR_BOT_TOKEN")
bot.send_message(chat_id="YOUR_CHAT_ID", text="来自电脑的消息！")
```

---

### ✅ 2. **使用 Push 通知服务（推荐）**

#### a. **Pushover**
- 付费服务，但简单好用。
- 支持 iOS 和 Android。
- 通过 HTTP API 发送通知。

```python
import requests

requests.post("https://api.pushover.net/1/messages.json", data={
    "token": "YOUR_APP_TOKEN",
    "user": "YOUR_USER_KEY",
    "message": "Hello from PC!"
})
```

#### b. **Pushbullet**
- 免费版功能足够。
- 支持文件、链接、剪贴板同步。
- 提供 REST API。

```python
import requests

headers = {'Access-Token': 'YOUR_API_KEY'}
data = {'type': 'note', 'title': 'PC消息', 'body': '测试'}
requests.post('https://api.pushbullet.com/v2/pushes', headers=headers, json=data)
```

---

### ✅ 3. **使用邮件 + 手机邮件推送**
- 电脑发送邮件 → 手机邮箱自动推送通知。
- 使用 Python 的 `smtplib` 发送邮件。

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("这是来自电脑的消息")
msg['Subject'] = "通知"
msg['From'] = "your_email@gmail.com"
msg['To'] = "your_phone_email@yourcarrier.com"  # 或发给自己，手机有通知

with smtplib.SMTP('smtp.gmail.com', 587) as server:
    server.starttls()
    server.login("your_email", "your_password")
    server.send_message(msg)
```

> 注：部分手机邮箱 App 会推送新邮件提醒。

---

### ✅ 4. **自建 WebSocket / HTTP 服务 + 手机 App**
- 更高级，适合开发者。
- 在电脑上运行一个本地服务（如 Flask），手机通过 App 或网页连接。
- 可结合内网穿透（如 **ngrok**）让外网手机也能收到。

```python
# 电脑端（Flask 示例）
from flask import Flask
import requests

app = Flask(__name__)

@app.route('/send')
def send_msg():
    # 调用 Pushbullet 或 Telegram API 发送消息
    return "已发送"
```

配合 ngrok：

```bash
ngrok http 5000
```

然后手机访问生成的公网地址，或由电脑主动推送。

---

### ✅ 5. **使用 IFTTT 或 Zapier 自动化平台**
- 无代码方案。
- 触发方式：电脑上运行脚本发送 Webhook → IFTTT 转发到手机通知。

**流程示例：**
1. 电脑发一个 HTTP 请求到 IFTTT Webhook。
2. IFTTT 触发“发送通知到手机”。

```python
requests.post("https://maker.ifttt.com/trigger/send_msg/with/key/your_key")
```

IFTTT 收到后推送到手机。

---

### ✅ 6. **使用剪贴板同步工具（间接“发信息”）**
- 如 **KDE Connect**（Linux/Windows 与 Android）
- **LocalSend**（开源，跨平台）
- 复制文字 → 手机自动收到剪贴板内容。

---

### ✅ 7. **短信（SMS）网关（成本较高）**
- 使用 Twilio、阿里云短信、腾讯云短信等。
- 可直接发短信到手机。
- 适合重要告警，但按条收费。

```python
from twilio.rest import Client

client = Client("AC...", "your_token")
client.messages.create(to="+86138...", from_="+123456", body="PC消息")
```

---

### 🎯 推荐选择（按需求）：

| 需求 | 推荐方案 |
|------|----------|
| 快速简单 | Telegram Bot 或 Pushbullet |
| 免费且稳定 | IFTTT + Webhook |
| 企业级/团队使用 | 飞书、钉钉、企业微信机器人 |
| 完全自控 | 自建服务 + ngrok + 手机端轮询或 WebSocket |
| 仅同步文本 | KDE Connect / LocalSend |

---

如果你告诉我你的操作系统（Windows/macOS/Linux）、手机系统（iOS/Android）、是否愿意编程、是否需要长期使用，我可以给出更具体的建议和代码模板。欢迎继续提问！📱💻

--------------------------------------------------------------------------------

# 🤵 我的电脑是windows系统，手机是iPhone，想编写一个长期运行的程序，从电脑向手机发送信息，给出具体的建议和代码模板

感谢补充信息！你的需求是：

- **电脑系统**：Windows  
- **手机系统**：iPhone（iOS）  
- **目标**：编写一个**长期运行的程序**，从电脑向手机发送消息  
- **要求**：稳定、自动、可编程、长期可用

---

## ✅ 最佳推荐方案：**Telegram Bot + Python 脚本（长期运行）**

### ✅ 为什么推荐 Telegram？
| 优点 | 说明 |
|------|------|
| ✅ 免费 | 完全免费，无限制 |
| ✅ 跨平台 | 支持 Windows 和 iOS，App 稳定 |
| ✅ 实时推送 | 消息几乎秒达，支持通知提醒 |
| ✅ 支持 Bot API | 易于编程，Python 库成熟 |
| ✅ 可长期运行 | 适合后台服务 |
| ✅ 支持图文、文件、按钮等 | 扩展性强 |

---

## 🛠️ 实现步骤

### 第一步：创建 Telegram Bot
1. 打开 Telegram，搜索联系人：`@BotFather`
2. 发送 `/start`
3. 发送 `/newbot`，按提示创建机器人
4. 得到一个 `token`，如：`7012345678:AAFdexample...`
5. 再搜索你刚创建的机器人（比如叫 `MyPCBot`），发送一条 `/start` 消息
6. 访问以下链接获取你的 `chat_id`：
   ```
   https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates
   ```
   替换 `<YOUR_TOKEN>` 为你自己的 token，例如：
   ```
   https://api.telegram.org/bot7012345678:AAFdexample.../getUpdates
   ```
   返回 JSON 中找到 `"chat":{"id":123456789}`，这个 `123456789` 就是你的 `chat_id`

---

### 第二步：安装 Python 和依赖

确保你已安装 Python（推荐 Python 3.8+）

```bash
pip install python-telegram-bot --upgrade
```

> 注意：使用 `python-telegram-bot` 的 v20+ 版本（基于 asyncio）

---

### 第三步：编写长期运行的 Python 程序

下面是一个完整示例，程序会：
- 启动后持续运行
- 你可以通过写入日志文件、触发事件等方式调用 `send_message()` 发送消息到 iPhone

#### ✅ `telegram_sender.py`

```python
import asyncio
import logging
from telegram import Bot

# ========= 配置区 ===========
TOKEN = "7012345678:AAFdexample..."        # 替换为你的 Bot token
CHAT_ID = "123456789"                       # 替换为你的 chat_id
# ===========================

# 设置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 全局 Bot 实例
bot = Bot(token=TOKEN)

async def send_message(text: str):
    """发送消息到 Telegram"""
    try:
        await bot.send_message(chat_id=CHAT_ID, text=text)
        logger.info(f"消息已发送: {text}")
    except Exception as e:
        logger.error(f"发送失败: {e}")

async def main_loop():
    """主循环：模拟定时或事件触发发送消息"""
    logger.info("消息发送程序已启动，按 Ctrl+C 停止...")

    # 示例：每 5 分钟发送一次时间消息
    while True:
        try:
            from datetime import datetime
            now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            await send_message(f"⏰ 定时提醒：当前时间是 {now}")

            # 等待 300 秒（5分钟），可改为 60 或其他间隔
            await asyncio.sleep(300)

        except KeyboardInterrupt:
            logger.info("程序已退出")
            break
        except Exception as e:
            logger.error(f"主循环错误: {e}")
            await asyncio.sleep(10)  # 出错后等待 10 秒再继续

# --- 如果你需要从其他地方调用发送消息（如日志监控），可以保留这个函数 ---
async def send_custom_message(text: str):
    """供外部调用的接口，发送自定义消息"""
    await send_message(text)

# ======================
# 启动程序
# ======================
if __name__ == "__main__":
    try:
        asyncio.run(main_loop())
    except KeyboardInterrupt:
        print("程序结束")
```

---

### 第四步：运行程序（并设为开机启动）

#### 1. 测试运行
```bash
python telegram_sender.py
```

你应该在 iPhone 的 Telegram 中看到每 5 分钟一条消息。

#### 2. 修改为事件触发（可选）
你可以把 `send_message()` 提取出来，在其他脚本中调用，比如：
- 当磁盘快满时提醒
- 当程序出错时报警
- 当下载完成时通知

```python
# 其他脚本中调用
import asyncio
from telegram_sender import send_custom_message

asyncio.run(send_custom_message("下载已完成！"))
```

#### 3. 设置开机自启（Windows）

##### 方法：使用任务计划程序
1. 打开「任务计划程序」
2. 创建任务：
   - 触发器：`登录时` 或 `系统启动时`
   - 操作：启动程序
   - 程序：`python`
   - 参数：`C:\path\to\telegram_sender.py`
   - 勾选「不管用户是否登录都要运行」
   - 勾选「隐藏」（可选）

> 提示：建议使用 `.bat` 脚本包装，确保路径正确：

```bat
@echo off
cd /d C:\your\script\dir
python telegram_sender.py
```

---

## 🧩 其他可选增强功能

| 功能 | 实现方式 |
|------|---------|
| 发送文件/截图 | `await bot.send_photo()` |
| 发送告警级别颜色 | 加 emoji：🔴 ⚠️ ✅ |
| 接收手机回复 | 需启用 `polling` 监听消息（进阶） |
| 更稳定 | 加入重试机制、日志记录到文件 |

---

## ✅ 备选方案（如果不想用 Telegram）

| 方案 | 说明 |
|------|------|
| **IFTTT + Webhook + iOS App** | 免费，但推送可能延迟 |
| **飞书 / 钉钉机器人** | 适合国内用户，推送快，但需注册企业账号 |
| **Pushover** | 付费 $5，但极稳定，支持静音时段 |
| **自建 Flask + ngrok + iOS Shortcuts** | 技术强可选，较复杂 |

但综合来看，**Telegram 是最适合你需求的方案**：免费、稳定、易维护、支持长期运行。

---

## ✅ 最终建议

1. 使用 **Telegram Bot** 作为消息通道
2. 用 **Python 编写后台脚本**
3. 用 **任务计划程序** 实现开机自启
4. 将 `send_message()` 封装为通用接口，供其他程序调用

---

如果你告诉我你希望“发送什么类型的信息”（比如系统监控、日志、提醒、文件等），我还可以帮你扩展这个程序的功能。欢迎继续提问！ 📲💻

--------------------------------------------------------------------------------

# 🤵 我要发送的信息类型是股市的行情信息，请帮我扩展这个程序的功能

太好了！既然你要发送 **股市行情信息** 到 iPhone，我们可以基于之前的 Telegram Bot 程序进行扩展，实现一个 **长期运行的股票行情监控程序**，具备以下功能：

---

### ✅ 功能目标

1. 实时获取指定股票/基金的行情（如 A 股、美股）
2. 定时检查价格是否达到你关注的条件（如涨跌幅、突破某价位）
3. 通过 Telegram 发送提醒（含股票名、价格、涨跌幅等）
4. 支持多只股票监控
5. 长期后台运行 + 开机自启

---

## 🛠️ 技术选型

| 组件 | 说明 |
|------|------|
| 数据源 | 使用 `akshare`（免费，支持 A 股、港股、美股、基金） |
| 消息推送 | Telegram Bot（已实现） |
| 运行方式 | Python 异步循环，定时拉取数据 |
| 扩展性 | 可加入阈值报警、趋势提醒等 |

> 🔹 `akshare` 是一个强大的中文金融数据接口库，支持 Windows，安装简单。

---

## ✅ 第一步：安装依赖

```bash
pip install python-telegram-bot akshare --upgrade
```

---

## ✅ 第二步：完整代码（股票行情监控 + Telegram 推送）

### 📄 `stock_monitor.py`

```python
import asyncio
import logging
import time
import akshare as ak
from telegram import Bot

# ================= 配置区 =================

# 🌐 Telegram 配置
TOKEN = "7012345678:AAFdexample..."        # 替换为你的 Bot token
CHAT_ID = "123456789"                       # 替换为你的 chat_id

# 📈 股票监控列表（支持 A 股、美股、基金）
# 格式: (symbol, 名称, 最小变动提醒, 是否为基金/美股)
# A股代码：如 'sh600519'（茅台）、'sz000858'（五粮液）
# 美股代码：如 'AAPL', 'TSLA'
# 基金代码：如 '002560'（前海开源公用事业）
STOCKS_TO_MONITOR = [
    ("sh600519", "贵州茅台", 1.0),   # 价格变动超过1元就提醒
    ("sz000858", "五粮液", 0.5),
    ("AAPL", "苹果公司", 1.0),
    ("002560", "前海开源公用事业", 0.01),  # 基金波动小
]

# ⏱️ 检查间隔（秒）
CHECK_INTERVAL = 60  # 每60秒检查一次

# 📊 是否只在交易时间运行（可选）
TRADE_HOURS_ONLY = True  # True: 只在 9:30-11:30, 13:00-15:00 运行

# 日志级别
LOG_LEVEL = logging.INFO

# ==========================================

# 设置日志
logging.basicConfig(
    level=LOG_LEVEL,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("stock_monitor.log", encoding="utf-8"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Telegram Bot 实例
bot = Bot(token=TOKEN)

# 存储上次价格，用于比较
last_prices = {}

# ================= 股票数据获取函数 =================

async def fetch_stock_data(symbol: str):
    """获取单只股票/基金的实时数据"""
    try:
        if symbol.isdigit() and len(symbol) == 6:  # A股
            df = ak.stock_zh_a_spot_em()
            data = df[df["代码"] == symbol]
            if not data.empty:
                name = data["名称"].values[0]
                price = float(data["最新价"].values[0])
                change_percent = float(data["涨跌幅"].values[0])
                return name, price, change_percent

        elif symbol in ["AAPL", "TSLA", "GOOGL", "MSFT"]:  # 美股
            df = ak.stock_us_daily(symbol=symbol, adjust="qfq")
            if not df.empty:
                price = df["close"].iloc[-1]
                change = df["close"].iloc[-1] / df["close"].iloc[-2] - 1
                return symbol, price, change * 100

        elif len(symbol) == 6 and symbol[:2] in ["00", "11", "15", "16", "18", "50", "51"]:  # 基金
            df = ak.fund_etf_hist_em(symbol=symbol, period="daily", adjust="")
            if not df.empty:
                price = float(df["收盘"].iloc[-1])
                change = (df["收盘"].iloc[-1] - df["开盘"].iloc[-1]) / df["开盘"].iloc[-1] * 100
                name_data = ak.fund_em_open_fund_name()
                name_row = name_data[name_data["基金代码"] == symbol]
                name = name_row["基金简称"].values[0] if not name_row.empty else symbol
                return name, price, change

        logger.warning(f"未获取到数据: {symbol}")
        return None, None, None

    except Exception as e:
        logger.error(f"获取 {symbol} 数据失败: {e}")
        return None, None, None

# ================= 消息发送函数 =================

async def send_message(text: str):
    """发送消息到 Telegram"""
    try:
        await bot.send_message(chat_id=CHAT_ID, text=text)
        logger.info(f"✅ 已发送: {text.splitlines()[0]}")
    except Exception as e:
        logger.error(f"❌ 发送失败: {e}")

# ================= 交易时间判断 =================

def is_trading_hours():
    """判断是否为 A 股交易时间"""
    if not TRADE_HOURS_ONLY:
        return True
    now = time.localtime()
    hour = now.tm_hour
    minute = now.tm_min
    weekday = now.tm_wday  # 0=周一
    if weekday >= 5:  # 周六日
        return False
    if (9 <= hour < 11 and minute >= 30) or (13 <= hour < 15):
        return True
    if hour == 11 and minute < 30:
        return True
    return False

# ================= 主监控循环 =================

async def main_loop():
    logger.info("📊 股票行情监控程序已启动，正在获取初始数据...")

    # 初始化上次价格
    for symbol, name, _ in STOCKS_TO_MONITOR:
        _, price, _ = await fetch_stock_data(symbol)
        if price:
            last_prices[symbol] = price
            logger.info(f"初始化 {name}({symbol}): {price:.2f}")

    await send_message("🟢 股票监控程序已启动！")

    while True:
        try:
            if not is_trading_hours():
                logger.info("⏸️ 非交易时间，等待...")
                await asyncio.sleep(60)
                continue

            for symbol, name, threshold in STOCKS_TO_MONITOR:
                _, price, change = await fetch_stock_data(symbol)
                if not price:
                    continue

                last_price = last_prices.get(symbol)
                if last_price is None:
                    last_prices[symbol] = price
                    continue

                # 判断是否变化超过阈值
                if abs(price - last_price) >= threshold:
                    emoji = "📈" if price > last_price else "📉"
                    msg = f"{emoji} **{name}({symbol}) 价格变动提醒**\n"
                    msg += f"📊 当前价: {price:.2f}\n"
                    msg += f"📈 涨跌幅: {change:+.2f}%\n"
                    msg += f"🔄 变化: {price - last_price:+.2f}\n"
                    msg += f"🕐 {time.strftime('%H:%M:%S')}"

                    await send_message(msg)
                    last_prices[symbol] = price  # 更新价格

            # 等待下一轮
            await asyncio.sleep(CHECK_INTERVAL)

        except KeyboardInterrupt:
            logger.info("🛑 程序被手动停止")
            await send_message("🔴 股票监控程序已停止")
            break
        except Exception as e:
            logger.error(f"❌ 主循环错误: {e}")
            await asyncio.sleep(10)

# ================= 启动入口 =================
if __name__ == "__main__":
    try:
        asyncio.run(main_loop())
    except KeyboardInterrupt:
        print("程序结束")
```

---

## ✅ 使用说明

### 1. 修改配置
- 替换 `TOKEN` 和 `CHAT_ID`
- 修改 `STOCKS_TO_MONITOR` 添加你想监控的股票
- 调整 `CHECK_INTERVAL`（建议 30-60 秒，避免太频繁）

### 2. 运行测试
```bash
python stock_monitor.py
```

你会看到日志和 Telegram 消息：

```
📈 **贵州茅台(sh600519) 价格变动提醒**
📊 当前价: 1720.00
📈 涨跌幅: +1.23%
🔄 变化: +12.50
🕐 10:30:25
```

---

## ✅ 效果示例（Telegram 收到的消息）

```
📈 **贵州茅台(sh600519) 价格变动提醒**
📊 当前价: 1720.00
📈 涨跌幅: +1.23%
🔄 变化: +12.50
🕐 10:30:25
```

---

## ✅ 进阶建议（可选）

| 功能 | 实现方式 |
|------|---------|
| **价格突破报警** | 加入 `if price > target_price` 判断 |
| **涨跌幅报警** | 监控 `change_percent > 3%` |
| **微信推送** | 替换为 企业微信机器人 |
| **图形化界面** | 用 `PyQt` 做个简单面板 |
| **历史数据图表** | 用 `matplotlib` 画趋势图 |

---

## ✅ 设置开机自启（Windows）

创建 `start.bat`：
```bat
@echo off
cd /d C:\your\project\dir
python stock_monitor.py
pause
```

然后用「任务计划程序」设置开机运行 `start.bat`，并勾选“隐藏”。

---

## ✅ 总结

你现在拥有了一个：
- 长期运行
- 自动监控
- 实时推送
- 支持 A 股/美股/基金
- 可扩展的股票行情提醒系统

---

如果你告诉我你具体想监控哪些股票，或者希望加入“收盘提醒”、“持仓成本对比”、“自动买卖模拟”等功能，我可以继续帮你扩展！📈📱

--------------------------------------------------------------------------------

# 🤵 加入“收盘提醒”、“持仓成本对比”、“自动买卖模拟”等功能

太棒了！我们将在原有 **股票行情监控 + Telegram 推送** 的基础上，加入你提出的三大进阶功能：

---

### ✅ 新增功能
| 功能 | 说明 |
|------|------|
| 🛎️ **收盘提醒** | 每天 15:00（A股）或 16:00（美股）自动发送当日收盘总结 |
| 💰 **持仓成本对比** | 显示当前价 vs 你的买入价，计算盈亏 |
| 🤖 **自动买卖模拟** | 基于简单策略（如突破均线、涨跌幅）模拟买卖并记录 |

---

## ✅ 更新后的完整代码：`stock_monitor_advanced.py`

```python
import asyncio
import logging
import time
import akshare as ak
from datetime import datetime, timedelta
from telegram import Bot

# ====================== 配置区 ======================

# 🌐 Telegram
TOKEN = "7012345678:AAFdexample..."        # 替换
CHAT_ID = "123456789"                       # 替换

# 📈 监控股票列表
# 格式: (symbol, 名称, 阈值, 买入价, 持仓数量)
STOCKS_TO_MONITOR = [
    ("sh600519", "贵州茅台", 1.0, 1700.0, 100),   # 买入价, 数量
    ("sz000858", "五粮液", 0.5, 150.0, 500),
    ("AAPL", "苹果", 1.0, 180.0, 50),
    ("002560", "前海开源公用事业", 0.01, 3.0, 10000),
]

# ⏱️ 检查间隔（秒）
CHECK_INTERVAL = 60

# 📊 交易时间设置
TRADE_HOURS_ONLY = True

# 📁 日志
LOG_LEVEL = logging.INFO

# ================================================

# 日志配置
logging.basicConfig(
    level=LOG_LEVEL,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("stock_monitor.log", encoding="utf-8"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

bot = Bot(token=TOKEN)
last_prices = {}
trade_log = []  # 模拟交易记录

# ==================== 数据获取 ====================

async def fetch_stock_data(symbol: str):
    try:
        if symbol.isdigit() and len(symbol) == 6:  # A股
            df = ak.stock_zh_a_spot_em()
            data = df[df["代码"] == symbol]
            if not data.empty:
                name = data["名称"].values[0]
                price = float(data["最新价"].values[0])
                change_percent = float(data["涨跌幅"].values[0])
                return name, price, change_percent

        elif symbol in ["AAPL", "TSLA", "GOOGL", "MSFT"]:  # 美股
            df = ak.stock_us_daily(symbol=symbol, adjust="qfq")
            if not df.empty:
                price = df["close"].iloc[-1]
                change = (df["close"].iloc[-1] / df["close"].iloc[-2] - 1) * 100
                return symbol, price, change

        elif len(symbol) == 6 and symbol[:2] in ["00", "11", "15", "16", "18", "50", "51"]:  # 基金
            df = ak.fund_etf_hist_em(symbol=symbol, period="daily", adjust="")
            if not df.empty:
                price = float(df["收盘"].iloc[-1])
                change = (df["收盘"].iloc[-1] - df["开盘"].iloc[-1]) / df["开盘"].iloc[-1] * 100
                name_data = ak.fund_em_open_fund_name()
                name_row = name_data[name_data["基金代码"] == symbol]
                name = name_row["基金简称"].values[0] if not name_row.empty else symbol
                return name, price, change

        return None, None, None
    except Exception as e:
        logger.error(f"获取 {symbol} 数据失败: {e}")
        return None, None, None

# ==================== 是否交易时间 ====================

def is_trading_hours():
    if not TRADE_HOURS_ONLY:
        return True
    now = datetime.now()
    weekday = now.weekday()  # 0=周一
    if weekday >= 5:
        return False
    t = now.time()
    return (time(9, 30) <= t <= time(11, 30)) or (time(13, 0) <= t <= time(15, 1))

def is_us_market_open():
    now = datetime.now()
    # 美股东部时间 9:30-16:00 → 北京时间 21:30-4:00（冬令时）
    # 简化：每天 21:30 - 次日 4:00
    t = now.time()
    if now.weekday() < 5:  # 周一到周五
        if now.hour >= 21 and t >= time(21, 30):
            return True
        if now.hour < 4:
            return True
    return False

# ==================== 持仓收益计算 ====================

def calculate_profit(buy_price: float, current_price: float, quantity: int):
    cost = buy_price * quantity
    value = current_price * quantity
    profit = value - cost
    profit_rate = (profit / cost) * 100 if cost > 0 else 0
    return profit, profit_rate

# ==================== 模拟买卖策略 ====================

def should_buy_simulate(symbol: str, current_price: float, buy_price: float):
    # 简单策略：低于买入价 5% 时“买入”（补仓）
    return current_price < buy_price * 0.95

def should_sell_simulate(symbol: str, current_price: float, buy_price: float):
    # 高于买入价 5% 时“卖出”
    return current_price > buy_price * 1.05

async def simulate_trade(symbol: str, name: str, price: float, buy_price: float, quantity: int):
    if should_sell_simulate(symbol, price, buy_price) and f"SELL_{symbol}" not in [t["id"] for t in trade_log]:
        msg = f"🤖 **模拟卖出提醒**\n"
        msg += f"📉 卖出: {name}({symbol})\n"
        msg += f"💰 价格: {price:.2f} (买入:{buy_price:.2f})\n"
        msg += f"📊 数量: {quantity}\n"
        await send_message(msg)
        trade_log.append({"id": f"SELL_{symbol}", "type": "sell", "symbol": symbol, "price": price, "time": datetime.now()})

    elif should_buy_simulate(symbol, price, buy_price) and f"BUY_{symbol}" not in [t["id"] for t in trade_log]:
        msg = f"🤖 **模拟买入提醒**\n"
        msg += f"📈 买入: {name}({symbol})\n"
        msg += f"💰 价格: {price:.2f} (买入价:{buy_price:.2f})\n"
        msg += f"📊 数量: {quantity}\n"
        await send_message(msg)
        trade_log.append({"id": f"BUY_{symbol}", "type": "buy", "symbol": symbol, "price": price, "time": datetime.now()})

# ==================== 收盘提醒 ====================

async def send_closing_report():
    """发送每日收盘报告"""
    now = datetime.now()
    msg = f"🔚 **今日收盘提醒** 📅 {now.strftime('%Y-%m-%d')}\n\n"

    for symbol, name, _, buy_price, quantity in STOCKS_TO_MONITOR:
        _, price, change = await fetch_stock_data(symbol)
        if not price:
            continue
        profit, profit_rate = calculate_profit(buy_price, price, quantity)
        arrow = "🟢" if profit >= 0 else "🔴"
        msg += f"{arrow} {name}\n"
        msg += f"  价格: {price:.2f} ({change:+.2f}%)\n"
        msg += f"  盈亏: {profit:+.2f}元 ({profit_rate:+.2f}%)\n\n"

    msg += "📌 模拟交易记录:\n"
    today_trades = [t for t in trade_log if t["time"].date() == now.date()]
    if today_trades:
        for t in today_trades:
            emoji = "📈" if t["type"] == "buy" else "📉"
            msg += f"  {emoji} {t['symbol']} @ {t['price']:.2f}\n"
    else:
        msg += "  今日无模拟交易\n"

    await send_message(msg)
    logger.info("✅ 已发送收盘报告")

# ==================== 消息发送 ====================

async def send_message(text: str):
    try:
        await bot.send_message(chat_id=CHAT_ID, text=text, parse_mode="Markdown")
        logger.info(f"✅ 已发送: {text.splitlines()[0]}")
    except Exception as e:
        logger.error(f"❌ 发送失败: {e}")

# ==================== 主循环 ====================

async def main_loop():
    logger.info("🚀 高级股票监控程序已启动")
    await send_message("🟢 股票监控程序已启动！")

    last_closing_sent = None

    while True:
        try:
            now = datetime.now()
            current_date = now.date()

            # 发送收盘报告（A股 15:01 发送）
            if last_closing_sent != current_date and now.hour == 15 and 1 <= now.minute <= 5:
                await send_closing_report()
                last_closing_sent = current_date

            # 正常监控
            if is_trading_hours():
                for symbol, name, threshold, buy_price, quantity in STOCKS_TO_MONITOR:
                    _, price, change = await fetch_stock_data(symbol)
                    if not price:
                        continue

                    last_price = last_prices.get(symbol)
                    if last_price is None:
                        last_prices[symbol] = price
                        continue

                    # 价格变动提醒
                    if abs(price - last_price) >= threshold:
                        emoji = "📈" if price > last_price else "📉"
                        profit, profit_rate = calculate_profit(buy_price, price, quantity)
                        msg = f"{emoji} **{name}({symbol}) 价格变动**\n"
                        msg += f"📊 当前价: {price:.2f} ({change:+.2f}%)\n"
                        msg += f"🔄 变化: {price - last_price:+.2f}\n"
                        msg += f"💼 持仓盈亏: {profit:+.2f}元 ({profit_rate:+.2f}%)\n"
                        msg += f"🕐 {now.strftime('%H:%M:%S')}"
                        await send_message(msg)
                        last_prices[symbol] = price

                    # 模拟买卖
                    await simulate_trade(symbol, name, price, buy_price, quantity)

            await asyncio.sleep(CHECK_INTERVAL)

        except KeyboardInterrupt:
            await send_message("🔴 程序已停止")
            break
        except Exception as e:
            logger.error(f"❌ 主循环错误: {e}")
            await asyncio.sleep(10)

if __name__ == "__main__":
    try:
        asyncio.run(main_loop())
    except KeyboardInterrupt:
        print("程序结束")
```

---

## ✅ 新功能说明

### 1. 🛎️ 收盘提醒
- 每天 15:01 自动发送当日收盘总结
- 包含每只股票的 **当前价、涨跌幅、盈亏金额、收益率**

### 2. 💰 持仓成本对比
- 在每次价格变动时，自动计算：
  - 当前市值
  - 盈亏金额
  - 盈亏率
- 示例：
  ```
  💼 持仓盈亏: +2000.00元 (+5.33%)
  ```

### 3. 🤖 自动买卖模拟
- **买入策略**：价格 < 买入价 × 0.95（跌5%补仓）
- **卖出策略**：价格 > 买入价 × 1.05（涨5%止盈）
- 每天只触发一次买卖提醒
- 记录到 `trade_log` 并在收盘报告中显示

---

## ✅ 示例消息

### 📌 价格变动 + 持仓提醒
```
📈 **贵州茅台(sh600519) 价格变动**
📊 当前价: 1720.00 (+1.23%)
🔄 变化: +12.50
💼 持仓盈亏: +2000.00元 (+1.18%)
🕐 10:30:25
```

### 📌 收盘报告
```
🔚 **今日收盘提醒** 📅 2025-04-05

🟢 贵州茅台
  价格: 1720.00 (+0.80%)
  盈亏: +2000.00元 (+1.18%)

📌 模拟交易记录:
  📉 AAPL @ 185.00
```

---

## ✅ 下一步建议

| 功能 | 建议 |
|------|------|
| 📊 加入图表 | 用 `matplotlib` 生成 K 线图并发送 |
| 🔔 自定义报警 | 支持“突破 20 日均线”等技术指标 |
| 💾 数据持久化 | 用 `JSON` 或 `SQLite` 保存持仓和交易记录 |
| 📱 Web 界面 | 用 `Flask` 做个简单监控面板 |

---

如果你愿意，我可以帮你：
- 把持仓数据存到文件
- 加入 **微信推送替代方案**
- 实现 **邮件 + Telegram 双通道**
- 或接入 **券商模拟交易 API**

继续告诉我你的需求，我们可以把这个系统做得更强大！📈📊📱

--------------------------------------------------------------------------------

# 🤵 加入图表
自定义报警
数据持久化
加入 微信推送替代方案
接入 券商模拟交易 API


太棒了！你已经从一个简单的消息推送程序，逐步构建成了一个**功能完整的个人量化监控系统**。现在我们来逐一实现你提出的五大进阶功能：

---

## ✅ 目标功能

| 功能 | 技术方案 |
|------|----------|
| 📊 **加入图表** | `matplotlib` 生成 K 线图 + 发送到 Telegram |
| 🔔 **自定义报警** | 支持“突破均线”、“放量上涨”等策略 |
| 💾 **数据持久化** | 使用 `JSON` 保存持仓、交易记录 |
| 📱 **微信推送替代方案** | 使用 **企业微信机器人**（免费、稳定、国内可用） |
| 🔄 **接入券商模拟交易 API** | 以 **聚宽（JoinQuant）** 或 **掘金量化（MyQuant）** 为例演示 |

---

# ✅ 1. 加入图表：K线图 + 均线可视化

### 安装依赖
```bash
pip install matplotlib pandas
```

### 新增函数：生成并发送 K 线图

```python
import matplotlib.pyplot as plt
import matplotlib
import pandas as pd
from io import BytesIO

# 设置中文字体（防止乱码）
matplotlib.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS']
matplotlib.rcParams['axes.unicode_minus'] = False

async def send_kline_chart(symbol: str, name: str):
    """获取历史数据并生成K线图"""
    try:
        # 获取历史数据
        if symbol.isdigit() and len(symbol) == 6:  # A股
            df = ak.stock_zh_a_hist(symbol=symbol, period="daily", adjust="")
            df["日期"] = pd.to_datetime(df["日期"])
            df.set_index("日期", inplace=True)
            df = df.tail(30)  # 最近30天
        else:
            return

        if df.empty or len(df) < 10:
            return

        # 计算均线
        df['MA5'] = df['收盘'].rolling(5).mean()
        df['MA10'] = df['收盘'].rolling(10).mean()

        # 绘图
        fig, ax = plt.subplots(figsize=(10, 6))
        ax.plot(df.index, df['收盘'], label='收盘价', color='blue')
        ax.plot(df.index, df['MA5'], label='5日均线', alpha=0.8, linestyle='--')
        ax.plot(df.index, df['MA10'], label='10日均线', alpha=0.8, linestyle='--')

        # 标注最新价格
        latest = df.iloc[-1]
        ax.scatter(df.index[-1], latest['收盘'], color='red', zorder=5)
        ax.text(df.index[-1], latest['收盘'], f"{latest['收盘']:.2f}", color='red')

        ax.set_title(f"{name}({symbol}) 近30日走势")
        ax.legend()
        ax.grid(True, alpha=0.3)

        # 保存到内存
        buf = BytesIO()
        plt.savefig(buf, format='png', dpi=100, bbox_inches='tight')
        buf.seek(0)
        plt.close(fig)

        # 发送图片
        await bot.send_photo(chat_id=CHAT_ID, photo=buf, caption=f"📊 {name} 走势图")
        logger.info(f"✅ 已发送 {name} 图表")
    except Exception as e:
        logger.error(f"📊 生成图表失败 {symbol}: {e}")
```

> 调用方式：在收盘报告中加入 `await send_kline_chart(symbol, name)`

---

# ✅ 2. 自定义报警：技术指标策略

### 新增策略函数

```python
def check_custom_alerts(symbol: str, df: pd.DataFrame) -> str:
    """检查多种自定义报警"""
    if len(df) < 20:
        return None

    close = df['收盘']
    volume = df['成交量']

    # 策略1：突破20日均线
    ma20 = close.rolling(20).mean()
    if close.iloc[-1] > ma20.iloc[-1] and close.iloc[-2] <= ma20.iloc[-2]:
        return f"🚀 **突破20日均线**：{close.iloc[-1]:.2f} > {ma20.iloc[-1]:.2f}"

    # 策略2：放量上涨 > 5%
    if len(df) >= 2 and close.pct_change().iloc[-1] > 0.05:
        vol_change = volume.pct_change().iloc[-1]
        if vol_change > 1.0:  # 成交量翻倍
            return f"💥 **放量大涨**：+{close.pct_change().iloc[-1]*100:.1f}%，成交量+{vol_change*100:.0f}%"

    # 策略3：MACD金叉（简化版）
    ema12 = close.ewm(span=12).mean()
    ema26 = close.ewm(span=26).mean()
    diff = ema12 - ema26
    dea = diff.ewm(span=9).mean()
    if diff.iloc[-1] > dea.iloc[-1] and diff.iloc[-2] <= dea.iloc[-2]:
        return "🟡 **MACD金叉**：趋势转强"

    return None
```

> 在主循环中调用，对每只股票检查技术信号。

---

# ✅ 3. 数据持久化：JSON 存储

### 创建 `portfolio.json`

```json
{
  "positions": [
    {
      "symbol": "sh600519",
      "name": "贵州茅台",
      "buy_price": 1700.0,
      "quantity": 100,
      "notes": "长期持有"
    }
  ],
  "trades": [
    {
      "symbol": "AAPL",
      "type": "sell",
      "price": 185.0,
      "quantity": 50,
      "time": "2025-04-05T10:30:00"
    }
  ]
}
```

### 读写函数

```python
import json
from datetime import datetime

def load_portfolio():
    try:
        with open("portfolio.json", "r", encoding="utf-8") as f:
            data = json.load(f)
            return data.get("positions", []), data.get("trades", [])
    except FileNotFoundError:
        return [], []

def save_trade(symbol: str, trade_type: str, price: float, quantity: int):
    data = {"positions": load_portfolio()[0], "trades": load_portfolio()[1]}
    trade = {
        "symbol": symbol,
        "type": trade_type,
        "price": price,
        "quantity": quantity,
        "time": datetime.now().isoformat()
    }
    data["trades"].append(trade)
    with open("portfolio.json", "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
```

> 替换原来的 `trade_log` 列表。

---

# ✅ 4. 微信推送替代方案：企业微信机器人

### 步骤：
1. 电脑版微信 → 工作台 → 创建群 → 添加「群机器人」
2. 复制 Webhook URL

### 发送函数

```python
import requests

def send_wechat_message(text: str):
    """通过企业微信机器人发送消息"""
    webhook = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=your-key-here"
    data = {
        "msgtype": "text",
        "text": {
            "content": text
        }
    }
    try:
        requests.post(webhook, json=data, timeout=5)
        logger.info("📤 已发送微信消息")
    except Exception as e:
        logger.error(f"❌ 微信发送失败: {e}")
```

> 可设置双通道：同时发 Telegram 和 企业微信

---

# ✅ 5. 接入券商模拟交易 API（以 掘金量化 为例）

### 注册掘金量化
- 官网：[https://www.myquant.cn](https://www.myquant.cn)
- 创建模拟交易账户
- 获取 `Access Token`

### 示例：下单接口

```python
import requests

def place_paper_order(symbol: str, order_type: str, price: float, quantity: int):
    """
    模拟下单（掘金量化）
    order_type: 'buy' / 'sell'
    """
    url = "https://api.myquant.cn/v2/trade/orders"
    headers = {
        "Authorization": "Bearer YOUR_ACCESS_TOKEN",
        "Content-Type": "application/json"
    }
    # 注意：掘金使用国际代码格式，如 'SHSE.600519'
    exchange = "SHSE" if symbol.startswith("sh") else "SZSE"
    code = symbol[2:] if symbol.startswith("sh") or symbol.startswith("sz") else symbol
    payload = {
        "account_id": "YOUR_ACCOUNT_ID",
        "strategy_id": "YOUR_STRATEGY_ID",
        "symbol": f"{exchange}.{code}",
        "order_type": "2",  # 限价
        "side": 1 if order_type == "buy" else 2,
        "position_effect": 1,
        "price_type": 1,
        "price": price,
        "volume": quantity
    }

    try:
        resp = requests.post(url, json=payload, headers=headers)
        if resp.status_code == 200:
            logger.info(f"✅ 模拟下单成功: {order_type} {quantity}股 {symbol}")
            save_trade(symbol, order_type, price, quantity)
        else:
            logger.error(f"❌ 模拟下单失败: {resp.text}")
    except Exception as e:
        logger.error(f"🌐 API请求失败: {e}")
```

> 可在 `should_buy/sell_simulate` 触发时调用此函数，实现**真实模拟交易**

---

# ✅ 整合建议

你现在可以将系统拆分为模块：

```
stock_monitor/
│
├── config.json           # 配置
├── portfolio.json        # 持仓
├── utils/
│   ├── telegram.py       # Telegram 发送
│   ├── wechat.py         # 企业微信
│   ├── data.py           # 数据获取
│   ├── chart.py          # 图表
│   ├── strategy.py       # 报警策略
│   └── trade_api.py      # 掘金API
├── main.py               # 主循环
└── requirements.txt
```

---

# ✅ 最终系统能力

| 功能 | 状态 |
|------|------|
| 实时行情监控 | ✅ |
| 持仓盈亏计算 | ✅ |
| 收盘总结报告 | ✅ |
| 图表可视化 | ✅ |
| 技术指标报警 | ✅ |
| 数据持久化 | ✅ |
| 微信/Telegram 双推送 | ✅ |
| 模拟交易下单 | ✅ |

---

# 🚀 下一步建议

1. **加入 Web 界面**：用 `Flask` + `Bootstrap` 做个仪表盘
2. **多设备同步**：用 `SQLite` + `Dropbox` 同步数据
3. **AI 预测**：用 `LSTM` 预测短期价格（实验性）
4. **语音提醒**：调用 TTS 播报“茅台突破20日均线！”

---

如果你告诉我：
- 你是否已有掘金/聚宽账号？
- 是否希望加入 Web 界面？
- 是否需要自动更新 `portfolio.json`？

我可以为你生成完整的项目结构和部署脚本。这个系统已经可以作为你的**个人量化助手**长期使用了！📈💻📱

--------------------------------------------------------------------------------
