package gzapi

import "fmt"

// Team represents a team in the GZCTF platform
//
//nolint:revive // Field names match API responses
type Team struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Bio     string `json:"bio"`
	Locked  bool   `json:"locked"`
	Members []User `json:"members"`
	CS      *GZAPI `json:"-"`
}

// TeamForm represents the data required to create or update a team
type TeamForm struct {
	Bio  string `json:"bio"`
	Name string `json:"name"`
}

// Delete removes the team from the platform
func (t *Team) Delete() error {
	if err := t.CS.delete(fmt.Sprintf("/api/admin/teams/%d", t.Id), nil); err != nil {
		return err
	}
	return nil
}

// CreateTeam creates a new team on the platform
func (cs *GZAPI) CreateTeam(teamForm *TeamForm) error {
	if err := cs.post("/api/team", teamForm, nil); err != nil {
		return err
	}
	return nil
}

// GetTeams retrieves teams for the current user
func (cs *GZAPI) GetTeams() ([]*Team, error) {
	var team []*Team
	if err := cs.get("/api/team/", &team); err != nil {
		return nil, err
	}
	return team, nil
}

// Teams retrieves all teams from the platform with pagination support
func (cs *GZAPI) Teams() ([]*Team, error) {
	var teams struct {
		Data []*Team `json:"data"`
	}
	if err := cs.get("/api/admin/teams", &teams); err != nil {
		return nil, err
	}
	for t := range teams.Data {
		teams.Data[t].CS = cs
	}
	return teams.Data, nil
}
