package main

import (
	"os"
	"fmt"
	"encoding/binary"
	"flag"
	"bytes"
	"io"
)

var (
	BLOCK_SIZE = int64(0x200)
)

func main() {
	filename := flag.String("filename", "", "path to the backup to analyze")

	flag.Parse()

	if *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	version, err := getInternalVersionFromBackup(*filename)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Internal Version: %d\n", version)
	fmt.Printf("SQL Server %s\n", getVersion(version))
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

		if bytes.Equal(blockHeader, []byte("MSCI")){
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

	_, err = f.Seek(offset + 0x0AC, 0)
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

// Returns the user-readable version corresponding to an
// internal SQL Server Version
// Versions from: https://sqlserverbuilds.blogspot.de/2014/01/sql-server-internal-database-versions.html
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