package proto

// User 用户信息
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// GetUserRequest 获取用户请求
type GetUserRequest struct {
	ID int64 `json:"id"`
}

// GetUserResponse 获取用户响应
type GetUserResponse struct {
	User *User `json:"user"`
}