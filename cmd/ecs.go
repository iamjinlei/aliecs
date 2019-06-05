package main

import (
	"flag"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/iamjinlei/ecs"
)

const (
	loopInterval = 500 * time.Millisecond
)

func acquireInstance(c *ecs.Client, region, ip string) (*ali.Instance, error) {
	instances, err := c.DescribeInstances(ecs.RegionId(region), ip)
	if err != nil {
		return nil, err
	}

	if len(instances) > 1 {
		return nil, fmt.Errorf("unexpected # of instances %v", len(instances))
	}

	if len(instances) == 0 {
		return nil, nil
	}

	return &instances[0], nil
}

type instanceList []ali.Instance

func (s instanceList) Len() int {
	return len(s)
}

func (s instanceList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s instanceList) Less(i, j int) bool {
	if s[i].ZoneId == s[j].ZoneId {
		return s[i].InstanceId < s[j].InstanceId
	}
	return s[i].ZoneId < s[j].ZoneId
}

func main() {
	op := flag.String("op", "up", "up, down, del, desc, run")
	idx := flag.Int("idx", 0, "idx")
	flag.Parse()

	cfg, err := ecs.NewConfig()
	if err != nil {
		ecs.Error("error creating config: %v", err)
		return
	}

	c, err := ecs.NewClient(cfg)
	if err != nil {
		ecs.Error("error creating ecs client: %v", err)
		return
	}
	//c.DescribeZones(ecs.RegionHk, ecs.PostPaid)

	regions := map[ecs.RegionId]bool{}
	for _, r := range ecs.ZoneToRegion {
		regions[r] = true
	}
	instances := []ali.Instance{}
	for r, _ := range regions {
		results, err := c.DescribeInstances(r, "")
		if err != nil {
			ecs.Error("error describe region %v: %v", r, err)
			return
		}
		instances = append(instances, results...)
	}
	sort.Sort(instanceList(instances))

	schema := "| %-3s | %-15s | %-22s | %-18s | %-7s | %-15s | %-17s |"
	rowSeparator := "+-----+-----------------+------------------------+--------------------+---------+-----------------+-------------------+"
	lines := []string{
		rowSeparator,
		fmt.Sprintf(schema, "Idx", "ZoneId", "InstanceId", "InstanceType", "Status", "Public IP", "CreationTime"),
		rowSeparator,
	}

	ip := ""
	region := ""
	if len(instances) == 0 {
		lines = append(lines, fmt.Sprintf(schema, "", "", "", "", "", "", ""))
		lines = append(lines, rowSeparator)
	} else {
		for idx, ins := range instances {
			insIp := ""
			if len(ins.PublicIpAddress.IpAddress) > 0 {
				insIp = ins.PublicIpAddress.IpAddress[0]
			}
			lines = append(lines, fmt.Sprintf(schema, fmt.Sprintf("%v", idx), ins.ZoneId, ins.InstanceId, ins.InstanceType, ins.Status, insIp, ins.CreationTime))
			lines = append(lines, rowSeparator)
		}
		if *idx < len(instances) {
			targetIns := &instances[*idx]
			if len(targetIns.PublicIpAddress.IpAddress) > 0 {
				ip = targetIns.PublicIpAddress.IpAddress[0]
			}
			region = targetIns.RegionId
		}
	}
	ecs.Text(strings.Join(lines, "\n"))

	switch *op {
	case "desc":
	case "up":
		instanceIp, isCreated := up(c, cfg)
		if isCreated {
			ticker := time.NewTicker(loopInterval)
			pt := ecs.NewProgressTracker()
			for range ticker.C {
				if _, err := ecs.NewSsh(instanceIp, cfg.RootPwd); err == nil {
					break
				}
				pt.Info("waiting instance to be ready")
			}

			if err := runCmds(instanceIp, cfg.RootPwd, cfg.InitCmds); err != nil {
				ecs.Error("error initializing instance environment: %v", err)
			}
		}
	case "down":
		if ip == "" {
			ecs.Error("no instance is running")
			return
		}
		down(c, region, ip)
	case "del":
		if ip == "" {
			ecs.Error("no instance is running")
			return
		}
		ecs.Info("deleting %v %v", region, ip)
		if down(c, region, ip) {
			del(c, region, ip)
		}
	case "run":
		if ip == "" {
			ecs.Error("no instance is running")
			return
		}
		if err := runCmds(ip, cfg.RootPwd, cfg.InitCmds); err != nil {
			ecs.Error("error running commands: %v", err)
		}
	case "proxy":
		if ip == "" {
			ecs.Error("target instance IP is missing")
			return
		}

		go func() {
			ecs.Info("start ssh tunnel")
			cmd := exec.Command("ssh", "-CN", "-D", "8080", "root@"+ip)
			res, err := cmd.CombinedOutput()
			if err != nil {
				ecs.Error("error executing ssh tunnel: %v", err)
				return
			}
			ecs.Info(string(res))
		}()

		if err := runCmds(ip, cfg.RootPwd, []string{ecs.RunProxy()}); err != nil {
			ecs.Error("error running proxy %v:", err)
		}
	}
}

func up(c *ecs.Client, cfg *ecs.Cfg) (string, bool) {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	isCreated := false
	instanceName := time.Now().Format("2006-01-02 15:04:05")
	for range ticker.C {
		if ins, err := acquireInstance(c, string(cfg.Derived.Region), instanceName); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				// instance does NOT exist
				if _, err := c.CreateInstance(cfg, instanceName); err != nil {
					ecs.Error("error creating instance %v", err)
				}
				isCreated = true
				continue
			}

			// instance exists
			ip := ""
			if len(ins.PublicIpAddress.IpAddress) > 0 {
				ip = ins.PublicIpAddress.IpAddress[0]
			}
			switch ins.Status {
			case string(ecs.Running):
				if len(ip) == 0 {
					ecs.Info("public IP address is missing, requesting a new one")
					if _, err := c.BindPublicIp(ins.InstanceId); err != nil {
						ecs.Error("error binding public ip to instance: %v", err)
					}
				} else {
					ecs.Info("instance is up running, IP: %s", ip)
					return ip, isCreated
				}
			case string(ecs.Starting):
				pt.Info("instance is being started up")
			case string(ecs.Stopping):
				pt.Info("instance is being stopped")
			case string(ecs.Stopped):
				ecs.Info("instance is stopped, trying to start it up")
				if err := c.StartInstance(ins.InstanceId); err != nil {
					ecs.Error("error starting ecs instance: %v", err)
				}
			}
		}
	}

	return "", false
}

func down(c *ecs.Client, region, ip string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstance(c, region, ip); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				ecs.Info("instance does NOT exist")
				return false
			}

			// instance exists
			switch ins.Status {
			case string(ecs.Running):
				ecs.Info("instance is running, trying to stop it")
				if err := c.StopInstance(ins.InstanceId); err != nil {
					ecs.Error("error starting ecs instance: %v", err)
				}
			case string(ecs.Starting):
				pt.Info("instance is being started up")
			case string(ecs.Stopping):
				pt.Info("instance is being stopped")
			case string(ecs.Stopped):
				ecs.Info("instance is stopped")
				return true
			}
		}
	}

	return false
}

func del(c *ecs.Client, region, ip string) {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstance(c, region, ip); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				ecs.Info("instance does NOT exist")
				return
			}

			// instance exists
			ecs.Info("instance exists, trying to delete it")
			if err := c.DeleteInstance(ecs.RegionId(region), ins.InstanceId); err != nil {
				ecs.Error("error deleting ecs instance: %v", err)
				continue
			}
			break
		}
	}

	for range ticker.C {
		if ins, err := acquireInstance(c, region, ip); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else if ins == nil {
			ecs.Info("instance is deleted")
			return
		}
		pt.Info("instance is being deleted")
	}
}

func runCmds(ip, rootPwd string, cmds []string) error {
	for _, cmd := range cmds {
		s, err := ecs.NewSsh(ip, rootPwd)
		if err != nil {
			return err
		}

		stopSignal := make(chan bool)
		go func() {
			for {
				select {
				case <-stopSignal:
					return
				default:
					ecs.Info(strings.TrimSpace(string(s.Next())))
				}
			}
		}()

		err = s.Run(cmd)

		s.Close()
		stopSignal <- true

		if err != nil {
			return err
		}
	}

	return nil
}
