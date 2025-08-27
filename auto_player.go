package main

import (
	"fmt"
	"time"
	"github.com/fatih/color"
	"github.com/EndlessCheng/mahjong-helper/util"
	"github.com/EndlessCheng/mahjong-helper/util/model"
)

// 自动出牌配置
type AutoPlayerConfig struct {
	Enabled           bool    // 是否启用自动出牌
	AutoDiscard       bool    // 自动切牌
	AutoMeld          bool    // 自动鸣牌
	AutoRiichi        bool    // 自动立直
	AutoAgari         bool    // 自动和牌
	MinConfidence     float64 // 最小置信度阈值
	DefenseThreshold  float64 // 防守阈值
	DelaySeconds      float64 // 操作延迟（秒）
	ConfirmActions    bool    // 是否需要确认
	Strategy          string  // 策略：aggressive/balanced/defensive
}

// 默认配置
var defaultAutoPlayerConfig = AutoPlayerConfig{
	Enabled:          false,
	AutoDiscard:      true,
	AutoMeld:         false,
	AutoRiichi:       false,
	AutoAgari:        true,
	MinConfidence:    0.7,
	DefenseThreshold: 0.15,
	DelaySeconds:     1.0,
	ConfirmActions:   true,
	Strategy:         "balanced",
}

// 全局自动出牌配置
var autoPlayerConfig = defaultAutoPlayerConfig

// 决策结果
type Decision struct {
	Action     string  // 动作类型：discard/meld/riichi/agari/pass
	Tile       int     // 相关牌（-1表示无）
	Confidence float64 // 置信度
	Reason     string  // 决策理由
}

// 自动出牌器
type AutoPlayer struct {
	config *AutoPlayerConfig
	lastAction string
}

// 创建新的自动出牌器
func NewAutoPlayer(config *AutoPlayerConfig) *AutoPlayer {
	return &AutoPlayer{
		config: config,
	}
}

// 分析并做出决策
func (ap *AutoPlayer) MakeDecision(playerInfo *model.PlayerInfo, mixedRiskTable riskTable, targetTile int, canMeld bool) Decision {
	if !ap.config.Enabled {
		return Decision{Action: "pass", Confidence: 0, Reason: "自动出牌已禁用"}
	}

	// 检查是否已和牌
	if util.CountOfTiles34(playerInfo.HandTiles34)%3 == 1 {
		shanten, results14, _ := util.CalculateShantenWithImproves14(playerInfo)
		if shanten == -1 {
			return Decision{Action: "agari", Confidence: 1.0, Reason: "已和牌"}
		}
	}

	// 分析手牌状态
	handCount := util.CountOfTiles34(playerInfo.HandTiles34)
	
	switch handCount % 3 {
	case 1: // 需要切牌
		return ap.makeDiscardDecision(playerInfo, mixedRiskTable)
	case 2: // 有选择权（鸣牌或切牌）
		if canMeld && targetTile != -1 {
			return ap.makeMeldDecision(playerInfo, targetTile, mixedRiskTable)
		}
		return ap.makeDiscardDecision(playerInfo, mixedRiskTable)
	}

	return Decision{Action: "pass", Confidence: 0, Reason: "无有效操作"}
}

// 做出切牌决策
func (ap *AutoPlayer) makeDiscardDecision(playerInfo *model.PlayerInfo, mixedRiskTable riskTable) Decision {
	shanten, results14, incShantenResults14 := util.CalculateShantenWithImproves14(playerInfo)
	
	// 评估危险度
	dangerLevel := ap.assessDangerLevel(mixedRiskTable, playerInfo)
	
	var bestDiscard int
	var confidence float64
	var reason string
	
	// 根据策略选择决策
	switch ap.config.Strategy {
	case "aggressive":
		return ap.aggressiveDiscardDecision(playerInfo, results14, incShantenResults14, dangerLevel)
	case "defensive":
		return ap.defensiveDiscardDecision(playerInfo, mixedRiskTable, dangerLevel)
	default: // balanced
		return ap.balancedDiscardDecision(playerInfo, results14, incShantenResults14, mixedRiskTable, dangerLevel)
	}
}

// 激进策略的切牌决策
func (ap *AutoPlayer) aggressiveDiscardDecision(playerInfo *model.PlayerInfo, results14, incShantenResults14 util.Hand14AnalysisResultList, dangerLevel float64) Decision {
	if len(results14) > 0 {
		best := results14[0]
		return Decision{
			Action:     "discard",
			Tile:       best.DiscardTile,
			Confidence: 0.9,
			Reason:     fmt.Sprintf("进攻切牌：%s (进张%d, 打点%d)", util.MahjongZH[best.DiscardTile], best.Result13.Waits.AllCount(), best.Result13.DamaPoint),
		}
	} else if len(incShantenResults14) > 0 {
		best := incShantenResults14[0]
		return Decision{
			Action:     "discard",
			Tile:       best.DiscardTile,
			Confidence: 0.7,
			Reason:     fmt.Sprintf("向听倒退切牌：%s (改良后进张%d)", util.MahjongZH[best.DiscardTile], best.Result13.AvgImproveWaitsCount),
		}
	}
	
	return Decision{Action: "pass", Confidence: 0, Reason: "无法找到合适切牌"}
}

// 防守策略的切牌决策
func (ap *AutoPlayer) defensiveDiscardDecision(playerInfo *model.PlayerInfo, mixedRiskTable riskTable, dangerLevel float64) Decision {
	if dangerLevel > ap.config.DefenseThreshold {
		// 高危险度时选择安全牌
		safestTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
		if safestTile >= 0 {
			return Decision{
				Action:     "discard",
				Tile:       safestTile,
				Confidence: 0.8,
				Reason:     fmt.Sprintf("防守切牌：%s (危险度%.2f)", util.MahjongZH[safestTile], mixedRiskTable[safestTile]),
			}
		}
	}
	
	// 危险度不高时按常规切牌
	return ap.balancedDiscardDecision(playerInfo, nil, nil, mixedRiskTable, dangerLevel)
}

// 平衡策略的切牌决策
func (ap *AutoPlayer) balancedDiscardDecision(playerInfo *model.PlayerInfo, results14, incShantenResults14 util.Hand14AnalysisResultList, mixedRiskTable riskTable, dangerLevel float64) Decision {
	// 高危险度时优先防守
	if dangerLevel > ap.config.DefenseThreshold {
		safestTile := mixedRiskTable.getBestDefenceTile(playerInfo.HandTiles34)
		if safestTile >= 0 {
			return Decision{
				Action:     "discard",
				Tile:       safestTile,
				Confidence: 0.8,
				Reason:     fmt.Sprintf("防守切牌：%s (危险度%.2f)", util.MahjongZH[safestTile], mixedRiskTable[safestTile]),
			}
		}
	}
	
	// 正常情况按进攻切牌
	if len(results14) > 0 {
		best := results14[0]
		confidence := 0.85
		if dangerLevel > 0.1 {
			confidence *= 0.8 // 有危险时降低置信度
		}
		return Decision{
			Action:     "discard",
			Tile:       best.DiscardTile,
			Confidence: confidence,
			Reason:     fmt.Sprintf("平衡切牌：%s (进张%d, 打点%d)", util.MahjongZH[best.DiscardTile], best.Result13.Waits.AllCount(), best.Result13.DamaPoint),
		}
	}
	
	return Decision{Action: "pass", Confidence: 0, Reason: "无法找到合适切牌"}
}

// 做出鸣牌决策
func (ap *AutoPlayer) makeMeldDecision(playerInfo *model.PlayerInfo, targetTile int, mixedRiskTable riskTable) Decision {
	if !ap.config.AutoMeld {
		return Decision{Action: "pass", Confidence: 0, Reason: "自动鸣牌已禁用"}
	}
	
	// 分析鸣牌效果
	shanten, results14, _ := util.CalculateMeld(playerInfo, targetTile, false, true)
	
	if len(results14) > 0 {
		best := results14[0]
		return Decision{
			Action:     "meld",
			Tile:       targetTile,
			Confidence: 0.75,
			Reason:     fmt.Sprintf("鸣牌：%s (向听%d, 进张%d)", util.MahjongZH[targetTile], best.Result13.Shanten, best.Result13.Waits.AllCount()),
		}
	}
	
	return Decision{Action: "pass", Confidence: 0, Reason: "鸣牌效果不佳"}
}

// 评估危险度
func (ap *AutoPlayer) assessDangerLevel(mixedRiskTable riskTable, playerInfo *model.PlayerInfo) float64 {
	if mixedRiskTable == nil {
		return 0
	}
	
	maxRisk := 0.0
	for tile, count := range playerInfo.HandTiles34 {
		if count > 0 {
			risk := mixedRiskTable[tile]
			if risk > maxRisk {
				maxRisk = risk
			}
		}
	}
	
	return maxRisk
}

// 执行决策
func (ap *AutoPlayer) ExecuteDecision(decision Decision) error {
	if decision.Action == "pass" {
		return nil
	}
	
	// 显示决策信息
	ap.displayDecision(decision)
	
	// 如果需要确认
	if ap.config.ConfirmActions {
		if !ap.confirmAction(decision) {
			return fmt.Errorf("用户取消操作")
		}
	}
	
	// 延迟执行
	if ap.config.DelaySeconds > 0 {
		time.Sleep(time.Duration(ap.config.DelaySeconds * float64(time.Second)))
	}
	
	// 执行操作
	switch decision.Action {
	case "discard":
		return ap.executeDiscard(decision.Tile)
	case "meld":
		return ap.executeMeld(decision.Tile)
	case "agari":
		return ap.executeAgari()
	case "riichi":
		return ap.executeRiichi()
	default:
		return fmt.Errorf("未知操作类型: %s", decision.Action)
	}
}

// 显示决策信息
func (ap *AutoPlayer) displayDecision(decision Decision) {
	actionColor := color.FgHiGreen
	if decision.Confidence < 0.8 {
		actionColor = color.FgHiYellow
	}
	if decision.Confidence < 0.6 {
		actionColor = color.FgHiRed
	}
	
	color.New(actionColor).Printf("🤖 自动出牌: %s", decision.Action)
	if decision.Tile >= 0 {
		fmt.Printf(" %s", util.MahjongZH[decision.Tile])
	}
	fmt.Printf(" (置信度: %.1f%%)", decision.Confidence*100)
	fmt.Printf("\n    理由: %s", decision.Reason)
	fmt.Println()
}

// 确认操作
func (ap *AutoPlayer) confirmAction(decision Decision) bool {
	fmt.Print("确认执行此操作? (y/N): ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// 平台操作发送器接口
type ActionSenderInterface interface {
	SendDiscard(tile34 int) error
	SendMeld(meldType int, targetTile int, combination []int) error
	SendRiichi() error
	SendAgari() error
	SendPass() error
}

// 全局操作发送器
var globalActionSender ActionSenderInterface

// 设置操作发送器
func SetActionSender(sender ActionSenderInterface) {
	globalActionSender = sender
}

// 执行切牌操作
func (ap *AutoPlayer) executeDiscard(tile int) error {
	if globalActionSender != nil {
		return globalActionSender.SendDiscard(tile)
	}
	
	// 如果没有设置发送器，只显示模拟信息
	fmt.Printf("模拟执行切牌: %s\n", util.MahjongZH[tile])
	return nil
}

// 执行鸣牌操作
func (ap *AutoPlayer) executeMeld(tile int) error {
	if globalActionSender != nil {
		// 这里需要根据具体情况确定鸣牌类型和组合
		// 暂时使用默认的碰操作
		return globalActionSender.SendMeld(1, tile, []int{tile, tile, tile})
	}
	
	fmt.Printf("模拟执行鸣牌: %s\n", util.MahjongZH[tile])
	return nil
}

// 执行和牌操作
func (ap *AutoPlayer) executeAgari() error {
	if globalActionSender != nil {
		return globalActionSender.SendAgari()
	}
	
	fmt.Println("模拟执行和牌")
	return nil
}

// 执行立直操作
func (ap *AutoPlayer) executeRiichi() error {
	if globalActionSender != nil {
		return globalActionSender.SendRiichi()
	}
	
	fmt.Println("模拟执行立直")
	return nil
}

// 全局自动出牌器实例
var globalAutoPlayer = NewAutoPlayer(&autoPlayerConfig)

// 设置自动出牌配置
func SetAutoPlayerConfig(config AutoPlayerConfig) {
	autoPlayerConfig = config
	globalAutoPlayer.config = &autoPlayerConfig
}

// 获取当前配置
func GetAutoPlayerConfig() AutoPlayerConfig {
	return autoPlayerConfig
}

// 启用/禁用自动出牌
func SetAutoPlayerEnabled(enabled bool) {
	autoPlayerConfig.Enabled = enabled
	globalAutoPlayer.config.Enabled = enabled
	
	if enabled {
		color.HiGreen("🚀 自动出牌已启用")
	} else {
		color.HiYellow("⏸️ 自动出牌已禁用")
	}
}

// 切换自动出牌状态
func ToggleAutoPlayer() {
	SetAutoPlayerEnabled(!autoPlayerConfig.Enabled)
}
