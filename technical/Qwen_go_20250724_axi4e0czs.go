package main

import (
    "fmt"
    "math"
)

// CrossType 交叉类型枚举
type CrossType int

const (
    NoCross     CrossType = iota // 无交叉
    GoldenCross                  // 金叉：短期均线上穿长期均线
    DeathCross                   // 死叉：短期均线下穿长期均线
)

// String 实现CrossType的字符串表示
func (c CrossType) String() string {
    switch c {
    case GoldenCross:
        return "金叉"
    case DeathCross:
        return "死叉"
    default:
        return "无交叉"
    }
}

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

// DetectMACross 检测均线交叉
func DetectMACross(shortMA, longMA []float64) []CrossType {
    if len(shortMA) != len(longMA) {
        panic("均线数据长度不一致")
    }
    
    crosses := make([]CrossType, len(shortMA))
    
    // 至少需要2个数据点才能检测交叉
    for i := 1; i < len(shortMA); i++ {
        // 跳过无效数据（值为0的情况）
        if shortMA[i-1] == 0 || shortMA[i] == 0 || longMA[i-1] == 0 || longMA[i] == 0 {
            crosses[i] = NoCross
            continue
        }
        
        // 判断是否发生交叉
        prevShortAbove := shortMA[i-1] > longMA[i-1]
        currShortAbove := shortMA[i] > longMA[i]
        
        if !prevShortAbove && currShortAbove {
            // 短期均线上穿长期均线 - 金叉
            crosses[i] = GoldenCross
        } else if prevShortAbove && !currShortAbove {
            // 短期均线下穿长期均线 - 死叉
            crosses[i] = DeathCross
        } else {
            // 无交叉
            crosses[i] = NoCross
        }
    }
    
    return crosses
}

// DetectMACrossWithPrices 使用价格数据检测均线交叉
func DetectMACrossWithPrices(prices []float64, shortPeriod, longPeriod int) ([]CrossType, []float64, []float64) {
    shortMA := MA(prices, shortPeriod)
    longMA := MA(prices, longPeriod)
    crosses := DetectMACross(shortMA, longMA)
    
    return crosses, shortMA, longMA
}

// MACrossResult 交叉检测结果结构
type MACrossResult struct {
    Date       string    // 日期（如果有的话）
    Price      float64   // 当前价格
    ShortMA    float64   // 短期均线值
    LongMA     float64   // 长期均线值
    CrossType  CrossType // 交叉类型
}

// AnalyzeMACross 详细分析均线交叉
func AnalyzeMACross(prices []float64, shortPeriod, longPeriod int, dates []string) []MACrossResult {
    shortMA := MA(prices, shortPeriod)
    longMA := MA(prices, longPeriod)
    crosses := DetectMACross(shortMA, longMA)
    
    results := make([]MACrossResult, len(prices))
    
    for i := 0; i < len(prices); i++ {
        results[i] = MACrossResult{
            Price:     prices[i],
            ShortMA:   shortMA[i],
            LongMA:    longMA[i],
            CrossType: crosses[i],
        }
        
        // 如果提供了日期，则设置日期
        if i < len(dates) {
            results[i].Date = dates[i]
        } else {
            results[i].Date = fmt.Sprintf("Day%d", i+1)
        }
    }
    
    return results
}

// GetRecentCrosses 获取最近的交叉信号
func GetRecentCrosses(prices []float64, shortPeriod, longPeriod int, dates []string) []MACrossResult {
    results := AnalyzeMACross(prices, shortPeriod, longPeriod, dates)
    
    // 从后往前查找最近的交叉信号
    recentCrosses := make([]MACrossResult, 0)
    
    for i := len(results) - 1; i >= 0; i-- {
        if results[i].CrossType != NoCross {
            recentCrosses = append(recentCrosses, results[i])
            // 只返回最近的3个交叉信号
            if len(recentCrosses) >= 3 {
                break
            }
        }
    }
    
    // 反转数组以按时间顺序排列
    for i, j := 0, len(recentCrosses)-1; i < j; i, j = i+1, j-1 {
        recentCrosses[i], recentCrosses[j] = recentCrosses[j], recentCrosses[i]
    }
    
    return recentCrosses
}

// GetCurrentTrend 获取当前趋势状态
func GetCurrentTrend(prices []float64, shortPeriod, longPeriod int) string {
    shortMA := MA(prices, shortPeriod)
    longMA := MA(prices, longPeriod)
    
    if len(shortMA) == 0 || len(longMA) == 0 {
        return "趋势不明"
    }
    
    // 获取最后一个有效的均线值
    var lastShortMA, lastLongMA float64
    for i := len(shortMA) - 1; i >= 0; i-- {
        if shortMA[i] != 0 && longMA[i] != 0 {
            lastShortMA = shortMA[i]
            lastLongMA = longMA[i]
            break
        }
    }
    
    if lastShortMA > lastLongMA {
        return "多头趋势"
    } else if lastShortMA < lastLongMA {
        return "空头趋势"
    } else {
        return "震荡整理"
    }
}

// CalculateMACDivergence 计算均线发散度（额外功能）
func CalculateMACDivergence(shortMA, longMA []float64) []float64 {
    if len(shortMA) != len(longMA) {
        panic("均线数据长度不一致")
    }
    
    divergence := make([]float64, len(shortMA))
    
    for i := 0; i < len(shortMA); i++ {
        if shortMA[i] != 0 && longMA[i] != 0 {
            divergence[i] = math.Abs(shortMA[i] - longMA[i])
        }
    }
    
    return divergence
}

// 使用示例
func main() {
    // 示例日线收盘价数据
    prices := []float64{
        10.0, 10.5, 11.0, 10.8, 11.2, 11.5, 11.8, 12.0, 11.9, 12.2,
        12.5, 12.8, 13.0, 12.9, 13.2, 13.5, 13.3, 13.0, 12.8, 13.1,
        13.4, 13.7, 14.0, 14.2, 14.5, 14.3, 14.0, 13.8, 14.1, 14.4,
    }
    
    // 对应的日期（可选）
    dates := []string{
        "2024-01-01", "2024-01-02", "2024-01-03", "2024-01-04", "2024-01-05",
        "2024-01-06", "2024-01-07", "2024-01-08", "2024-01-09", "2024-01-10",
        "2024-01-11", "2024-01-12", "2024-01-13", "2024-01-14", "2024-01-15",
        "2024-01-16", "2024-01-17", "2024-01-18", "2024-01-19", "2024-01-20",
        "2024-01-21", "2024-01-22", "2024-01-23", "2024-01-24", "2024-01-25",
        "2024-01-26", "2024-01-27", "2024-01-28", "2024-01-29", "2024-01-30",
    }
    
    fmt.Println("=== 均线交叉分析 ===")
    
    // 方法1：直接检测5日和20日均线交叉
    crosses, ma5, ma20 := DetectMACrossWithPrices(prices, 5, 20)
    
    fmt.Println("\n交叉信号详情：")
    for i := 0; i < len(prices); i++ {
        if crosses[i] != NoCross {
            fmt.Printf("日期: %s, 价格: %.2f, 5日均线: %.2f, 20日均线: %.2f, 信号: %s\n",
                dates[i], prices[i], ma5[i], ma20[i], crosses[i])
        }
    }
    
    // 方法2：详细分析
    fmt.Println("\n=== 详细分析 ===")
    results := AnalyzeMACross(prices, 5, 20, dates)
    
    // 显示最近几天的数据
    fmt.Println("最近10天的均线分析：")
    for i := len(results) - 10; i < len(results); i++ {
        if i >= 0 {
            fmt.Printf("日期: %s, 价格: %.2f, 5日MA: %.2f, 20日MA: %.2f, 状态: %s\n",
                results[i].Date, results[i].Price, results[i].ShortMA, results[i].LongMA, results[i].CrossType)
        }
    }
    
    // 方法3：获取最近的交叉信号
    fmt.Println("\n=== 最近的交叉信号 ===")
    recentCrosses := GetRecentCrosses(prices, 5, 20, dates)
    for _, cross := range recentCrosses {
        fmt.Printf("日期: %s, 价格: %.2f, 信号: %s\n", cross.Date, cross.Price, cross.CrossType)
    }
    
    // 方法4：当前趋势判断
    fmt.Println("\n=== 当前趋势 ===")
    trend := GetCurrentTrend(prices, 5, 20)
    fmt.Printf("当前趋势: %s\n", trend)
    
    // 方法5：计算发散度
    divergence := CalculateMACDivergence(ma5, ma20)
    fmt.Println("\n=== 最近5天的均线发散度 ===")
    for i := len(divergence) - 5; i < len(divergence); i++ {
        if i >= 0 && divergence[i] != 0 {
            fmt.Printf("日期: %s, 发散度: %.2f\n", dates[i], divergence[i])
        }
    }
}

// 工具函数：查找特定类型的交叉
func FindCrossByType(prices []float64, shortPeriod, longPeriod int, crossType CrossType, limit int) []int {
    crosses, _, _ := DetectMACrossWithPrices(prices, shortPeriod, longPeriod)
    
    indices := make([]int, 0)
    
    for i := len(crosses) - 1; i >= 0; i-- {
        if crosses[i] == crossType {
            indices = append(indices, i)
            if len(indices) >= limit {
                break
            }
        }
    }
    
    // 反转数组以按时间顺序排列
    for i, j := 0, len(indices)-1; i < j; i, j = i+1, j-1 {
        indices[i], indices[j] = indices[j], indices[i]
    }
    
    return indices
}

// 工具函数：获取交叉统计信息
func GetCrossStatistics(prices []float64, shortPeriod, longPeriod int) map[string]int {
    crosses, _, _ := DetectMACrossWithPrices(prices, shortPeriod, longPeriod)
    
    stats := make(map[string]int)
    stats["金叉"] = 0
    stats["死叉"] = 0
    stats["无交叉"] = 0
    
    for _, cross := range crosses {
        switch cross {
        case GoldenCross:
            stats["金叉"]++
        case DeathCross:
            stats["死叉"]++
        default:
            stats["无交叉"]++
        }
    }
    
    return stats
}