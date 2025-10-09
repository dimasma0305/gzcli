package challenge

// AppSettings represents application settings
//
//nolint:revive // Field names match JSON structure
type AppSettings struct {
	ContainerProvider struct {
		PublicEntry string `json:"PublicEntry"`
	} `json:"ContainerProvider"`
	EmailConfig struct {
		UserName string `json:"UserName"`
		Password string `json:"Password"`
		Smtp     struct {
			Host string `json:"Host"`
			Port int    `json:"Port"`
		} `json:"Smtp"`
	} `json:"EmailConfig"`
}
