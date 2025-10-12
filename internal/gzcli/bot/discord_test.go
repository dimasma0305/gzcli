package bot

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: &Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "gzctf",
				WebhookURL: "https://discord.com/api/webhooks/123/abc",
				IconURL:    "https://example.com/icon.png",
			},
			wantErr: false,
		},
		{
			name: "valid config with minimal fields",
			config: &Config{
				WebhookURL: "https://discord.com/api/webhooks/123/abc",
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing webhook URL",
			config: &Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "gzctf",
			},
			wantErr: true,
		},
		{
			name:    "invalid config - nil config",
			config:  &Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && bot == nil {
				t.Error("New() returned nil bot without error")
			}
			if !tt.wantErr {
				// Verify defaults were set
				if bot.config.DBHost == "" {
					t.Error("DBHost not set to default")
				}
				if bot.config.DBPort == 0 {
					t.Error("DBPort not set to default")
				}
				if bot.config.DBUser == "" {
					t.Error("DBUser not set to default")
				}
				if bot.config.DBName == "" {
					t.Error("DBName not set to default")
				}
			}
		})
	}
}

func TestSanitizeTeamName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal team name",
			input: "TeamA",
			want:  "TeamA",
		},
		{
			name:  "team name with special chars",
			input: "Team@A!",
			want:  "Team@A!",
		},
		{
			name:  "team name with invalid unicode",
			input: "Team\u200BAI",
			want:  "TeamAI",
		},
		{
			name:  "team name with emojis",
			input: "Team🚀AI",
			want:  "TeamAI",
		},
		{
			name:  "allowed special characters",
			input: "Team!@#$%^&*()_+-={}[]:\";'<>,.?/\\",
			want:  "Team!@#$%^&*()_+-={}[]:\";'<>,.?/\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeTeamName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeTeamName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeNoticeValues(t *testing.T) {
	tests := []struct {
		name   string
		notice *GameNotice
		want   []string
	}{
		{
			name: "first blood with @everyone",
			notice: &GameNotice{
				NoticeType: 1,
				Values:     []string{"Team@everyone", "Challenge1"},
				GameTitle:  "Test CTF",
			},
			want: []string{"Team@everyon3", "Challenge1"},
		},
		{
			name: "second blood with @here",
			notice: &GameNotice{
				NoticeType: 2,
				Values:     []string{"@here Team", "Challenge2"},
				GameTitle:  "Test CTF",
			},
			want: []string{"@her3Team", "Challenge2"}, // Space removed by sanitizeTeamName
		},
		{
			name: "third blood with unicode",
			notice: &GameNotice{
				NoticeType: 3,
				Values:     []string{"Team\u200BAI", "Challenge3"},
				GameTitle:  "Test CTF",
			},
			want: []string{"TeamAI", "Challenge3"},
		},
		{
			name: "new hint (no team sanitization)",
			notice: &GameNotice{
				NoticeType: 4,
				Values:     []string{"Challenge@everyone"},
				GameTitle:  "Test CTF",
			},
			want: []string{"Challenge@everyon3"},
		},
		{
			name: "new challenge",
			notice: &GameNotice{
				NoticeType: 5,
				Values:     []string{"New@here Challenge"},
				GameTitle:  "Test CTF",
			},
			want: []string{"New@her3 Challenge"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizeNoticeValues(tt.notice)
			for i, got := range tt.notice.Values {
				if got != tt.want[i] {
					t.Errorf("sanitizeNoticeValues() value[%d] = %q, want %q", i, got, tt.want[i])
				}
			}
		})
	}
}

func TestCreateEmbed(t *testing.T) {
	config := &Config{
		WebhookURL: "https://discord.com/api/webhooks/123/abc",
	}
	bot, _ := New(config)

	tests := []struct {
		name       string
		notice     *GameNotice
		wantNil    bool
		wantColor  int
		wantFields int
	}{
		{
			name: "first blood",
			notice: &GameNotice{
				NoticeType: 1,
				Values:     []string{"TeamA", "Web Challenge"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil:    false,
			wantColor:  0xE74C3C,
			wantFields: 2, // Challenge + Event
		},
		{
			name: "second blood",
			notice: &GameNotice{
				NoticeType: 2,
				Values:     []string{"TeamB", "Crypto Challenge"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil:    false,
			wantColor:  0xF1C40F,
			wantFields: 2, // Challenge + Event
		},
		{
			name: "third blood",
			notice: &GameNotice{
				NoticeType: 3,
				Values:     []string{"TeamC", "Pwn Challenge"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil:    false,
			wantColor:  0x2ECC71,
			wantFields: 2, // Challenge + Event
		},
		{
			name: "new hint",
			notice: &GameNotice{
				NoticeType: 4,
				Values:     []string{"Hard Challenge"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil:    false,
			wantColor:  0x3498DB,
			wantFields: 2, // Challenge + Event
		},
		{
			name: "new challenge",
			notice: &GameNotice{
				NoticeType: 5,
				Values:     []string{"Another Challenge"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil:    false,
			wantColor:  0x9B59B6,
			wantFields: 2, // Challenge + Event
		},
		{
			name: "unknown notice type",
			notice: &GameNotice{
				NoticeType: 99,
				Values:     []string{"Something"},
				GameTitle:  "Test CTF 2024",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embed := bot.createEmbed(tt.notice)
			if (embed == nil) != tt.wantNil {
				t.Errorf("createEmbed() nil = %v, wantNil %v", embed == nil, tt.wantNil)
				return
			}
			if !tt.wantNil {
				if embed.Color != tt.wantColor {
					t.Errorf("createEmbed() color = %v, want %v", embed.Color, tt.wantColor)
				}
				if len(embed.Fields) != tt.wantFields {
					t.Errorf("createEmbed() fields = %v, want %v", len(embed.Fields), tt.wantFields)
				}
			}
		})
	}
}
