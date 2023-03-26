package telejoon

type UserStateRepository interface {
	Store(userID int64, state string) error
	Find(userID int64) (string, error)
}
