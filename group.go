package netman

import (
	"errors"
	"sync"
)

type GroupPool struct {
	Pool map[string]*Group `json:"pool"`
	mux  *sync.RWMutex
}

type Group struct {
	Name     string    `json:"name"`      // 组名称，用于说明组性质
	Rule     string    `json:"rule"`      // 组规则，例如：192.168.16.0/22,192.168.32.0/22
	IpPool   *IpPool   `json:"ip_pool"`   // ip池
	UserPool *UserPool `json:"user_pool"` // 用户池
}

func newGroupPool() *GroupPool {
	return &GroupPool{
		make(map[string]*Group),
		new(sync.RWMutex),
	}
}

// 获取指定group
func (this *GroupPool) get(name string) (*Group, bool) {
	// get from map by key
	group, ok := this.Pool[name]
	return group, ok
}

// 增加一个group
func (this *GroupPool) add(name, rule string) (*Group, error) {
	// 先看有没有这个name
	_, exist := this.Pool[name]
	if exist {
		return nil, errors.New("group already exist")
	}
	// 新建一个Group，设置name和rule，Pool初始化为空
	// 新的组要解析rule，并检查是否已经被占用
	group := &Group{
		Name:     name,
		IpPool:   newIpPool(),
		UserPool: newUserPool(),
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	if err := this.checkRule(name, rule); err != nil {
		return nil, err
	}
	// 解析rule
	if err := group.IpPool.parseRule(rule); err != nil {
		return nil, err
	}
	group.Rule = rule
	// 把Group加到GroupPool中
	this.Pool[name] = group
	return group, nil
}

func (this *GroupPool) update(name, rule string) error {
	// 先看有没有这个group
	group, exist := this.Pool[name]
	if !exist {
		return errors.New("group not exist")
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	if err := this.checkRule(name, rule); err != nil {
		return err
	}
	// 解析rule
	if err := group.IpPool.parseRule(rule); err != nil {
		return err
	}
	group.Rule = rule
	return nil
}

func (this *GroupPool) del(name string) error {
	group, exist := this.Pool[name]
	if !exist {
		return errors.New("can't find group")
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	if len(group.UserPool.Pool) > 0 {
		return errors.New("can't delete nonempty group")
	}
	delete(this.Pool, name)
	return nil
}

func (this *GroupPool) updateUserGroup(userName, groupName, ip string) error {
	user, exist := this.getUser(userName)
	if !exist {
		return errors.New("can't find current user")
	}
	// 把一个用户从一个组改到另一个组，或者改ip
	currentGroup, exist := this.get(user.GroupName)
	if !exist {
		return errors.New("can't find current group")
	}
	if err := currentGroup.delUser(user); err != nil {
		return err
	}
	targetGroup, exist := this.get(groupName)
	if !exist {
		return errors.New("can't find target group")
	}
	if err := targetGroup.addUser(user, ip); err != nil {
		return err
	}
	return nil

}

func (this *GroupPool) getUser(name string) (*User, bool) {
	for _, group := range this.Pool {
		if user, exist := group.getUser(name); exist {
			return user, true
		}
	}
	return nil, false
}

func (this *GroupPool) addUser(userName, groupName, ip string) error {
	group, ok := this.get(groupName)
	if !ok {
		return errors.New("group not exist!")
	}
	if _, exist := this.getUser(userName); exist {
		return errors.New("user already exist")
	}
	user := newUser(userName)
	if err := group.addUser(user, ip); err != nil {
		return err
	}
	return nil
}

func (this *GroupPool) delUser(name string) error {
	user, exist := this.getUser(name)
	if !exist {
		return errors.New("user not exist")
	}
	group, exist := this.get(user.GroupName)
	if !exist {
		return errors.New("group not exist!")
	}
	if err := group.delUser(user); err != nil {
		return err
	}
	return nil
}

func (this *GroupPool) checkRule(name, rule string) error {
	ipPool := newIpPool()
	if err := ipPool.parseRule(rule); err != nil {
		return err
	}
	// 看是否和正在被使用的冲突
	for groupName, group := range this.Pool {
		if groupName != name && ipPool.conflict(group.IpPool) {
			return errors.New("conflict with exist ip pool")
		}
	}
	return nil
}

// 增加组成员
//     一个user只能属于一个group
//     ip可以指定，也可以自动分配
func (this *Group) addUser(user *User, ip string) error {
	if _, exist := this.UserPool.get(user.Name); exist {
		return errors.New("user already in this group")
	}
	userIp, err := this.IpPool.use(ip)
	if err != nil {
		return err
	}
	user.Ip = userIp
	user.GroupName = this.Name
	if err := this.UserPool.add(user); err != nil {
		return err
	}
	return nil
}

func (this *Group) getUser(name string) (*User, bool) {
	return this.UserPool.get(name)
}

// 删除用户
func (this *Group) delUser(user *User) error {
	if err := this.UserPool.del(user.Name); err != nil {
		return err
	}
	if err := this.IpPool.release(user.Ip); err != nil {
		return err
	}
	return nil
}
