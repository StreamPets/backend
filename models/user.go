package models

type TwitchID string

type User struct {
	UserID   TwitchID `gorm:"primaryKey"`
	Username string
}

type ChatMessageEvent struct {
	User User
	Text string
}

type SubEvent struct {
	User     User
	IsSubbed bool
}

type CMD int

const (
	CMD_JUMP = CMD(iota)
	CMD_COLOR
)

type CommandEvent struct {
	User    User
	Command CMD
	Args    []string
}
