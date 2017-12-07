package netman

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
)

func TestInterface(t *testing.T) {
	groups := GetGroupList()
	// 正常添加组
	if err := AddGroup("group1", "192.168.16.0/22"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 正常修改组
	if err := UpdateGroupRule("group1", "192.168.32.0/22,192.168.16.0/22"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 这时候ip池应该是1022+1022
	group1, _ := groups.get("group1")
	if group1.IpPool.size() != 1022+1022 {
		t.Error("group size should be 2044 but", group1.IpPool.size())
		printGroup(t)
	}
	// 添加一个用户，自动分配ip
	if err := AddUser("user1", "group1", ""); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 这时候用户ip应该是192.168.32.1
	user1, _ := group1.getUser("user1")
	if !user1.Ip.Equal(net.ParseIP("192.168.32.1")) {
		t.Error("user ip should be 192.168.32.1 but", user1.Ip)
		printGroup(t)
	}
	// 把用户ip指定为192.168.33.0
	if err := UpdateUser("user1", "group1", "192.168.33.0"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 这时候用户ip应该是192.168.33.0
	if !user1.Ip.Equal(net.ParseIP("192.168.33.0")) {
		t.Error("user ip should be 192.168.33.0 but", user1.Ip)
		printGroup(t)
	}

	// 添加一个用户2，指定重复的ip，应该报错
	if err := AddUser("user2", "group1", "192.168.33.0"); err != nil {
		t.Log(err)
	} else {
		t.Error("conflict user ip should have err")
		printGroup(t)
	}
	// 这时候依然只有一个用户
	if len(group1.UserPool.Pool) != 1 {
		t.Error("group user pool size should be 1 but", len(group1.UserPool.Pool))
		printGroup(t)
	}
	// 修改组ip池，因为192.168.32.0/22网段有用户在使用，这里应该报错
	if err := UpdateGroupRule("group1", "192.168.16.0/22"); err != nil {
		t.Log(err)
	} else {
		t.Error("ip pool confilct should have err")
		printGroup(t)
	}
	// 组的rule还是原来的
	if group1.Rule != "192.168.32.0/22,192.168.16.0/22" {
		t.Error("group rule should not change")
		printGroup(t)
	}
	// 重复用户不能加
	if err := AddUser("user1", "group1", ""); err != nil {
		t.Log(err)
	} else {
		t.Error("same user name shuold have err")
		printGroup(t)
	}
	// 这时候依然只有一个用户
	if len(group1.UserPool.Pool) != 1 {
		t.Error("group user pool size should be 1 but", len(group1.UserPool.Pool))
		printGroup(t)
	}
	// 广播ip不能加
	if err := AddUser("user2", "group1", "192.168.31.255"); err != nil {
		t.Log(err)
	} else {
		t.Error("add broad cast ip should have err")
		printGroup(t)
	}
	// 这时候依然只有一个用户
	if len(group1.UserPool.Pool) != 1 {
		t.Error("group user pool size should be 1 but", len(group1.UserPool.Pool))
		printGroup(t)
	}
	// zero ip不能加
	if err := AddUser("user2", "group1", "192.168.32.0"); err != nil {
		t.Log(err)
	} else {
		t.Error("add zero ip should have err")
		printGroup(t)
	}
	// 这时候依然只有一个用户
	if len(group1.UserPool.Pool) != 1 {
		t.Error("group user pool size should be 1 but", len(group1.UserPool.Pool))
		printGroup(t)
	}
	// 正常添加一个用户2
	if err := AddUser("user2", "group1", "192.168.33.1"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 正常删除用户1
	if err := DelUser("user1"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 不能添加重复ip段的组
	if err := AddGroup("group2", "192.168.32.0/20"); err != nil {
		t.Log(err)
	} else {
		t.Error("confilct ip pool should have err")
		printGroup(t)
	}
	// 不能添加重名的组
	if err := AddGroup("group1", "192.168.64.0/22"); err != nil {
		t.Log(err)
	} else {
		t.Error("same group name should have err")
		printGroup(t)
	}
	if len(groups.Pool) != 1 {
		t.Error("group pool size should be 1 but", len(groups.Pool))
		printGroup(t)
	}
	// 正常添加组2
	if err := AddGroup("group2", "192.168.64.0/22"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 正常删除组2
	if err := DelGroup("group2"); err != nil {
		t.Error(err)
		printGroup(t)
	}
	// 不能删除组1
	if err := DelGroup("group1"); err != nil {
		t.Log(err)
	} else {
		t.Error("should have err here")
		printGroup(t)
	}
	printGroup(t)
	str := Serialize()
	t.Log(string(str))
	if err := Deserialize(str); err != nil {
		t.Error(err)
	}
	printGroup(t)
}

func printGroup(t *testing.T) {
	groups := GetGroupList()
	ret, _ := json.Marshal(groups)
	t.Log(string(ret))
}

func debug(msg ...interface{}) {
	fmt.Println("debug")
	fmt.Println(msg...)
}
