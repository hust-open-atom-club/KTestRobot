package main

import(
	"log"
	"fmt"
	"reflect"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Bug struct {
	Kernel			string `db:"kernel"`
	BugInfo			string `db:"bug_info"`
	Type			string `db:"type"`
	CausedBy		string `db:"caused_by"`
	// CommitLink		string `db:"commit_link"`
	ReportedTime	string `db:"reported_time"`
	ReportedBy		string `db:"reported_by"`
	// Copilot      string `db:"copilot"`
	Status			string `db:"status"`
	// FixedTime    string `db:"fixed_time"`
	// FixedBy      string `db:"fixed_by"`
	Comment			string `db:"comment"`
}

type TBug struct {
	Title string
	PatchBugs  []*Bug
	OtherBugs []*Bug
}

type ToInsert struct {
	table string
	kernel string
	buginfo string
	btype string
	bhash string
	causedby string
	reportedby string
}

// var db *sql.DB

var (
    dbUser  string = "root"
    dbPasswd  string = "lsc20011130"
    dbipAddr string = "127.0.0.1"
    dbport      int    = 3306
    dbName    string = "buglist"
    // charset   string = "utf8"
)

func initMysql() (*sql.DB){
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local", 
			dbUser, dbPasswd, dbipAddr, dbport, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("initMysql: ", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Println("connect to database: ", err)
		return nil
	}
	return db
}

func ListBug(table string, db *sql.DB) []*Bug {
	rows, err := db.Query("SELECT * FROM " + table)
	if err != nil {
		log.Println("ListBug: ", err)
		return nil
	}

	bugs := DoRowsMapper(rows)
	// log.Println(bugs)
	// for _, v := range bugs {
	// 	log.Println(*v)
	// }
	return bugs
}

func DoRowsMapper(rows *sql.Rows) []*Bug {

	// get col
	columns, err := rows.Columns()
	if err != nil {
		log.Println("sql: ", err)
		return nil
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	var res []*Bug

	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic
		}

		rowMap := make(map[string]string)
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col != nil {
				value = string(col)
				rowMap[columns[i]] = value
			}
		}

		var bug Bug
		t := reflect.TypeOf(bug)
		v := reflect.ValueOf(&bug).Elem()
		for i := 0; i < t.NumField(); i++ {
			f := v.Field(i)
			fieldName := t.Field(i).Tag.Get("db")
			f.SetString(rowMap[fieldName])
		}
		res = append(res, &bug)
	}
	// log.Println(res)
	return res
}

func getBugList() ([]*Bug, []*Bug) {

	db := initMysql()
	if db != nil {
		patchbugs := ListBug("patchbugs", db)
		otherbugs := ListBug("otherbugs", db)

		defer db.Close()
		return patchbugs, otherbugs
	}
	return nil, nil
}

func InsertBugList(in ToInsert, list []string, btype string, table string) {
	in.btype = btype
	in.table = table
	if table != "patchbugs" {
		in.causedby = ""
	}
	for _, single := range list {
		indata := in
		bhash := BugHash(in.kernel + single)
		indata.bhash = bhash
		indata.buginfo = single
		InsertBug(indata)
	}
}

func InsertBug(in ToInsert) {
	db := initMysql()
	if db != nil {
		defer db.Close()
	}
	sqlStr := "insert into " + in.table
	sqlStr += "(hash, kernel, bug_info, type, caused_by, reported_by) values (?,?,?,?,?,?)"
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		log.Printf("InsertBug prepare failed, err:%v\n", err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(in.bhash, in.kernel, in.buginfo, in.btype, in.causedby, in.reportedby)
	if err != nil {
		log.Printf("InsertBug insert failed, err:%v\n", err)
		// log.Println(in.buginfo)
		return
	}
}