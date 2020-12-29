协议说明，业务服务通知到推荐系统
=====================================

> 1. 用户登录
```go
type ProtoDoyoRecUserLogin struct {
	UserID       string `json:"UserID"`
	UserLanguage string `json:"UserLanguage"` // http://www.lingoes.cn/zh/translator/langcode.htm
	UserCountry  string `json:"UserCountry"`  // https://zh.wikipedia.org/wiki/ISO_3166-1
	UserGender   int    `json:"UserGender"`   // 性别
}
```

> 2. 用户下线，前端业务层通过IM回调获得下线通知
```go
type ProtoDoyoRecUserLogout struct {
	UserID       string `json:"UserID"`
}
```

> 3. 主播开播
```go
type ProtoDoyoRecAnchorStartRoom struct {
	AnchorID        string `json:"AnchorID"`
	RoomID          string `json:"RoomID"`
	AnchorLanguage  string `json:"AnchorLanguage"` // 主播语言
	AnchorCountry   string `json:"AnchorCountry"`  // 主播国家
	RoomTitle       string `json:"RoomTitle"`
	DefinitionLabel string `json:"DefinitionLabel"`
	GameName        string `json:"GameName"`
}
```

> 4. 主播关播
```go
type ProtoDoyoRecAnchorStopRoom struct {
	AnchorID string `json:"AnchorID"`
}
```

> 5. 进入房间
```go
type ProtoDoyoRecUserEnterRoom struct {
	UserID string `json:"UserID"`
	RoomID string `json:"RoomID"`
}
```

> 6. 离开房间
```go
type ProtoDoyoRecUserLeaveRoom ProtoDoyoRecUserEnterRoom
```

> 7. 添加好友
```go
type ProtoDoyoRecAddFriends struct {
	UserID    string   `json:"UserID"`
	FriendLst []string `json:"FriendLst"`
}
```

> 8. 用户添加关注
```go
type ProtoDoyoRecAddFollow struct {
	UserID   string `json:"UserID"`
	FollowID string `json:"FollowID"`
}
```

> 9. 用户取消关注
```go
type ProtoDoyoRecDelFollow struct {
	UserID   string `json:"UserID"`
	FollowID string `json:"FollowID"`
}
```