package netman

import (
	"encoding/json"
)

var (
	groupPool *GroupPool
)

func init() {
	groupPool = newGroupPool()
}

func Serialize() []byte {
	ret, _ := json.Marshal(groupPool)
	return ret
}

func Deserialize(s []byte) error {
	groupPool = newGroupPool()
	return json.Unmarshal(s, groupPool)
}

// 创建组
// @param groupName 组名称
// @param rule 组规则，例如：192.168.16.0/22,192.168.32.0/22
// 当组ip池与现有组的ip池冲突时，将会返回错误
// 当名称重复时，返回错误
func AddGroup(groupName, rule string) error {
	_, err := groupPool.add(groupName, rule)
	return err
}

// 删除组
// @param groupName 组名称
// 当名称对应的数据不存在时，返回错误
func DelGroup(groupName string) error {
	return groupPool.del(groupName)
}

// 修改组规则
// @param groupName 组名称
// @param rule 组规则，例如：192.168.16.0/22,192.168.32.0/22
// 当组ip池与现有组的ip池冲突时，将会返回错误
// 当名称对应的数据不存在时，返回错误
func UpdateGroupRule(groupName, rule string) error {
	return groupPool.update(groupName, rule)
}

// 获取组列表
func GetGroupList() *GroupPool {
	return groupPool
}

// 创建用户
// @param userName 用户名
// @param groupName 组名称
// @param ip 指定ip，例如：192.168.16.1
// 当指定ip字段不是合法ip时，将会自动分配一个可用ip
// 当指定ip已被使用或者不在组ip池内时，返回错误
// 当没有可用ip时，返回错误
// 当名称重复时，返回错误
// 当名称对应的数据不存在时，返回错误
func AddUser(userName, groupName, ip string) error {
	return groupPool.addUser(userName, groupName, ip)
}

// 改组或者ip
// @param userName 用户名
// @param groupName 组名称
// @param ip 指定ip，例如：192.168.16.1
// 当指定ip字段不是合法ip时，将会自动分配一个可用ip
// 当指定ip已被使用或者不在组ip池内时，返回错误
// 当没有可用ip时，返回错误
// 当名称对应的数据不存在时，返回错误
func UpdateUser(userName, groupName, ip string) error {
	return groupPool.updateUserGroup(userName, groupName, ip)
}

// 删除用户
// @param userName 用户名
// 当名称对应的数据不存在时，返回错误
func DelUser(userName string) error {
	return groupPool.delUser(userName)
}
