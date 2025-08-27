package main

import (
	"github.com/EndlessCheng/mahjong-helper/util"
	"fmt"
	"os"
	"github.com/EndlessCheng/mahjong-helper/util/model"
	"strings"
)

func interact(humanTilesInfo *model.HumanTilesInfo) error {
	if !debugMode {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("内部错误：", err)
			}
		}()
	}

	playerInfo, err := analysisHumanTiles(humanTilesInfo)
	if err != nil {
		return err
	}
	tiles34 := playerInfo.HandTiles34
	leftTiles34 := playerInfo.LeftTiles34
	var tile string
	
	fmt.Println("💡 输入 'help' 查看帮助，'auto-help' 查看自动出牌帮助")
	
	for {
		count := util.CountOfTiles34(tiles34)
		switch count % 3 {
		case 0:
			return fmt.Errorf("参数错误: %d 张牌", count)
		case 1:
			fmt.Print("> 摸 ")
			fmt.Scanf("%s\n", &tile)
			
			// 处理特殊命令
			if handleSpecialCommands(tile) {
				continue
			}
			
			tile, isRedFive, err := util.StrToTile34(tile)
			if err != nil {
				// 让用户重新输入
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if tiles34[tile] == 4 {
				// 让用户重新输入
				fmt.Fprintln(os.Stderr, "不可能摸更多的牌了")
				continue
			}
			if isRedFive {
				playerInfo.NumRedFives[tile/9]++
			}
			leftTiles34[tile]--
			tiles34[tile]++
		case 2:
			fmt.Print("> 切 ")
			fmt.Scanf("%s\n", &tile)
			
			// 处理特殊命令
			if handleSpecialCommands(tile) {
				continue
			}
			
			tile, isRedFive, err := util.StrToTile34(tile)
			if err != nil {
				// 让用户重新输入
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if tiles34[tile] == 0 {
				// 让用户重新输入
				fmt.Fprintln(os.Stderr, "切掉的牌不存在")
				continue
			}
			if isRedFive {
				playerInfo.NumRedFives[tile/9]--
			}
			tiles34[tile]--
			playerInfo.DiscardTiles = append(playerInfo.DiscardTiles, tile) // 仅判断振听用
		}
		if err := analysisPlayerWithRisk(playerInfo, nil); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

// 处理特殊命令
func handleSpecialCommands(input string) bool {
	input = strings.TrimSpace(input)
	
	switch input {
	case "help":
		fmt.Println("💡 可用命令:")
		fmt.Println("  help         - 显示此帮助")
		fmt.Println("  auto-help    - 显示自动出牌帮助")
		fmt.Println("  quit/exit    - 退出交互模式")
		fmt.Println("  牌名         - 输入牌名进行摸牌或切牌")
		fmt.Println("               例如: 1m, 2p, 3s, 1z")
		fmt.Println()
		return true
		
	case "auto-help":
		printAutoPlayerHelp()
		return true
		
	case "quit", "exit":
		fmt.Println("👋 退出交互模式")
		os.Exit(0)
		return true
	}
	
	// 处理自动出牌命令
	if strings.HasPrefix(input, "auto-") {
		return handleAutoPlayerCommand(input)
	}
	
	return false
}
