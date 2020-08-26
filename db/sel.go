package apexdb

import (
	//"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" //comment
)

//Selvars will return []string of mode requested
func Selvars(mode string) []string {
	qry := getqry(mode)
	res, err := db.Query(qry)
	handleError(err)
	var sl []string
	for res.Next() {
		var s string
		// for each row, scan the result into our tag composite object
		err := res.Scan(&s)
		handleError(err)

		sl = append(sl, s)
	}
	res.Close()
	return sl
}
func getqry(mode string) string {
	qry := fmt.Sprintf("SELECT descrip from vars WHERE cat='%s';", mode)
	return qry
}
