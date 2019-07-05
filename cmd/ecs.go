package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/iamjinlei/aliecs"
	"github.com/iamjinlei/gossh"
)

const (
	loopInterval = 500 * time.Millisecond
)

func acquireInstanceByIp(c *aliecs.Client, region, ip string) (*ali.Instance, error) {
	instances, err := c.DescribeInstances(aliecs.RegionId(region), ip)
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

func acquireInstanceByName(c *aliecs.Client, region, name string) (*ali.Instance, error) {
	instances, err := c.DescribeInstances(aliecs.RegionId(region), "")
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
	op := flag.String("op", "up", "up, down, del, desc, run, reboot")
	idx := flag.Int("idx", 0, "idx")
	flag.Parse()

	cfg, err := aliecs.NewConfig()
	if err != nil {
		aliecs.Error("error creating config: %v", err)
		return
	}

	c, err := aliecs.NewClient(cfg)
	if err != nil {
		aliecs.Error("error creating ecs client: %v", err)
		return
	}
	//c.DescribeZones(ecs.RegionHk, ecs.PostPaid)

	regions := map[aliecs.RegionId]bool{}
	for _, r := range aliecs.ZoneToRegion {
		regions[r] = true
	}
	instances := []ali.Instance{}
	for r, _ := range regions {
		results, err := c.DescribeInstances(r, "")
		if err != nil {
			aliecs.Error("error describe region %v: %v", r, err)
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
	aliecs.Text(strings.Join(lines, "\n"))

	switch *op {
	case "desc":
	case "up":
		instanceIp, isCreated := up(c, cfg)
		if isCreated {
			if err := runCmds(instanceIp, cfg.RootPwd, cfg.InitCmds); err != nil {
				aliecs.Error("error initializing instance environment: %v", err)
			}
		}
	case "reboot":
		if name == "" {
			aliecs.Error("no instance is running")
			return
		}
		reboot(c, region, name)
	case "down":
		if name == "" {
			aliecs.Error("no instance is running")
			return
		}
		down(c, region, name)
	case "del":
		if name == "" {
			aliecs.Error("no instance is running")
			return
		}
		if down(c, region, name) {
			del(c, region, name)
		}
	case "run":
		if ip == "" {
			aliecs.Error("no instance has no public IP")
			return
		}
		if err := runCmds(ip, cfg.RootPwd, cfg.InitCmds); err != nil {
			aliecs.Error("error running commands: %v", err)
		}
	}
}

func up(c *aliecs.Client, cfg *aliecs.Cfg) (string, bool) {
	ticker := time.NewTicker(loopInterval)
	pt := aliecs.NewProgressTracker()
	isCreated := false

	instanceName := aliecs.RegionToBr[cfg.Derived.Region] + "-" + time.Now().Format("20060102T1504")
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, string(cfg.Derived.Region), instanceName); err != nil {
			aliecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				// instance does NOT exist
				if _, err := c.CreateInstance(cfg, instanceName); err != nil {
					aliecs.Error("error creating instance %v", err)
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
			case string(aliecs.Running):
				if len(ip) == 0 {
					aliecs.Info("public IP address is missing, requesting a new one")
					if _, err := c.BindPublicIp(ins.InstanceId); err != nil {
						aliecs.Error("error binding public ip to instance: %v", err)
					}
				} else {
					aliecs.Info("instance is up running, IP: %s", ip)
					return ip, isCreated
				}
			case string(aliecs.Starting):
				pt.Info("instance is being started up")
			case string(aliecs.Stopping):
				pt.Info("instance is being stopped")
			case string(aliecs.Stopped):
				aliecs.Info("instance is stopped, trying to start it up")
				if err := c.StartInstance(ins.InstanceId); err != nil {
					aliecs.Error("error starting ecs instance: %v", err)
				}
			}
		}
	}

	return "", false
}

func reboot(c *aliecs.Client, region, name string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := aliecs.NewProgressTracker()
	rebooted := false
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliecs.Info("instance does NOT exist")
				return false
			}

			if !rebooted {
				if err := c.RebootInstance(ins.InstanceId); err != nil {
					aliecs.Error("error starting ecs instance: %v", err)
				} else {
					rebooted = true
				}
				continue
			}

			// instance exists
			switch ins.Status {
			case string(aliecs.Running):
				aliecs.Info("instance is up running")
				return true
			case string(aliecs.Starting):
				pt.Info("instance is being started up")
			case string(aliecs.Stopping):
				pt.Info("instance is being stopped")
			case string(aliecs.Stopped):
				aliecs.Info("instance is stopped")
			}
		}
	}

	return false
}

func down(c *aliecs.Client, region, name string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := aliecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliecs.Info("instance does NOT exist")
				return false
			}

			// instance exists
			switch ins.Status {
			case string(aliecs.Running):
				aliecs.Info("instance is running, trying to stop it")
				if err := c.StopInstance(ins.InstanceId); err != nil {
					aliecs.Error("error starting ecs instance: %v", err)
				}
			case string(aliecs.Starting):
				pt.Info("instance is being started up")
			case string(aliecs.Stopping):
				pt.Info("instance is being stopped")
			case string(aliecs.Stopped):
				aliecs.Info("instance is stopped")
				return true
			}
		}
	}

	return false
}

func del(c *aliecs.Client, region, name string) {
	ticker := time.NewTicker(loopInterval)
	pt := aliecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				aliecs.Info("instance does NOT exist")
				return
			}

			// instance exists
			aliecs.Info("instance exists, trying to delete it")
			if err := c.DeleteInstance(aliecs.RegionId(region), ins.InstanceId); err != nil {
				aliecs.Error("error deleting ecs instance: %v", err)
				continue
			}
			break
		}
	}

	for range ticker.C {
		if ins, err := acquireInstanceByName(c, region, name); err != nil {
			aliecs.Error("error querying instances: %v", err)
			continue
		} else if ins == nil {
			aliecs.Info("instance is deleted")
			return
		}
		pt.Info("instance is being deleted")
	}
}

func runCmds(ip, rootPwd string, cmds []string) error {
	s, err := gossh.NewSessionWithRetry(ip+":22", "root", rootPwd, 10*time.Minute)
	if err != nil {
		return err
	}
	defer s.Close()

	for _, cmd := range cmds {
		c, err := s.Run(cmd)
		if err != nil {
			return err
		}

		c.TailLog()
		c.Close()
	}

	return nil
}
