package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const maxDisplayLines = 500

func scanKeys(ctx context.Context, rdb *redis.Client, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, next, err := rdb.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func formatKeyDetail(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	keyType, err := rdb.Type(ctx, key).Result()
	if err != nil {
		return "", err
	}

	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[yellow]Key[-]    %s\n", key)
	fmt.Fprintf(&b, "[yellow]Type[-]   %s\n", keyType)
	fmt.Fprintf(&b, "[yellow]TTL[-]    %s\n", formatTTL(ttl))
	b.WriteString("\n[yellow]Value[-]\n")

	valueText, err := formatValue(ctx, rdb, key, keyType)
	if err != nil {
		return "", err
	}
	b.WriteString(valueText)
	return b.String(), nil
}

func formatTTL(ttl time.Duration) string {
	switch {
	case ttl < 0:
		return "永久"
	case ttl == 0:
		return "已过期"
	default:
		return ttl.String()
	}
}

func formatValue(ctx context.Context, rdb *redis.Client, key, keyType string) (string, error) {
	switch keyType {
	case "string":
		return formatString(ctx, rdb, key)
	case "hash":
		return formatHash(ctx, rdb, key)
	case "list":
		return formatList(ctx, rdb, key)
	case "set":
		return formatSet(ctx, rdb, key)
	case "zset":
		return formatZSet(ctx, rdb, key)
	case "none":
		return "(键不存在)", nil
	default:
		return fmt.Sprintf("(不支持的类型: %s)", keyType), nil
	}
}

func formatString(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return truncateLines(val, maxDisplayLines), nil
}

func formatHash(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	fields, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for i, field := range keys {
		if i >= maxDisplayLines {
			fmt.Fprintf(&b, "\n... 还有 %d 个字段未显示", len(keys)-maxDisplayLines)
			break
		}
		fmt.Fprintf(&b, "%s = %s\n", field, fields[field])
	}
	if len(keys) == 0 {
		b.WriteString("(空哈希)")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func formatList(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return "", err
	}
	items, err := rdb.LRange(ctx, key, 0, maxDisplayLines-1).Result()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for i, item := range items {
		fmt.Fprintf(&b, "[%d] %s\n", i, item)
	}
	if length > int64(len(items)) {
		fmt.Fprintf(&b, "\n... 还有 %d 个元素未显示", length-int64(len(items)))
	}
	if len(items) == 0 {
		b.WriteString("(空列表)")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func formatSet(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	members, err := rdb.SMembers(ctx, key).Result()
	if err != nil {
		return "", err
	}
	sort.Strings(members)

	var b strings.Builder
	for i, member := range members {
		if i >= maxDisplayLines {
			fmt.Fprintf(&b, "\n... 还有 %d 个成员未显示", len(members)-maxDisplayLines)
			break
		}
		fmt.Fprintf(&b, "%s\n", member)
	}
	if len(members) == 0 {
		b.WriteString("(空集合)")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func formatZSet(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	members, err := rdb.ZRangeWithScores(ctx, key, 0, maxDisplayLines-1).Result()
	if err != nil {
		return "", err
	}
	total, err := rdb.ZCard(ctx, key).Result()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for _, member := range members {
		fmt.Fprintf(&b, "%g  %s\n", member.Score, member.Member)
	}
	if total > int64(len(members)) {
		fmt.Fprintf(&b, "\n... 还有 %d 个成员未显示", total-int64(len(members)))
	}
	if len(members) == 0 {
		b.WriteString("(空有序集合)")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func truncateLines(text string, maxLines int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}
	truncated := strings.Join(lines[:maxLines], "\n")
	return fmt.Sprintf("%s\n\n... 内容过长，已截断显示前 %d 行", truncated, maxLines)
}
