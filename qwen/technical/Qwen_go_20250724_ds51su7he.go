package technical

import (
    "math"
)

// MA 计算简单移动平均线
func MA(data []float64, period int) []float64 {
    if len(data) < period || period <= 0 {
        return make([]float64, len(data))
    }
    
    result := make([]float64, len(data))
    
    // 前period-1个值无法计算，设为0
    for i := 0; i < period-1; i++ {
        result[i] = 0
    }
    
    // 计算移动平均
    for i := period - 1; i < len(data); i++ {
        sum := 0.0
        for j := 0; j < period; j++ {
            sum += data[i-j]
        }
        result[i] = sum / float64(period)
    }
    
    return result
}

// EMA 计算指数移动平均线
func EMA(data []float64, period int) []float64 {
    if len(data) == 0 || period <= 0 {
        return make([]float64, len(data))
    }
    
    result := make([]float64, len(data))
    multiplier := 2.0 / float64(period+1)
    
    // 第一个EMA值等于第一个MA值
    ma := MA(data, period)
    for i := period - 1; i < len(data); i++ {
        if ma[i] != 0 {
            result[i] = ma[i]
            break
        }
    }
    
    // 计算后续EMA值
    for i := period; i < len(data); i++ {
        if result[i-1] != 0 {
            result[i] = data[i]*multiplier + result[i-1]*(1-multiplier)
        } else {
            result[i] = data[i]
        }
    }
    
    return result
}

// MACD 计算MACD指标
type MACDResult struct {
    MACDLine  []float64 // MACD线
    SignalLine []float64 // 信号线
    Histogram []float64 // 柱状图
}

func MACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) MACDResult {
    result := MACDResult{
        MACDLine:  make([]float64, len(data)),
        SignalLine: make([]float64, len(data)),
        Histogram: make([]float64, len(data)),
    }
    
    if len(data) == 0 {
        return result
    }
    
    // 计算快速EMA和慢速EMA
    fastEMA := EMA(data, fastPeriod)
    slowEMA := EMA(data, slowPeriod)
    
    // 计算MACD线（快速EMA - 慢速EMA）
    for i := 0; i < len(data); i++ {
        result.MACDLine[i] = fastEMA[i] - slowEMA[i]
    }
    
    // 计算信号线（MACD线的EMA）
    result.SignalLine = EMA(result.MACDLine, signalPeriod)
    
    // 计算柱状图（MACD线 - 信号线）
    for i := 0; i < len(data); i++ {
        result.Histogram[i] = result.MACDLine[i] - result.SignalLine[i]
    }
    
    return result
}

// RSI 计算相对强弱指数
func RSI(data []float64, period int) []float64 {
    if len(data) <= period || period <= 0 {
        return make([]float64, len(data))
    }
    
    result := make([]float64, len(data))
    
    // 前period个值无法计算，设为0
    for i := 0; i < period; i++ {
        result[i] = 0
    }
    
    // 计算第一个RSI
    gainSum := 0.0
    lossSum := 0.0
    
    for i := 1; i <= period; i++ {
        change := data[i] - data[i-1]
        if change > 0 {
            gainSum += change
        } else {
            lossSum -= change
        }
    }
    
    rs := gainSum / float64(period) / (lossSum / float64(period))
    result[period] = 100 - (100 / (1 + rs))
    
    // 计算后续RSI值（使用平滑算法）
    avgGain := gainSum / float64(period)
    avgLoss := lossSum / float64(period)
    
    for i := period + 1; i < len(data); i++ {
        change := data[i] - data[i-1]
        var gain, loss float64
        if change > 0 {
            gain = change
        } else {
            loss = -change
        }
        
        avgGain = (avgGain*float64(period-1) + gain) / float64(period)
        avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
        
        if avgLoss == 0 {
            result[i] = 100
        } else {
            rs := avgGain / avgLoss
            result[i] = 100 - (100 / (1 + rs))
        }
    }
    
    return result
}

// KDJ 计算随机指标
type KDJResult struct {
    K []float64 // %K值
    D []float64 // %D值
    J []float64 // %J值
}

func KDJ(high, low, close []float64, period, signalPeriod int) KDJResult {
    result := KDJResult{
        K: make([]float64, len(close)),
        D: make([]float64, len(close)),
        J: make([]float64, len(close)),
    }
    
    if len(high) != len(low) || len(high) != len(close) || len(high) <= period {
        return result
    }
    
    // 计算K值
    for i := period - 1; i < len(close); i++ {
        highestHigh := high[i]
        lowestLow := low[i]
        
        // 找到周期内的最高价和最低价
        for j := i - period + 1; j <= i; j++ {
            if high[j] > highestHigh {
                highestHigh = high[j]
            }
            if low[j] < lowestLow {
                lowestLow = low[j]
            }
        }
        
        // 计算%K值
        if highestHigh-lowestLow != 0 {
            result.K[i] = (close[i] - lowestLow) / (highestHigh - lowestLow) * 100
        } else {
            result.K[i] = 50
        }
    }
    
    // 计算D值（K值的移动平均）
    result.D = MA(result.K, signalPeriod)
    
    // 计算J值
    for i := 0; i < len(close); i++ {
        result.J[i] = 3*result.K[i] - 2*result.D[i]
    }
    
    return result
}

// BOLL 计算布林带（额外提供）
type BOLLResult struct {
    Upper  []float64 // 上轨
    Middle []float64 // 中轨（MA）
    Lower  []float64 // 下轨
}

func BOLL(data []float64, period int, multiplier float64) BOLLResult {
    result := BOLLResult{
        Upper:  make([]float64, len(data)),
        Middle: make([]float64, len(data)),
        Lower:  make([]float64, len(data)),
    }
    
    if len(data) < period || period <= 0 {
        return result
    }
    
    // 计算中轨（MA）
    result.Middle = MA(data, period)
    
    // 计算标准差和上下轨
    for i := period - 1; i < len(data); i++ {
        if result.Middle[i] == 0 {
            continue
        }
        
        // 计算标准差
        variance := 0.0
        for j := 0; j < period; j++ {
            diff := data[i-j] - result.Middle[i]
            variance += diff * diff
        }
        stdDev := math.Sqrt(variance / float64(period))
        
        // 计算上下轨
        result.Upper[i] = result.Middle[i] + multiplier*stdDev
        result.Lower[i] = result.Middle[i] - multiplier*stdDev
    }
    
    return result
}

// 使用示例
func Example() {
    // 示例数据
    prices := []float64{10, 12, 13, 11, 15, 16, 14, 18, 17, 19, 20, 18, 22, 21, 23}
    high := []float64{11, 13, 14, 12, 16, 17, 15, 19, 18, 20, 21, 19, 23, 22, 24}
    low := []float64{9, 11, 12, 10, 14, 15, 13, 17, 16, 18, 19, 17, 21, 20, 22}
    
    // 计算各种指标
    ma5 := MA(prices, 5)
    ema12 := EMA(prices, 12)
    macd := MACD(prices, 12, 26, 9)
    rsi14 := RSI(prices, 14)
    kdj := KDJ(high, low, prices, 9, 3)
    boll := BOLL(prices, 20, 2.0)
    
    // 输出结果（这里只是示例，实际使用时可以根据需要处理结果）
    _ = ma5
    _ = ema12
    _ = macd
    _ = rsi14
    _ = kdj
    _ = boll
}