package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"

	"net/http"
	"net/url"

	"database/sql"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

var (
	BLOCK_SIZE = int64(0x200)
)

func main() {
	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}

	for i := range cfg.Servers {
		server := &cfg.Servers[i]
		getServerVersion(cfg, server)
	}

	var srv service
	srv.Configuration = cfg
	http.ListenAndServe(":2408", srv)

	/*filename := flag.String("filename", "", "path to the backup to analyze")
	flag.Parse()

	if *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	version, err := getInternalVersionFromBackup(*filename)
	if err != nil {
		panic(err)
	}

	fmt.Println("File Information:")
	fmt.Printf("Internal Version: %d\n", version)
	fmt.Printf("SQL Server %s (%d.0)\n", getVersion(version), getMajorVersion(version))
	fmt.Println("")

	serverIndex, err := prompt(cfg.Servers, "Server: ")

	server := cfg.Servers[serverIndex]
	db, err := getConnection(server)
	if err != nil {
		panic(err)
	}

	if db == nil {
		fmt.Println("Invalid choice")
	}

	fmt.Println(server.VersionDescription)

	databases, err := getDatabases(db)
	if err != nil {
		panic(err)
	}

	databaseIndex, err := prompt(databases, "Database: ")
	database := databases[databaseIndex]
	fmt.Println(database)*/
}

/*
func prompt(values selectable, prompt string) (int, error) {
	for i := 0; i < values.getLength(); i++ {
		val := values.getElement(i)
		fmt.Printf("%d: %s\n", i, val)
	}

	var serverIndex int
	fmt.Printf(prompt)
	_, err := fmt.Scanf("%d\n", &serverIndex)
	if err != nil {
		return -1, err
	}

	return serverIndex, nil
}*/

func getServerVersion(cfg *configuration, server *server) {
	db, err := getConnection(*server)
	if err != nil {
		log.Fatal("Connection failed:", err.Error())
	}

	stmt, err := db.Prepare("select @@VERSION, SERVERPROPERTY('ProductLevel'), SERVERPROPERTY('Edition'), SERVERPROPERTY('ProductVersion')")
	if err != nil {
		log.Fatal("Prepare failed:", err.Error())
	}
	defer stmt.Close()

	row := stmt.QueryRow()
	var version string
	var productlevel string
	var edition string
	var productversion string
	err = row.Scan(&version, &productlevel, &edition, &productversion)
	if err != nil {
		log.Fatal("Scan failed:", err.Error())
	}

	server.Edition = edition
	server.VersionDescription = version
	server.Version = productversion
	server.Level = productlevel
}

func getDatabases(db *sql.DB) ([]string, error) {
	dbList, err := db.Prepare("SELECT name FROM master.sys.databases")
	if err != nil {
		return nil, err
	}
	defer dbList.Close()

	var databases []string
	rows, err := dbList.Query()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		databases = append(databases, name)
	}

	sort.Strings([]string(databases))
	return databases, nil
}

func getConnection(server server) (*sql.DB, error) {
	query := url.Values{}
	query.Add("app name", "SQL Backup")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(server.User, server.Password),
		Host:     server.Host,
		Path:     server.Instance,
		RawQuery: query.Encode(),
	}
	return sql.Open("sqlserver", u.String())
}

func findMSCIBlock(file io.ReadSeeker) (int64, error) {
	offset := int64(0)
	blockHeader := make([]byte, 4)
	for i := 0; i < 100; i++ {
		offset = offset + BLOCK_SIZE
		_, err := file.Seek(offset, 0)
		if err != nil {
			return 0, err
		}

		_, err = file.Read(blockHeader)
		if err != nil {
			return 0, err
		}

		if bytes.Equal(blockHeader, []byte("MSCI")) {
			return offset, nil
		}
	}

	return 0, fmt.Errorf("Could not find MSCI-Block")
}

func getInternalVersionFromBackup(filename string) (uint16, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	offset, err := findMSCIBlock(f)
	if err != nil {
		return 0, err
	}

	_, err = f.Seek(offset+0x0AC, 0)
	if err != nil {
		return 0, err
	}

	intVersion := make([]byte, 2)
	c, err := f.Read(intVersion)
	if err != nil {
		return 0, err
	}

	if c < 2 {
		return 0, fmt.Errorf("Error reading two bytes")
	}

	version := binary.LittleEndian.Uint16(intVersion)
	return version, nil
}

// Returns the Major Version corresponding to an internal SQL Server Version
// Versions from: https://sqlserverbuilds.blogspot.com/
func getMajorVersion(version uint16) uint16 {
	switch version {
	case 869:
		return 14
	case 852:
		return 13
	case 782:
		return 12
	case 706:
		fallthrough
	case 684:
		return 11
	case 660:
		fallthrough
	case 661:
		fallthrough
	case 655:
		return 10
	case 612:
		fallthrough
	case 611:
		return 9
	case 539:
		return 8
	case 515:
		return 7
	case 408:
		return 6
	}

	if version < 408 {
		return 5
	}

	if version > 869 {
		return 15
	}

	return 0
}

// Returns the user-readable version corresponding to an internal SQL Server Version
// Versions from: https://sqlserverbuilds.blogspot.com/
func getVersion(version uint16) string {
	switch version {
	case 869:
		return "2017"
	case 852:
		return "2016"
	case 782:
		return "2014"
	case 706:
		return "2012"
	case 684:
		return "2012 CTP1"
	case 660:
		fallthrough
	case 661:
		return "2008 R2"
	case 655:
		return "2008"
	case 612:
		return "2005 SP2+"
	case 611:
		return "2005"
	case 539:
		return "2000"
	case 515:
		return "7.0"
	case 408:
		return "6.5"
	}

	if version < 408 {
		return "Version before 6.5"
	}

	if version > 869 {
		return "Version after 2017"
	}

	return "Unrecognized version"
}
