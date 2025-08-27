package majsoul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// 雀魂操作发送器
type ActionSender struct {
	serverURL string
	client    *http.Client
}

// 创建新的操作发送器
func NewActionSender(serverURL string) *ActionSender {
	return &ActionSender{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// 雀魂操作类型
const (
	ActionTypeChi   = 1  // 吃
	ActionTypePon   = 2  // 碰
	ActionTypeKan   = 3  // 杠
	ActionTypeRiichi = 4 // 立直
	ActionTypeAgari = 5  // 和牌
	ActionTypePass  = 6  // 过
)

// 操作请求结构
type ActionRequest struct {
	Type         int    `json:"type"`
	Tile         string `json:"tile,omitempty"`
	Combination  string `json:"combination,omitempty"`
	Pass         bool   `json:"pass,omitempty"`
	Timestamp    int64  `json:"timestamp"`
}

// 发送切牌操作
func (as *ActionSender) SendDiscard(tile34 int) error {
	// 将34种牌转换为雀魂格式
	tileStr := Tile34ToMajsoulStr(tile34)
	
	req := ActionRequest{
		Type:      ActionTypePass, // 切牌在雀魂中通过过操作实现
		Timestamp: time.Now().UnixMilli(),
	}
	
	return as.sendAction(req)
}

// 发送鸣牌操作
func (as *ActionSender) SendMeld(meldType int, targetTile int, combination []int) error {
	var actionType int
	switch meldType {
	case 0: // 吃
		actionType = ActionTypeChi
	case 1: // 碰
		actionType = ActionTypePon
	case 2: // 杠
		actionType = ActionTypeKan
	default:
		return fmt.Errorf("未知的鸣牌类型: %d", meldType)
	}
	
	req := ActionRequest{
		Type:        actionType,
		Tile:        Tile34ToMajsoulStr(targetTile),
		Combination: formatCombination(combination),
		Timestamp:   time.Now().UnixMilli(),
	}
	
	return as.sendAction(req)
}

// 发送立直操作
func (as *ActionSender) SendRiichi() error {
	req := ActionRequest{
		Type:      ActionTypeRiichi,
		Timestamp: time.Now().UnixMilli(),
	}
	
	return as.sendAction(req)
}

// 发送和牌操作
func (as *ActionSender) SendAgari() error {
	req := ActionRequest{
		Type:      ActionTypeAgari,
		Timestamp: time.Now().UnixMilli(),
	}
	
	return as.sendAction(req)
}

// 发送过操作
func (as *ActionSender) SendPass() error {
	req := ActionRequest{
		Type:      ActionTypePass,
		Pass:      true,
		Timestamp: time.Now().UnixMilli(),
	}
	
	return as.sendAction(req)
}

// 发送操作到雀魂服务器
func (as *ActionSender) sendAction(req ActionRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化操作请求失败: %v", err)
	}
	
	resp, err := as.client.Post(as.serverURL+"/action", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("发送操作失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}
	
	return nil
}

// 将34种牌转换为雀魂格式字符串
func Tile34ToMajsoulStr(tile34 int) string {
	// 雀魂使用 "1m", "2m", ..., "9m", "1p", ..., "9p", "1s", ..., "9s", "1z", ..., "7z" 格式
	if tile34 < 0 || tile34 > 33 {
		return ""
	}
	
	suits := []string{"m", "p", "s", "z"}
	numbers := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	
	suit := tile34 / 9
	number := tile34 % 9
	
	if suit == 3 { // 字牌
		return fmt.Sprintf("%dz", number+1)
	} else {
		return fmt.Sprintf("%d%s", number+1, suits[suit])
	}
}

// 格式化组合字符串
func formatCombination(tiles []int) string {
	// 雀魂的组合格式: "1m|2m|3m"
	parts := make([]string, len(tiles))
	for i, tile := range tiles {
		parts[i] = Tile34ToMajsoulStr(tile)
	}
	
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "|"
		}
		result += part
	}
	
	return result
}
