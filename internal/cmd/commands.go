package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"redisshow/internal/client"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "测试 Redis 连接",
	RunE: func(cmd *cobra.Command, args []string) error {
		return clientPing()
	},
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "获取字符串键值",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := rdb.Get(rootCtx, args[0]).Result()
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置字符串键值",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := strings.Join(args[1:], " ")

		ttl, _ := cmd.Flags().GetDuration("ttl")
		if ttl > 0 {
			return rdb.Set(rootCtx, key, value, ttl).Err()
		}
		return rdb.Set(rootCtx, key, value, 0).Err()
	},
}

var delCmd = &cobra.Command{
	Use:   "del <key> [key...]",
	Short: "删除一个或多个键",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := rdb.Del(rootCtx, args...).Result()
		if err != nil {
			return err
		}
		fmt.Printf("deleted %d key(s)\n", n)
		return nil
	},
}

var keysCmd = &cobra.Command{
	Use:   "keys <pattern>",
	Short: "按模式列出键",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keys, err := rdb.Keys(rootCtx, args[0]).Result()
		if err != nil {
			return err
		}
		for _, k := range keys {
			fmt.Println(k)
		}
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info [section]",
	Short: "查看 Redis 服务信息",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		section := "default"
		if len(args) > 0 {
			section = args[0]
		}
		info, err := rdb.Info(rootCtx, section).Result()
		if err != nil {
			return err
		}
		fmt.Print(info)
		return nil
	},
}

var typeCmd = &cobra.Command{
	Use:   "type <key>",
	Short: "查看键类型",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := rdb.Type(rootCtx, args[0]).Result()
		if err != nil {
			return err
		}
		fmt.Println(t)
		return nil
	},
}

var ttlCmd = &cobra.Command{
	Use:   "ttl <key>",
	Short: "查看键剩余过期时间（秒）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ttl, err := rdb.TTL(rootCtx, args[0]).Result()
		if err != nil {
			return err
		}
		fmt.Println(int64(ttl.Seconds()))
		return nil
	},
}

var hgetCmd = &cobra.Command{
	Use:   "hget <key> <field>",
	Short: "获取哈希字段值",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := rdb.HGet(rootCtx, args[0], args[1]).Result()
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var hsetCmd = &cobra.Command{
	Use:   "hset <key> <field> <value>",
	Short: "设置哈希字段值",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return rdb.HSet(rootCtx, args[0], args[1], args[2]).Err()
	},
}

var lpushCmd = &cobra.Command{
	Use:   "lpush <key> <value> [value...]",
	Short: "从列表左侧插入元素",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		values := make([]interface{}, len(args)-1)
		for i, v := range args[1:] {
			values[i] = v
		}
		n, err := rdb.LPush(rootCtx, key, values...).Result()
		if err != nil {
			return err
		}
		fmt.Println(n)
		return nil
	},
}

var lrangeCmd = &cobra.Command{
	Use:   "lrange <key> <start> <stop>",
	Short: "获取列表区间元素",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		start, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid start: %w", err)
		}
		stop, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid stop: %w", err)
		}
		vals, err := rdb.LRange(rootCtx, args[0], start, stop).Result()
		if err != nil {
			return err
		}
		for _, v := range vals {
			fmt.Println(v)
		}
		return nil
	},
}

func init() {
	setCmd.Flags().Duration("ttl", 0, "过期时间，如 30s、5m、1h")
}

func clientPing() error {
	return client.Ping(rootCtx, rdb)
}
