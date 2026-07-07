package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "进入交互式 Redis 命令行",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("已连接 %s (db=%d)，输入 help 查看帮助，exit 退出\n", cfg.Addr, cfg.DB)
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("redisshow> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				return nil
			}
			if err := runShellLine(line); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}
	},
}

func runShellLine(line string) error {
	parts := splitShellArgs(line)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "help":
		printShellHelp()
		return nil
	case "ping":
		return clientPing()
	case "get":
		if len(parts) != 2 {
			return fmt.Errorf("usage: get <key>")
		}
		val, err := rdb.Get(rootCtx, parts[1]).Result()
		if err != nil {
			return err
		}
		fmt.Println(val)
	case "set":
		if len(parts) < 3 {
			return fmt.Errorf("usage: set <key> <value>")
		}
		return rdb.Set(rootCtx, parts[1], strings.Join(parts[2:], " "), 0).Err()
	case "del":
		if len(parts) < 2 {
			return fmt.Errorf("usage: del <key> [key...]")
		}
		n, err := rdb.Del(rootCtx, parts[1:]...).Result()
		if err != nil {
			return err
		}
		fmt.Printf("deleted %d key(s)\n", n)
	case "keys":
		if len(parts) != 2 {
			return fmt.Errorf("usage: keys <pattern>")
		}
		keys, err := rdb.Keys(rootCtx, parts[1]).Result()
		if err != nil {
			return err
		}
		for _, k := range keys {
			fmt.Println(k)
		}
	case "info":
		section := "default"
		if len(parts) > 1 {
			section = parts[1]
		}
		info, err := rdb.Info(rootCtx, section).Result()
		if err != nil {
			return err
		}
		fmt.Print(info)
	default:
		return fmt.Errorf("unknown command: %s (type help)", parts[0])
	}
	return nil
}

func printShellHelp() {
	fmt.Println(`可用命令:
  ping
  get <key>
  set <key> <value>
  del <key> [key...]
  keys <pattern>
  info [section]
  help
  exit`)
}

func splitShellArgs(line string) []string {
	var args []string
	var cur strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case (c == '"' || c == '\'') && !inQuote:
			inQuote = true
			quoteChar = c
		case inQuote && c == quoteChar:
			inQuote = false
			quoteChar = 0
		case !inQuote && c == ' ':
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}
