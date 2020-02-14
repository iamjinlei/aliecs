package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/iamjinlei/aliecs"
	"github.com/iamjinlei/gossh"
)

const (
	loopInterval = 500 * time.Millisecond
)

func acquireInstanceByIp(c *aliyun.EcsClient, region, ip string) (*ecs.Instance, error) {
	instances, err := c.DescribeInstances(aliyun.RegionId(region), ip)
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

func acquireInstanceByName(c *aliyun.EcsClient, region, name string) (*ecs.Instance, error) {
	instances, err := c.DescribeInstances(aliyun.RegionId(region), "")
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, nil
	}

	for _, ins := range instances {
		if ins.InstanceName == name {
			return &ins, nil
		}
	}

	return nil, nil
}

type instanceList []ecs.Instance

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
	op := flag.String("op", "up", "up, down, del, desc, run, reboot")
	idx := flag.Int("idx", 0, "idx")
	flag.Parse()

	cfg, err := aliyun.NewEcsConfig()
	if err != nil {
		aliyun.Error("error creating config: %v", err)
		return
	}

	c, err := aliyun.NewEcsClient(cfg)
	if err != nil {
		aliyun.Error("error creating ecs client: %v", err)
		return
	}
	//c.DescribeZones(ecs.RegionHk, ecs.PostPaid)

	regions := map[aliyun.RegionId]bool{}
	for _, r := range aliyun.ZoneToRegion {
		regions[r] = true
	}
	instances := []ecs.Instance{}
	for r, _ := range regions {
		results, err := c.DescribeInstances(r, "")
		if err != nil {
			aliyun.Error("error describe region %v: %v", r, err)
			return
		}
		instances = append(instances, results...)
	}
	sort.Sort(instanceList(instances))

	schema := "| %-3s | %-15s | %-22s | %-22s | %-18s | %-7s | %-15s | %-17s |"
	rowSeparator := "+-----+-----------------+------------------------+------------------------+--------------------+---------+-----------------+-------------------+"
	lines := []string{
		rowSeparator,
		fmt.Sprintf(schema, "Idx", "ZoneId", "InstanceId", "InstanceName", "InstanceType", "Status", "Public IP", "CreationTime"),
		rowSeparator,
	}

	ip := ""
	region := ""
	name := ""
	if len(instances) == 0 {
		lines = append(lines, fmt.Sprintf(schema, "", "", "", "", "", "", "", ""))
		lines = append(lines, rowSeparator)
	} else {
		for idx, ins := range instances {
			insIp := ""
			if len(ins.PublicIpAddress.IpAddress) > 0 {
				insIp = ins.PublicIpAddress.IpAddress[0]
			}
			lines = append(lines, fmt.Sprintf(schema, fmt.Sprintf("%v", idx), ins.ZoneId, ins.InstanceId, ins.InstanceName, ins.InstanceType, ins.Status, insIp, ins.CreationTime))
			lines = append(lines, rowSeparator)
		}
		if *idx < len(instances) {
			targetIns := &instances[*idx]
			if len(targetIns.PublicIpAddress.IpAddress) > 0 {
				ip = targetIns.PublicIpAddress.IpAddress[0]
			}
			region = targetIns.RegionId
			name = targetIns.InstanceName
		}
	}
	aliyun.Text(strings.Join(lines, "\n"))

	switch *op {
	case "desc":
	case "up":
		instanceIp, isCreated := up(c, cfg)
		if isCreated {
			if err := runCmds(instanceIp, cfg.RootPwd, cfg.InitCmds); err != nil {
				aliyun.Error("error initializing instance environment: %v", err)
			}
		}
	case "reboot":
		if name == "" {
			aliyun.Error("no instance is running")
			return
		}
		reboot(c, region, name)
	case "down":
		if name == "" {
			aliyun.Error("no instance is running")
			return
		}
		down(c, region, name)
	case "del":
		if name == "" {
			aliyun.Error("no instance is running")
			return
		}
		if down(c, region, name) {
			del(c, region, name)
		}
	case "run":
		if ip == "" {
			aliyun.Error("no instance has no public IP")
			return
		}
		if err := runCmds(ip, cfg.RootPwd, cfg.InitCmds); err != nil {
			aliyun.Error("error running commands: %v", err)
		}
	}
}

func up(c *aliyun.EcsClient, cfg *aliyun.EcsCfg) (string, bool) {
	ticker := time.NewTicker(loopInterval)
	pt := aliyun.NewProgressTracker()
	isCreated := false

	instanceName := aliyun.RegionToBr[cfg.Derived.Region] + "-" + time.Now().Format("20060102T1504")
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, string(cfg.Derived.Region), instanceName); err != nil {
			aliyun.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				// instance does NOT exist
				if _, err := c.CreateInstance(cfg, instanceName); err != nil {
					aliyun.Error("error creating instance %v", err)
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
			case string(aliyun.Running):
				if len(ip) == 0 {
					aliyun.Info("public IP address is missing, requesting a new one")
					if _, err := c.BindPublicIp(ins.InstanceId); err != nil {
						aliyun.Error("error binding public ip to instance: %v", err)
					}
				} else {
					aliyun.Info("instance is up running, IP: %s", ip)
					return ip, isCreated
				}
			case string(aliyun.Starting):
				pt.Info("instance is being started up")
			case string(aliyun.Stopping):
				pt.Info("instance is being stopped")
			case string(aliyun.Stopped):
				aliyun.Info("instance is stopped, trying to start it up")
				if err := c.StartInstance(ins.InstanceId); err != nil {
					aliyun.Error("error starting ecs instance: %v", err)
				}
			}
		}
	}

	return "", false
}

func reboot(c *aliyun.EcsClient, region, name string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := aliyun.NewProgressTracker()
	rebooted := false
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliyun.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliyun.Info("instance does NOT exist")
				return false
			}

			if !rebooted {
				if err := c.RebootInstance(ins.InstanceId); err != nil {
					aliyun.Error("error starting ecs instance: %v", err)
				} else {
					rebooted = true
				}
				continue
			}

			// instance exists
			switch ins.Status {
			case string(aliyun.Running):
				aliyun.Info("instance is up running")
				return true
			case string(aliyun.Starting):
				pt.Info("instance is being started up")
			case string(aliyun.Stopping):
				pt.Info("instance is being stopped")
			case string(aliyun.Stopped):
				aliyun.Info("instance is stopped")
			}
		}
	}

	return false
}

func down(c *aliyun.EcsClient, region, name string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := aliyun.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliyun.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliyun.Info("instance does NOT exist")
				return false
			}

			// instance exists
			switch ins.Status {
			case string(aliyun.Running):
				aliyun.Info("instance is running, trying to stop it")
				if err := c.StopInstance(ins.InstanceId); err != nil {
					aliyun.Error("error starting ecs instance: %v", err)
				}
			case string(aliyun.Starting):
				pt.Info("instance is being started up")
			case string(aliyun.Stopping):
				pt.Info("instance is being stopped")
			case string(aliyun.Stopped):
				aliyun.Info("instance is stopped")
				return true
			}
		}
	}

	return false
}

func del(c *aliyun.EcsClient, region, name string) {
	ticker := time.NewTicker(loopInterval)
	pt := aliyun.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliyun.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliyun.Info("instance does NOT exist")
				return
			}

			// instance exists
			aliyun.Info("instance exists, trying to delete it")
			if err := c.DeleteInstance(aliyun.RegionId(region), ins.InstanceId); err != nil {
				aliyun.Error("error deleting ecs instance: %v", err)
				continue
			}
			break
		}
	}

	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliyun.Error("error querying instances: %v", err)
			continue
		} else if ins == nil {
			aliyun.Info("instance is deleted")
			return
		}
		pt.Info("instance is being deleted")
	}
}

func runCmds(ip, rootPwd string, cmds []string) error {
	s, err := gossh.NewSessionWithRetry(ip+":22", "root", rootPwd, "", 10*time.Minute)
	if err != nil {
		return err
	}
	defer s.Close()

	for _, cmd := range cmds {
		c, err := s.Run(cmd)
		if err != nil {
			return err
		}

		for line := range c.CombinedOut() {
			fmt.Printf(string(line) + "\n")
		}
		c.Close()
	}

	return nil
}
