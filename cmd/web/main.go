package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"snip/db/mysql"
)

type application struct {
	errorLog *log.Logger
	infoLog *log.Logger
	mysql mysql.Mysql
}

func main() {

	m, err := mysql.New()
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()

	//res, err := mysql.GetSnip()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer res.Close()
	//
	//mysql.Insert("Title","Data","")


	addr := flag.String("addr", ":4000", "Сетевой адрес HTTP")
	dsn := flag.String("dsn", "web:12345@/snippetbox?parseTime=true", "snippetbox.sql")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate | log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate | log.Ltime | log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog: infoLog,
		mysql: *m,
	}

	srv := &http.Server{
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
	}

	infoLog.Printf("Запуск сервера на %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}