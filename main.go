package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
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
	fmt.Printf("SQL Server %s (%d.0)\n", getVersion(version), getMajorVersion(version))
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
