package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SECTOR_SIZE       = 256
	SECTORS_PER_TRACK = 13
)

var sectorBuffer [SECTOR_SIZE]byte

func readSector(file *os.File, track, sector int) []byte {
	offset := int64((track*SECTORS_PER_TRACK + sector) * SECTOR_SIZE)
	file.Seek(offset, 0)
	file.Read(sectorBuffer[:])
	return sectorBuffer[:]
}

func extractFile(file *os.File, track, sector int, outdir, filename string) {
	outfile, _ := os.Create(filepath.Join(outdir, filename+".bin"))
	defer outfile.Close()
	var dataBuffer [SECTOR_SIZE]byte
	for track != 0 {
		ts := readSector(file, track, sector)
		nextTrack := int(ts[1])
		nextSector := int(ts[2])
		for i := 0; i < 122; i += 2 {
			t := int(ts[12+i])
			s := int(ts[12+i+1])
			if t == 0 {
				break
			}
			offset := int64((t*SECTORS_PER_TRACK + s) * SECTOR_SIZE)
			file.Seek(offset, 0)
			file.Read(dataBuffer[:])
			outfile.Write(dataBuffer[:])
		}
		track, sector = nextTrack, nextSector
	}
}

func writeInformationFile(outdir string) {
	infoPath := filepath.Join(outdir, "information.doc")
	infoFile, _ := os.Create(infoPath)
	defer infoFile.Close()
	content := `Extracted the disk by "di3Doker"
The Author Account is "www.github.com/CoolyDucks"
The Project is Licensed by "LinkFurry v1.0"`
	infoFile.WriteString(content)
}

func d13forge(path string) {
	file, _ := os.Open(path)
	defer file.Close()
	name := filepath.Base(path[:len(path)-len(filepath.Ext(path))])
	outdir := name + "_extractor"
	os.Mkdir(outdir, 0755)
	vtoc := readSector(file, 17, 0)
	catalogTrack := int(vtoc[1])
	catalogSector := int(vtoc[2])
	for catalogTrack != 0 {
		sector := readSector(file, catalogTrack, catalogSector)
		nextTrack := int(sector[1])
		nextSector := int(sector[2])
		for i := 0; i < 7; i++ {
			entry := sector[11+i*35 : 11+i*35+35]
			ft := int(entry[0])
			fs := int(entry[1])
			if ft == 0 {
				continue
			}
			fn := string(entry[3:33])
			for j, r := range fn {
				if r == ' ' {
					fn = fn[:j] + "_" + fn[j+1:]
				}
			}
			fmt.Println("Extracting", fn)
			extractFile(file, ft, fs, outdir, fn)
		}
		catalogTrack, catalogSector = nextTrack, nextSector
	}
	writeInformationFile(outdir)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("D13docker> ")
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]
		if input == "exit" {
			break
		} else if input != "" {
			d13forge(input)
		}
	}
}
