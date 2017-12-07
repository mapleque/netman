package netman

import (
	"errors"
	"net"
	"strings"
	"sync"
)

type IpPool struct {
	mux  *sync.RWMutex
	Pool []*IpNet `json:"pool"`
}

type IpNet struct {
	rule  string
	ipNet *net.IPNet
	Using []net.IP `json:"using"`
}

func newIpPool() *IpPool {
	return &IpPool{
		mux:  new(sync.RWMutex),
		Pool: []*IpNet{},
	}
}

func (this *IpPool) size() int {
	sum := 0
	for _, subnet := range this.Pool {
		sum += subnet.size()
	}
	return sum
}

func (this *IpPool) conflict(pool *IpPool) bool {
	for _, oSubnet := range this.Pool {
		for _, tSubnet := range pool.Pool {
			if oSubnet.ipNet.Contains(tSubnet.ipNet.IP) || tSubnet.ipNet.Contains(oSubnet.ipNet.IP) {
				return true
			}
		}
	}
	return false
}

func (this *IpPool) use(ipName string) (net.IP, error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, subnet := range this.Pool {
		ip, ok := subnet.getAvailableIp(ipName)
		if ok {
			subnet.Using = append(subnet.Using, ip)
			return ip, nil
		}
	}
	return nil, errors.New("can not use this ip or no available ip to use")
}

func (this *IpPool) release(ip net.IP) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, subnet := range this.Pool {
		index := -1
		for i, curIp := range subnet.Using {
			if curIp.Equal(ip) {
				index = i
			}
		}
		if index >= 0 {
			subnet.Using = append(subnet.Using[:index], subnet.Using[index+1:]...)
			return nil
		}
	}
	return errors.New("ip is not using")
}

// rule: 192.168.16.0/22,192.168.32.0/22
func (this *IpPool) parseRule(rules string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	// 这里先准备一个新的ipNet，不实际去改数据，以方便中间发现错误直接退出
	prepareIpNet := []*IpNet{}
	for _, rule := range strings.Split(rules, ",") {
		_, ipNet, err := net.ParseCIDR(rule)
		if err != nil {
			return err
		}
		prepareIpNet = append(prepareIpNet, &IpNet{
			rule:  rule,
			ipNet: ipNet,
			Using: []net.IP{},
		})
	}
	if len(this.Pool) != 0 {
		// 已经有rule了，要把已有的都放到新的网段里边，如果有不涵盖的，就报错
		for _, subnet := range this.Pool {
			for _, ip := range subnet.Using {
				contain := false
				for _, tarNet := range prepareIpNet {
					if tarNet.ipNet.Contains(ip) {
						tarNet.Using = append(tarNet.Using, ip)
						contain = true
					}
				}
				if !contain {
					return errors.New("must contain ip in using")
				}
			}
		}
	}
	// **** 这里需要观察是不是有内存泄漏 ****//
	// 直接换指针，旧的就丢掉等回收了
	this.Pool = prepareIpNet
	return nil
}

func (this *IpNet) nextIp(ip net.IP) net.IP {
	ip = ip.To4()
	tarIp := net.IPv4(0, 0, 0, 0).To4()
	carry := true // 进位标记
	for i := len(tarIp) - 1; i >= 0; i-- {
		if !carry {
			tarIp[i] = ip[i]
		} else {
			tarIp[i] = ip[i] + 1
			if tarIp[i] != 0 {
				carry = false
				// 这是进位标记，如果当前段是0，就需要进位
			}
		}
	}
	return tarIp
}

func (this *IpNet) inUsing(ip net.IP) bool {
	for _, ipInUsing := range this.Using {
		if ip.Equal(ipInUsing) {
			return true
		}
	}
	return false
}

func (this *IpNet) getAvailableIp(ipName string) (net.IP, bool) {
	if tarIp := net.ParseIP(ipName); tarIp != nil {
		// 如果有指定ip，就尝试给分配这个ip
		if this.ipNet.Contains(tarIp) && !this.ipNet.IP.Equal(tarIp) && !this.inUsing(tarIp) {
			// 找到了，并且没被使用
			return tarIp, true
		}
		return nil, false
	}
	// 如果没有指定ip，就自动分配一个
	// 找一个有空闲网段找最小空闲ip分配
	if len(this.Using) < this.size() {
		// 注意这里要从1开始，0 ip不能用
		curIp := this.nextIp(this.ipNet.IP)
		for tarIp := curIp; tarIp != nil; tarIp = this.nextIp(tarIp) {
			if !this.inUsing(tarIp) {
				if this.ipNet.Contains(tarIp) && !this.ipNet.Contains(this.nextIp(tarIp)) {
					// 不能是本段最后一个ip
					return nil, false
				}
				return tarIp, true
			}
		}
	}
	return nil, false
}

func (this *IpNet) size() int {
	ones, bits := this.ipNet.Mask.Size()
	return 1<<uint(bits-ones) - 2
}
