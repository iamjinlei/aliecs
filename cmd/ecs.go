package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/iamjinlei/ecs"
)

const (
	loopInterval = 500 * time.Millisecond
)

func acquireInstance(c *ecs.Client, instanceName string) (*ali.Instance, error) {
	instances, err := c.DescribeInstances(instanceName)
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

func main() {
	op := flag.String("op", "up", "up, down, del, desc")
	instanceName := flag.String("name", "hk", "instanceName")
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
	rowSeparator := "+------------------------+--------------------+---------+-----------------+-------------------+"
	lines := []string{
		rowSeparator,
		fmt.Sprintf("| %-22s | %-18s | %-7s | %-15s | %-17s |", "InstanceId", "InstanceType", "Status", "Public IP", "CreationTime"),
		rowSeparator,
	}
	instances, err := c.DescribeInstances("*")
	if len(instances) == 0 {
		lines = append(lines, "|                        |                    |         |                   |")
		lines = append(lines, rowSeparator)
	} else {
		for _, ins := range instances {
			ip := ""
			if len(ins.PublicIpAddress.IpAddress) > 0 {
				ip = ins.PublicIpAddress.IpAddress[0]
			}
			lines = append(lines, fmt.Sprintf("| %-22s | %-18s | %-7s | %-15s | %-17s |", ins.InstanceId, ins.InstanceType, ins.Status, ip, ins.CreationTime))
			lines = append(lines, rowSeparator)
		}
	}
	ecs.Text(strings.Join(lines, "\n"))

	switch *op {
	case "desc":
	case "up":
		ip, isCreated := up(c, cfg, *instanceName)
		if isCreated {
			if err := initEnv(ip, cfg.RootPwd, cfg.InitCmds); err != nil {
				ecs.Error("error initializing instance environment %v:", err)
			}
		}
	case "down":
		down(c, *instanceName)
	case "del":
		if down(c, *instanceName) {
			del(c, *instanceName)
		}
	}
}

func up(c *ecs.Client, cfg *ecs.Cfg, instanceName string) (string, bool) {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	isCreated := false
	for range ticker.C {
		if ins, err := acquireInstance(c, instanceName); err != nil {
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

func down(c *ecs.Client, instanceName string) bool {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstance(c, instanceName); err != nil {
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
func del(c *ecs.Client, instanceName string) {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	for range ticker.C {
		if ins, err := acquireInstance(c, instanceName); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else {
			if ins == nil {
				ecs.Info("instance does NOT exist")
				return
			}

			// instance exists
			ecs.Info("instance exists, trying to delete it")
			if err := c.DelInstance(ins.InstanceId); err != nil {
				ecs.Error("error deleting ecs instance: %v", err)
				continue
			}
			break
		}
	}

	for range ticker.C {
		if ins, err := acquireInstance(c, instanceName); err != nil {
			ecs.Error("error querying instances: %v", err)
			continue
		} else if ins == nil {
			ecs.Info("instance is deleted")
			return
		}
		pt.Info("instance is being deleted")
	}
}

func initEnv(ip, rootPwd string, cmds []string) error {
	ticker := time.NewTicker(loopInterval)
	pt := ecs.NewProgressTracker()
	var s *ecs.Ssh
	for range ticker.C {
		var err error
		if s, err = ecs.NewSsh(ip, rootPwd); err == nil {
			break
		}
		pt.Info("waiting instance to be ready")
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
	defer s.Close()
	defer func() { stopSignal <- true }()

	for _, cmd := range cmds {
		if err := s.Run(cmd); err != nil {
			return err
		}
	}

	return nil
}
