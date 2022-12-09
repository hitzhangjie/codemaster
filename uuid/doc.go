// Package uuid 对需要生成唯一id的场景进行了测试，使用随机数？uuidv1？uuidv4？
//
// -----------------------------------------------------------------------
// 使用uuidv1作为唯一id，基本上能满足需要了。
//
// uuid v1使用如下几个部分来构建全局唯一id：
//
//   - nodeid: 通常为mac地址或者ip (更推荐的做法是ip);
//   - time: 本地时间;
//   - clock sequence: mac地址或者ip相同、时间相同（多进程同时生成，或请求量大，或时间被拨回，如时间设置错误或者ntp同步问题等），会导致碰撞问题;
//     clock sequence可以对抗这种碰撞情况，它通常被初始化为一个随机值.
//
// 可以使用uuid v1来生成全局唯一id，详见：https://www.rfc-editor.org/rfc/rfc4122.
package uuid
