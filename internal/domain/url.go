// Пакет domain содержит структуры относящиеся к бизнес логике.
package domain

// URL структура описывающая ссылку.
type URL struct {
	Short string `db:"short"`
	Long  string `db:"long"`
	User  string `db:"userID"`
}
