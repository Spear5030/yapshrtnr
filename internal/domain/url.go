package domain

type URL struct {
	Short string `db:"short"`
	Long  string `db:"long"`
	User  string `db:"userID"`
}
