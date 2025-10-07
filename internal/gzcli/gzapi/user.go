package gzapi

import "fmt"

// User represents a user in the GZCTF platform
//
//nolint:revive // Field names match API responses
type User struct {
	Id       string `json:"id"`
	UserName string `json:"username"`
	Bio      string `json:"bio"`
	Captain  bool   `json:"captain"`
	API      *GZAPI `json:"-"`
}

// Delete removes the user from the platform
func (user *User) Delete() error {
	if err := user.API.delete(fmt.Sprintf("/api/admin/users/%s", user.Id), nil); err != nil {
		return err
	}
	return nil
}

// Users retrieves all users from the platform (admin only)
func (api *GZAPI) Users() ([]*User, error) {
	var users struct {
		Data []*User `json:"data"`
	}
	if err := api.get("/api/admin/users", &users); err != nil {
		return nil, err
	}
	for t := range users.Data {
		users.Data[t].API = api
	}
	return users.Data, nil
}

// JoinGame allows a user/team to join a game
//
//nolint:revive // gameId parameter name matches API specification
func (api *GZAPI) JoinGame(gameId int, joinModel *GameJoinModel) error {
	return api.post(fmt.Sprintf("/api/game/%d", gameId), joinModel, nil)
}
