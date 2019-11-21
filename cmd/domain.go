package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/iamjinlei/aliecs"
)

const (
	loopInterval = 500 * time.Millisecond
)

func main() {
	op := flag.String("op", "list", "list, check")
	domain := flag.String("domain", "", "domain name")
	flag.Parse()

	cfg, err := aliyun.NewEcsConfig()
	if err != nil {
		aliyun.Error("error creating config: %v", err)
		return
	}

	c, err := aliyun.NewDomainClient(cfg.ToDomainCfg())
	if err != nil {
		aliyun.Error("error creating ecs client: %v", err)
		return
	}

	statusMap := map[int]string{
		1:  "可注册",
		3:  "预登记",
		4:  "可删除预订",
		0:  "不可注册",
		-1: "异常",
		-2: "暂停注册",
		-3: "黑名单",
	}

	switch *op {
	case "list":
		domains, err := c.ListDomains()
		if err != nil {
			aliyun.Error("error listing domains: %v", err)
			return
		}

		schema := "| %-3s | %-20s | %-6s | %-8s |"
		rowSeparator := "+-----+----------------------+--------+----------+"
		lines := []string{
			rowSeparator,
			fmt.Sprintf(schema, "Idx", "Name", "Status", "Type"),
			rowSeparator,
		}
		for idx, d := range domains {
			lines = append(lines, fmt.Sprintf(schema, fmt.Sprintf("%v", idx), d.DomainName, d.DomainStatus, d.DomainType))
		}
		lines = append(lines, rowSeparator)
		aliyun.Text(strings.Join(lines, "\n"))

	case "check":
		name, status, price, err := c.CheckDomain(*domain)
		if err != nil {
			aliyun.Error("error listing domains: %v", err)
			return
		}

		aliyun.Text("Domain = %v, Status = %v, Price = %v", name, statusMap[status], price)
	}
}
