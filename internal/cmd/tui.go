package cmd

import (
	"fmt"

	"redisshow/internal/tui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "打开 TUI 窗口（左侧键列表，右侧详情）",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rdb.Ping(rootCtx).Err(); err != nil {
			return fmt.Errorf("无法连接 Redis: %w", err)
		}
		return tui.Run(rootCtx, rdb, cfg)
	},
}
