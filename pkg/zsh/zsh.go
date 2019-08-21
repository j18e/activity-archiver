package zsh

type Command struct {
	Time    int64  `json:"timestamp"`
	Command string `json:"command"`
}

type Service interface {
	LastCMD() (Command, error)
	WriteCMDs([]Command) error
}
