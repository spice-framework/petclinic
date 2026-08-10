package mysql

type resumableMigration struct {
	version uint64
	name    string
	steps   []string
}
