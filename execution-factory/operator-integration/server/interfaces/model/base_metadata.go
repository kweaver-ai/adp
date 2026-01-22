package model

// import "time"

// // BaseMetadata 函数元数据基础字段
// type BaseMetadata struct {
// 	ID         int64  `json:"f_id" db:"f_id"`                   // 函数元数据ID
// 	Version    string `json:"f_version" db:"f_version"`         // 函数版本
// 	CreateUser string `json:"f_create_user" db:"f_create_user"` // 创建用户
// 	CreateTime int64  `json:"f_create_time" db:"f_create_time"` // 创建时间
// 	UpdateUser string `json:"f_update_user" db:"f_update_user"` // 更新用户
// 	UpdateTime int64  `json:"f_update_time" db:"f_update_time"` // 更新时间
// }

// // GetID 获取函数元数据ID
// func (b *BaseMetadata) GetID() int64 {
// 	return b.ID
// }

// // GetVersion 获取函数版本
// func (b *BaseMetadata) GetVersion() string {
// 	return b.Version
// }

// // SetCreateInfo 设置创建信息
// func (b *BaseMetadata) SetCreateInfo(user string) {
// 	b.CreateUser = user
// 	b.CreateTime = time.Now().UnixNano()
// }

// // SetUpdateInfo 设置更新信息
// func (b *BaseMetadata) SetUpdateInfo(user string) {
// 	b.UpdateUser = user
// 	b.UpdateTime = time.Now().UnixNano()
// }

// func (f *BaseMetadata) SetVersion(version string) {
// 	f.Version = version
// }
