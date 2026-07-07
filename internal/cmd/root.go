package cmd

import (
	"context"
	"fmt"
	"os"

	"redisshow/internal/client"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var (
	cfg        client.Config
	rdb        *redis.Client
	rootCtx    = context.Background()
	rootCmd    = &cobra.Command{
		Use:   "redisshow",
		Short: "Redis 命令行客户端",
		Long:  "redisshow 是一个可直接运行的 Redis 命令行工具，支持常用读写与交互模式。",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			rdb = client.New(cfg)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if rdb != nil {
				_ = rdb.Close()
			}
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	defaultCfg := client.ConfigFromEnv()

	rootCmd.PersistentFlags().StringVarP(&cfg.Addr, "addr", "a", defaultCfg.Addr, "Redis 地址，如 localhost:6379")
	rootCmd.PersistentFlags().StringVarP(&cfg.Password, "password", "p", defaultCfg.Password, "Redis 密码")
	rootCmd.PersistentFlags().IntVarP(&cfg.DB, "db", "n", defaultCfg.DB, "Redis 数据库编号")

	rootCmd.AddCommand(
		pingCmd,
		getCmd,
		setCmd,
		delCmd,
		keysCmd,
		infoCmd,
		typeCmd,
		ttlCmd,
		hgetCmd,
		hsetCmd,
		lpushCmd,
		lrangeCmd,
		shellCmd,
		tuiCmd,
	)
}
