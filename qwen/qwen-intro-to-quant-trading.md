# 🤵 请向一个初学者介绍量化交易

当然可以！下面是一个面向初学者的量化交易介绍，尽量用简单易懂的语言来解释：

---

### 什么是量化交易？

**量化交易**（Quantitative Trading）是一种利用数学模型、统计方法和计算机程序来进行投资决策的交易方式。简单来说，就是“用数据和算法来买卖股票、期货、加密货币等金融资产”。

和传统交易靠“感觉”或“经验”不同，量化交易强调的是**客观、系统、可重复**的策略。

---

### 举个生活中的例子 🌰

想象你在经营一家奶茶店，你发现：

- 每到下午2点，销量会突然上升；
- 天气越热，卖得越好；
- 周末比工作日多卖出30%。

于是你决定：  
“以后每天下午1点提前多准备原料，天气预报说高温就多备货，周末多安排人手。”

这就是一种**基于数据的决策**。

量化交易也是一样：  
通过分析历史数据，找出市场中的“规律”，然后让计算机自动执行买卖。

---

### 量化交易的核心步骤

1. **提出想法（策略构思）**  
   比如：“过去5天涨得少的股票，接下来可能会补涨。”

2. **数据收集**  
   获取股票的历史价格、成交量、财务数据等。

3. **建立模型（写代码）**  
   用Python等编程语言把想法变成可计算的规则。

4. **回测（Backtesting）**  
   把策略放到过去的数据中测试，看它在过去能不能赚钱。

5. **优化与验证**  
   调整参数，避免“过度拟合”（即只适合过去，不适合未来）。

6. **实盘交易**  
   在真实市场中用小资金试运行，逐步扩大规模。

7. **风险管理**  
   控制每笔交易的风险，比如单笔亏损不超过总资金的1%。

---

### 量化交易常用工具

- **编程语言**：Python 最常用（有丰富的金融库，如 Pandas、NumPy、Backtrader）。
- **数据源**：Tushare（A股）、Yahoo Finance（美股）、AKShare（免费中文数据）等。
- **交易平台**：聚宽（JoinQuant）、掘金量化、QuantConnect（支持自动交易）。

---

### 量化交易的优点

✅ 客观理性：避免情绪干扰（贪婪、恐惧）。  
✅ 执行高效：计算机可以24小时盯盘，快速交易。  
✅ 可复制性强：策略成功后可以应用到多个市场。

---

### 初学者常见误区

❌ 认为“写个程序就能稳赚” → 实际上大多数策略会失效。  
❌ 过度优化（过拟合）→ 在过去表现很好，但未来亏钱。  
❌ 忽视交易成本（手续费、滑点）→ 小利润可能被吃掉。

---

### 给初学者的建议

1. **先学基础**：了解股票、K线、收益率、风险等基本概念。
2. **学点Python**：不需要成为程序员，但要会处理数据和写简单逻辑。
3. **从小策略开始**：比如“均线交叉策略”（简单但经典）。
4. **动手做回测**：在聚宽或掘金量化上模拟试试。
5. **保持学习**：读经典书籍，如《打开量化投资的黑箱》《主动投资组合管理》。

---

### 一个简单的策略例子：双均线策略

- 当短期均线（如5日）上穿长期均线（如20日）→ 买入。
- 当短期均线下穿长期均线 → 卖出。

这个策略可以用几十行Python代码实现，并进行回测。

---

### 总结

量化交易 = **想法 + 数据 + 代码 + 验证 + 执行**

它不是“印钞机”，而是一种更科学的投资方法。  
对初学者来说，最重要的是：**动手实践，从小做起，持续学习**。

---

如果你感兴趣，我可以一步步带你写一个简单的量化策略代码，你想试试吗？😊

--------------------------------------------------------------------------------

# 🤵 请一步步带我写一个简单的量化策略代码

太好了！我们这就一步步来写一个**简单的量化交易策略代码**——**双均线策略（Dual Moving Average Strategy）**，用 Python 实现，并在 A 股市场的一只股票上做**回测演示**。

我们会使用：
- `akshare`：免费的中文金融数据接口（获取股票数据）
- `pandas` 和 `numpy`：数据处理
- `matplotlib`：画图看结果

---

### ✅ 第一步：安装必要的库

打开你的命令行（终端或 Anaconda Prompt），运行以下命令：

```bash
pip install akshare pandas numpy matplotlib
```

> 如果你用的是 Jupyter Notebook，也可以在代码单元格前面加 `!` 来安装：
>
> ```python
> !pip install akshare pandas numpy matplotlib
> ```

---

### ✅ 第二步：导入所需库

```python
import akshare as ak
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# 设置中文字体显示（可选）
plt.rcParams['font.sans-serif'] = ['SimHei']
plt.rcParams['axes.unicode_minus'] = False
```

---

### ✅ 第三步：获取股票数据

我们以 **贵州茅台（600519）** 为例，获取它的日线数据。

```python
# 获取贵州茅台的后复权日线数据
stock_code = "600519"
stock_data = ak.stock_zh_a_hist(symbol=stock_code, period="daily", adjust="hfq")

# 检查数据
print(stock_data.head())
```

输出类似：

```
        日期     开盘     收盘     最高     最低     成交量  成交额
0  2015-01-05  192.0  194.8  195.5  191.5  123456  789012
...
```

我们需要把列名改成英文方便处理：

```python
stock_data.rename(columns={
    "日期": "date",
    "开盘": "open",
    "收盘": "close",
    "最高": "high",
    "最低": "low",
    "成交量": "volume"
}, inplace=True)

# 将日期设为索引
stock_data["date"] = pd.to_datetime(stock_data["date"])
stock_data.set_index("date", inplace=True)
```

---

### ✅ 第四步：计算移动平均线

我们计算：
- 短期均线：5日收盘价均值（fast MA）
- 长期均线：20日收盘价均值（slow MA）

```python
# 计算移动平均线
stock_data['ma_short'] = stock_data['close'].rolling(window=5).mean()
stock_data['ma_long'] = stock_data['close'].rolling(window=20).mean()
```

---

### ✅ 第五步：生成交易信号

我们定义：
- **买入信号**：短期均线上穿长期均线（前一天短 < 长，今天短 > 长）
- **卖出信号**：短期均线下穿长期均线

```python
# 判断是否上穿
stock_data['signal'] = 0  # 0 表示无操作
stock_data['signal'] = np.where(stock_data['ma_short'] > stock_data['ma_long'], 1, 0)
stock_data['position'] = stock_data['signal'].diff()  # 位置变化

# 买入信号：上穿（从0变1）
stock_data['buy_signal'] = np.where(stock_data['position'] == 1, 1, 0)

# 卖出信号：下穿（从1变0）
stock_data['sell_signal'] = np.where(stock_data['position'] == -1, 1, 0)
```

---

### ✅ 第六步：可视化策略（画图）

我们画出股价、两条均线和买卖点。

```python
plt.figure(figsize=(14, 7))
plt.plot(stock_data['close'], label='收盘价', alpha=0.7)
plt.plot(stock_data['ma_short'], label='5日均线', alpha=0.7)
plt.plot(stock_data['ma_long'], label='20日均线', alpha=0.7)

# 标出买入信号
buy_signals = stock_data[stock_data['buy_signal'] == 1]
plt.scatter(buy_signals.index, buy_signals['close'], marker='^', color='green', s=100, label='买入信号')

# 标出卖出信号
sell_signals = stock_data[stock_data['sell_signal'] == 1]
plt.scatter(sell_signals.index, sell_signals['close'], marker='v', color='red', s=100, label='卖出信号')

plt.title('贵州茅台 - 双均线策略交易信号')
plt.xlabel('日期')
plt.ylabel('价格')
plt.legend()
plt.grid(True)
plt.tight_layout()
plt.show()
```

---

### ✅ 第七步：简单回测（计算收益）

我们模拟：
- 初始资金：10万元
- 每次全仓买入/卖出（简化版）

```python
initial_capital = 100000  # 初始资金
cash = initial_capital
holdings = 0  # 持有股数
position_value = 0  # 持仓价值
equity_curve = []  # 净值曲线

# 遍历每一行数据
for i in range(len(stock_data)):
    row = stock_data.iloc[i]
    
    # 如果是买入信号，且没有持仓
    if row['buy_signal'] == 1 and cash > 0:
        holdings = cash / row['close']  # 全仓买入
        cash = 0
        print(f"买入: {row.name.date()} 价格={row['close']:.2f}")
    
    # 如果是卖出信号，且有持仓
    elif row['sell_signal'] == 1 and holdings > 0:
        cash = holdings * row['close']  # 全仓卖出
        holdings = 0
        print(f"卖出: {row.name.date()} 价格={row['close']:.2f}")
    
    # 计算当前总资产（现金 + 持仓价值）
    if holdings > 0:
        position_value = holdings * row['close']
    else:
        position_value = 0
    
    total_value = cash + position_value
    equity_curve.append(total_value)

# 添加净值曲线到数据中
stock_data['equity'] = equity_curve
```

---

### ✅ 第八步：计算最终收益

```python
final_value = equity_curve[-1]
profit = final_value - initial_capital
return_rate = (profit / initial_capital) * 100

print(f"\n初始资金: {initial_capital:,.0f}")
print(f"最终资金: {final_value:,.2f}")
print(f"总收益: {profit:+,.2f}")
print(f"收益率: {return_rate:+.2f}%")
```

---

### 🎉 完整代码汇总（可直接运行）

```python
import akshare as ak
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

plt.rcParams['font.sans-serif'] = ['SimHei']
plt.rcParams['axes.unicode_minus'] = False

# 1. 获取数据
stock_data = ak.stock_zh_a_hist(symbol="600519", period="daily", adjust="hfq")
stock_data.rename(columns={
    "日期": "date", "开盘": "open", "收盘": "close",
    "最高": "high", "最低": "low", "成交量": "volume"
}, inplace=True)
stock_data["date"] = pd.to_datetime(stock_data["date"])
stock_data.set_index("date", inplace=True)

# 2. 计算均线
stock_data['ma_short'] = stock_data['close'].rolling(window=5).mean()
stock_data['ma_long'] = stock_data['close'].rolling(window=20).mean()

# 3. 生成信号
stock_data['signal'] = np.where(stock_data['ma_short'] > stock_data['ma_long'], 1, 0)
stock_data['position'] = stock_data['signal'].diff()
stock_data['buy_signal'] = np.where(stock_data['position'] == 1, 1, 0)
stock_data['sell_signal'] = np.where(stock_data['position'] == -1, 1, 0)

# 4. 画图
plt.figure(figsize=(14, 7))
plt.plot(stock_data['close'], label='收盘价')
plt.plot(stock_data['ma_short'], label='5日均线')
plt.plot(stock_data['ma_long'], label='20日均线')
buy_signals = stock_data[stock_data['buy_signal'] == 1]
sell_signals = stock_data[stock_data['sell_signal'] == 1]
plt.scatter(buy_signals.index, buy_signals['close'], marker='^', color='green', s=100, label='买入')
plt.scatter(sell_signals.index, sell_signals['close'], marker='v', color='red', s=100, label='卖出')
plt.title('双均线策略 - 贵州茅台')
plt.legend()
plt.grid(True)
plt.show()

# 5. 回测
initial_capital = 100000
cash = initial_capital
holdings = 0
equity_curve = []

for i in range(len(stock_data)):
    row = stock_data.iloc[i]
    if row['buy_signal'] == 1 and cash > 0:
        holdings = cash / row['close']
        cash = 0
        print(f"买入: {row.name.date()} 价格={row['close']:.2f}")
    elif row['sell_signal'] == 1 and holdings > 0:
        cash = holdings * row['close']
        holdings = 0
        print(f"卖出: {row.name.date()} 价格={row['close']:.2f}")
    
    total_value = cash + (holdings * row['close'])
    equity_curve.append(total_value)

stock_data['equity'] = equity_curve
final_value = equity_curve[-1]
return_rate = (final_value - initial_capital) / initial_capital * 100

print(f"\n初始资金: {initial_capital:,.0f}")
print(f"最终资金: {final_value:,.2f}")
print(f"收益率: {return_rate:+.2f}%")
```

---

### 🔍 下一步你可以尝试：

1. **加入手续费**（比如 0.1%）
2. **限制持仓时间**
3. **用更多股票测试**
4. **使用专业回测框架**（如 `backtrader` 或 `zipline`）
5. **尝试其他策略**：布林带、RSI 超买超卖等

---

如果你想，我可以帮你把这段代码升级成使用 `backtrader` 框架的专业回测版本，包含手续费、滑点、夏普比率等指标。需要吗？ 😊

--------------------------------------------------------------------------------

# 🤵 上面的代码用的是双均线策略，有没有其它的策略或指标？

当然有！双均线策略只是量化交易的“入门级”策略。实际上，市场上有**成百上千种**量化策略和指标，适用于不同市场环境（牛市、震荡市、熊市）。

下面我为你介绍 **5 种常见且适合初学者尝试的量化策略/指标**，每种都配有**核心思想 + 公式/逻辑 + 适用场景**，你可以用类似前面的方法在 Python 中实现。

---

## ✅ 1. RSI 超买超卖策略（相对强弱指数）

### 📌 核心思想：
当股票涨得太猛（超买），可能会回调；跌得太狠（超卖），可能会反弹。RSI 就是用来衡量“涨跌强度”的指标。

### 🔢 计算方式：
- 通常计算过去 **14 天**的涨跌幅，得出一个 0~100 的值。
- **RSI > 70**：超买 → 考虑卖出
- **RSI < 30**：超卖 → 考虑买入

### 🧠 策略逻辑：
```python
if RSI < 30:
    买入
elif RSI > 70:
    卖出
```

### ✅ 优点：
- 适合震荡市
- 简单直观

### ⚠️ 注意：
- 在强势趋势中容易“抄底抄在半山腰”或“逃顶逃得太早”

---

## ✅ 2. 布林带策略（Bollinger Bands）

### 📌 核心思想：
价格通常在“波动通道”内运行。布林带由三条线组成：
- 中轨：20日移动平均线
- 上轨：中轨 + 2倍标准差
- 下轨：中轨 - 2倍标准差

### 🧠 策略逻辑（常见版本）：
```python
if 价格触及下轨（且低于下轨）:
    买入（认为会回归中轨）

if 价格触及上轨（且高于上轨）:
    卖出（认为会回落）
```

也可以反过来用（趋势跟踪）：
```python
if 价格突破上轨并继续上涨:
    跟进做多（认为趋势开始）
```

### ✅ 优点：
- 同时捕捉波动率和价格位置
- 可用于震荡市或趋势市

---

## ✅ 3. MACD 策略（异同移动平均线）

### 📌 核心思想：
通过快慢均线的“差离”来判断趋势强弱和转折点。

### 🔢 组成部分：
- **DIF** = 12日EMA - 26日EMA（快线）
- **DEA** = DIF 的 9日EMA（慢线）
- **MACD柱** = (DIF - DEA) × 2

### 🧠 策略逻辑：
```python
if DIF 上穿 DEA:
    买入信号（金叉）

if DIF 下穿 DEA:
    卖出信号（死叉）
```

也可以结合0轴过滤：
```python
只有当 DIF 和 DEA 都在0轴以上时才做多（避免在下跌趋势中乱交易）
```

### ✅ 优点：
- 对趋势反应较灵敏
- 广泛用于股票、期货、加密货币

---

## ✅ 4. 动量策略（Momentum Strategy）

### 📌 核心思想：
“强者恒强”——过去涨得好的股票，未来一段时间可能继续涨。

### 🧠 策略逻辑：
```python
计算每只股票过去 20 天的收益率

选出涨幅最高的前 N 只股票 → 买入
持有一段时间（如10天）后卖出
```

> 例如：每月初选涨幅前10的股票，持有一个月。

### ✅ 优点：
- 在趋势明显的市场中表现很好（如牛市）

### ⚠️ 风险：
- 趋势反转时可能大幅回撤

---

## ✅ 5. 均值回归策略（Mean Reversion）

### 📌 核心思想：
“物极必反”——价格偏离正常水平后，终将回归均值。

### 🧠 常见实现方式：
- 使用 RSI、布林带判断超买超卖
- 或计算“价格与移动平均线的偏离度”：
  
```python
z-score = (当前价格 - 近20日均值) / 近20日标准差

if z-score < -2:  # 显著低于均值 → 买入
if z-score > 2:   # 显著高于均值 → 卖出
```

### ✅ 优点：
- 适合震荡行情
- 数学基础强（统计套利的基础）

---

## 🎯 如何选择策略？

| 市场环境 | 推荐策略 |
|--------|---------|
| **单边上涨（牛市）** | 动量策略、MACD趋势跟踪 |
| **单边下跌（熊市）** | 空仓、或做空动量（高级） |
| **横盘震荡** | RSI、布林带、均值回归 |
| **波动加大** | 布林带突破、波动率策略 |

---

## 🚀 下一步建议：你可以尝试

1. **用前面的代码框架**，把 RSI 或布林带策略写出来
2. **比较不同策略在同一股票上的表现**
3. **组合多个指标**（比如：只有 RSI < 30 且价格在布林带下轨才买入）

---

### 💡 示例：如何计算 RSI（Python 片段）

```python
def calculate_rsi(data, window=14):
    delta = data['close'].diff()
    gain = (delta.where(delta > 0, 0)).rolling(window).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window).mean()
    rs = gain / loss
    rsi = 100 - (100 / (1 + rs))
    return rsi

# 使用
stock_data['rsi'] = calculate_rsi(stock_data)
```

然后就可以加信号：

```python
stock_data['buy_rsi'] = np.where((stock_data['rsi'] < 30), 1, 0)
stock_data['sell_rsi'] = np.where((stock_data['rsi'] > 70), 1, 0)
```

---

如果你想，我可以：
- 帮你写出完整的 **RSI策略代码**
- 或做一个 **双策略对比回测**（比如：双均线 vs RSI）
- 或教你如何用 `backtrader` 框架做专业回测

你最想先尝试哪一个？告诉我，我们继续深入！ 😊

--------------------------------------------------------------------------------

# 🤵 帮我写出完整的 RSI策略代码
做一个 双策略对比回测

太棒了！我们现在就来完成两个目标：

---

## ✅ 目标一：写出完整的 **RSI 超买超卖策略代码**

## ✅ 目标二：做一个 **双均线 vs RSI 策略的回测对比**

我们将使用 **贵州茅台（600519）** 的历史数据，对两种策略在同一时间段内进行简单回测，并绘制**净值曲线对比图**，直观看出哪个表现更好。

---

### 📦 所需库（确保已安装）

```bash
pip install akshare pandas numpy matplotlib
```

---

### 🚀 完整代码：双均线 vs RSI 策略回测对比

```python
import akshare as ak
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['SimHei']
plt.rcParams['axes.unicode_minus'] = False

# ==================== 第一步：获取数据 ====================
stock_code = "600519"
stock_name = "贵州茅台"

# 获取后复权日线数据
data = ak.stock_zh_a_hist(symbol=stock_code, period="daily", adjust="hfq")

# 重命名并设置索引
data.rename(columns={
    "日期": "date",
    "开盘": "open",
    "收盘": "close",
    "最高": "high",
    "最低": "low",
    "成交量": "volume"
}, inplace=True)
data["date"] = pd.to_datetime(data["date"])
data.set_index("date", inplace=True)

print(f"已获取 {stock_name} 数据，时间范围：{data.index[0].date()} 至 {data.index[-1].date()}")

# ==================== 第二步：定义 RSI 计算函数 ====================
def calculate_rsi(series, window=14):
    delta = series.diff()
    gain = delta.where(delta > 0, 0)
    loss = -delta.where(delta < 0, 0)
    avg_gain = gain.rolling(window).mean()
    avg_loss = loss.rolling(window).mean()
    rs = avg_gain / avg_loss
    rsi = 100 - (100 / (1 + rs))
    return rsi

# ==================== 第三步：计算指标 ====================
# 双均线策略所需
data['ma_short'] = data['close'].rolling(5).mean()
data['ma_long'] = data['close'].rolling(20).mean()

# RSI 策略所需
data['rsi'] = calculate_rsi(data['close'], window=14)

# ==================== 第四步：生成交易信号 ====================

# --- 策略1：双均线策略 ---
data['signal_ma'] = 0
data['signal_ma'] = np.where(data['ma_short'] > data['ma_long'], 1, 0)
data['position_ma'] = data['signal_ma'].diff()  # 1:买入, -1:卖出

# --- 策略2：RSI 策略 ---
data['signal_rsi'] = 0
# 超卖买入，超买卖出
data.loc[data['rsi'] < 30, 'signal_rsi'] = 1  # 买入信号
data.loc[data['rsi'] > 70, 'signal_rsi'] = 0  # 卖出信号（退出）
# 保持持仓直到超买
data['position_rsi'] = data['signal_rsi'].diff()  # 用于触发买卖

# ==================== 第五步：回测引擎（简化版）====================
def backtest_strategy(position_series, close_prices, initial_capital=100000):
    """
    简化回测函数
    position_series: 1=持有, 0=空仓, 由 position.diff() 触发买卖
    close_prices: 收盘价序列
    """
    cash = initial_capital
    holdings = 0
    equity_curve = []

    for i in range(len(position_series)):
        pos_change = position_series.iloc[i]  # 信号变化
        price = close_prices.iloc[i]

        if pos_change == 1 and cash > 0:        # 买入
            holdings = cash / price
            cash = 0
        elif pos_change == -1 and holdings > 0: # 卖出
            cash = holdings * price
            holdings = 0

        total_value = cash + holdings * price
        equity_curve.append(total_value)

    return np.array(equity_curve)

# 运行两个策略的回测
equity_ma = backtest_strategy(data['position_ma'], data['close'])
equity_rsi = backtest_strategy(data['position_rsi'], data['close'])

# ==================== 第六步：结果分析与可视化 ====================
dates = data.index

plt.figure(figsize=(14, 8))

# 净值曲线对比
plt.plot(dates, equity_ma, label='双均线策略', linewidth=2)
plt.plot(dates, equity_rsi, label='RSI 策略', linewidth=2)
plt.plot(dates, data['close'] / data['close'].iloc[0] * 100000, label='买入持有', linestyle='--', alpha=0.7)

plt.title(f'{stock_name} 回测对比：双均线 vs RSI vs 买入持有')
plt.xlabel('日期')
plt.ylabel('账户净值（元）')
plt.legend()
plt.grid(True, alpha=0.3)
plt.tight_layout()
plt.show()

# ==================== 第七步：输出绩效指标 ====================
def calculate_performance(equity, trading_days=252):
    total_return = (equity[-1] - equity[0]) / equity[0] * 100
    annual_return = ((equity[-1] / equity[0]) ** (trading_days / len(equity)) - 1) * 100
    drawdown = (equity - np.maximum.accumulate(equity)) / np.maximum.accumulate(equity)
    max_drawdown = drawdown.min() * 100

    return {
        '总收益(%)': round(total_return, 2),
        '年化收益(%)': round(annual_return, 2),
        '最大回撤(%)': round(max_drawdown, 2)
    }

perf_ma = calculate_performance(equity_ma)
perf_rsi = calculate_performance(equity_rsi)
perf_buy_hold = calculate_performance(data['close'].values / data['close'].iloc[0] * 100000)

# 汇总表格
results = pd.DataFrame([perf_ma, perf_rsi, perf_buy_hold], 
                       index=['双均线策略', 'RSI策略', '买入持有'])
print("\n📊 策略绩效对比：")
print(results)

```

---

### 📈 输出内容包括：

1. **净值曲线图**：三种方式的收益对比（双均线、RSI、买入持有）
2. **绩效表格**：总收益、年化收益、最大回撤

---

### 🧪 示例输出可能长这样（实际结果取决于数据区间）：

```
📊 策略绩效对比：
              总收益(%)  年化收益(%)  最大回撤(%)
双均线策略         280.50     18.20     -38.45
RSI策略           195.30     13.80     -45.20
买入持有          320.00     20.10     -48.00
```

> ⚠️ 注意：过去表现 ≠ 未来收益！这只是教学演示。

---

### 🔍 策略特点总结

| 策略 | 优点 | 缺点 | 适合市场 |
|------|------|------|----------|
| **双均线** | 趋势跟踪，牛市表现好 | 震荡市频繁交易、手续费高 | 趋势市（牛市） |
| **RSI** | 震荡市高抛低吸 | 牛市中容易“卖飞” | 震荡市、回调市 |
| **买入持有** | 简单省心 | 回撤大 | 长期牛市 |

---

### 🚀 下一步你可以尝试：

1. **加入手续费**（如 0.1% 买入 + 0.1% 卖出）
2. **限制持仓时间**（比如最多持10天）
3. **组合策略**：只有 RSI < 30 且价格在布林带下轨才买入
4. **多只股票测试**：比较不同股票上的表现
5. **用 backtrader 框架** 做更专业的回测（支持仓位、滑点、夏普比率等）

---

如果你想，我现在就可以：
- 帮你加上 **交易手续费**
- 或升级到 **Backtrader 框架版本**
- 或做一个 **策略组合（双均线 + RSI 过滤）**

你最想先看哪一个？告诉我，我们继续深入！ 😊
--------------------------------------------------------------------------------
