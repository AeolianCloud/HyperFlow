package main

import (
	"fmt"

	mysql "github.com/go-sql-driver/mysql"
	"hyperflow/internal/timeutil"
)

// normalizeMySQLDSN 强制 MySQL 连接按固定上海时区处理时间值，并为每个会话设置 +08:00。
func normalizeMySQLDSN(dsn string) (string, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", fmt.Errorf("invalid MYSQL_DSN: %w", err)
	}

	cfg.ParseTime = true
	cfg.Loc = timeutil.ShanghaiLocation()

	if cfg.Params == nil {
		cfg.Params = map[string]string{}
	}
	cfg.Params["time_zone"] = "'+08:00'"

	return cfg.FormatDSN(), nil
}
