package telejoon

type UserStateRepository interface {
	Find(userID int64) (string, error)
	Store(userID int64, state string) error
	Update(userID int64, state string) error
}
