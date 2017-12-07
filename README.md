NETWORK MANAGER
========

一个用于管理网络的库。

入口文件[](interface.go)

划分子网
====
`AddGroup`接口可以添加子网，通过`rule`参数指定子网网段。

支持修改`UpdateGroupRule`和删除`DelGroup`。

分配ip
====
`AddUser`接口可以将用户添加到子网，如果`ip`参数是一个合法的IP，将尝试给用户分配该IP。

支持修改`UpdateUser`和删除`DelUser`。

数据
====
`GetGroupList`接口可以获得当前系统数据。

也可以通过序列化`Serialize`和反序列化`Deserialize`接口对数据进行持久化管理。

数据结构如下：
```
{
    pool:{
        <group_name>:{
            name:<group_name>,
            rule:<rule>,
            ip_pool:{
                pool:[{
                    using:[<ip>,...]
                },...]
            },
            user_pool:{
                pool:{
                    <user_name>:{
                        name:<user_name>,
                        ip:<ip>,
                        group_name:<group_name>
                    },...
                }
            }
        },...
    }
}
```
