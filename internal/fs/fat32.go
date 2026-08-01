package fs

import (
	"encoding/binary"
	"fmt"
)

const (
	fat32SectorSize      = 512
	fat32ReservedSectors = 32
	fat32NumberOfFATs    = 2
	fat32RootCluster     = 2
	fat32FSInfoSector    = 1
	fat32BackupBoot      = 6
	fat32MinClusters     = 65525
)

// CreateEmptyFAT32Image creates a minimal FAT32 super-floppy image.
//
// The image starts directly with a FAT32 volume boot record. It does not
// contain an MBR or partition table.
func CreateEmptyFAT32Image(totalSectors uint32) []byte {
	const sectorsPerCluster uint32 = 1

	fatSectors, clusterCount := calculateFATSize(
		totalSectors,
		sectorsPerCluster,
	)

	if clusterCount < fat32MinClusters {
		panic(fmt.Sprintf(
			"FAT32 image is too small: %d clusters; need at least %d",
			clusterCount,
			fat32MinClusters,
		))
	}

	img := make([]byte, uint64(totalSectors)*fat32SectorSize)

	writeBootSector(
		img[0:fat32SectorSize],
		totalSectors,
		uint8(sectorsPerCluster),
		fatSectors,
	)

	// Backup boot sector, as declared in BPB_BkBootSec.
	backupBootOffset := fat32BackupBoot * fat32SectorSize
	copy(
		img[backupBootOffset:backupBootOffset+fat32SectorSize],
		img[0:fat32SectorSize],
	)

	writeFSInfo(
		img[fat32FSInfoSector*fat32SectorSize : (fat32FSInfoSector+1)*fat32SectorSize],
	)

	// A backup FSInfo sector normally follows the backup boot sector.
	backupFSInfoSector := fat32BackupBoot + fat32FSInfoSector
	copy(
		img[backupFSInfoSector*fat32SectorSize:(backupFSInfoSector+1)*fat32SectorSize],
		img[fat32FSInfoSector*fat32SectorSize:(fat32FSInfoSector+1)*fat32SectorSize],
	)

	fat1Sector := uint32(fat32ReservedSectors)
	fat2Sector := fat1Sector + fatSectors

	initializeFAT(img, fat1Sector)
	initializeFAT(img, fat2Sector)

	// The root directory is cluster 2. Its data is already zero-filled,
	// which represents an empty directory.
	return img
}

func calculateFATSize(
	totalSectors uint32,
	sectorsPerCluster uint32,
) (fatSectors uint32, clusterCount uint32) {
	fatSectors = 1

	for i := 0; i < 100; i++ {
		overhead := uint32(fat32ReservedSectors) +
			uint32(fat32NumberOfFATs)*fatSectors

		if totalSectors <= overhead {
			panic("FAT32 image is too small for filesystem metadata")
		}

		dataSectors := totalSectors - overhead
		clusterCount = dataSectors / sectorsPerCluster

		requiredEntries := clusterCount + 2
		requiredBytes := uint64(requiredEntries) * 4
		requiredSectors := uint32(
			(requiredBytes + fat32SectorSize - 1) /
				fat32SectorSize,
		)

		if requiredSectors <= fatSectors {
			return fatSectors, clusterCount
		}

		fatSectors = requiredSectors
	}

	panic("failed to calculate FAT32 FAT size")
}

func writeBootSector(
	sector []byte,
	totalSectors uint32,
	sectorsPerCluster uint8,
	fatSectors uint32,
) {
	copy(sector[0:3], []byte{0xEB, 0x58, 0x90})
	copy(sector[3:11], []byte("EMULATR"))

	binary.LittleEndian.PutUint16(sector[11:13], fat32SectorSize)
	sector[13] = sectorsPerCluster
	binary.LittleEndian.PutUint16(
		sector[14:16],
		fat32ReservedSectors,
	)
	sector[16] = fat32NumberOfFATs

	// FAT12/FAT16-only BPB fields must be zero for FAT32.
	binary.LittleEndian.PutUint16(sector[17:19], 0) // RootEntCnt
	binary.LittleEndian.PutUint16(sector[19:21], 0) // TotSec16
	sector[21] = 0xF8
	binary.LittleEndian.PutUint16(sector[22:24], 0) // FATSz16

	// Geometry values are informational for this emulated device.
	binary.LittleEndian.PutUint16(sector[24:26], 63)
	binary.LittleEndian.PutUint16(sector[26:28], 255)
	binary.LittleEndian.PutUint32(sector[28:32], 0) // Hidden sectors

	binary.LittleEndian.PutUint32(sector[32:36], totalSectors)
	binary.LittleEndian.PutUint32(sector[36:40], fatSectors)
	binary.LittleEndian.PutUint16(sector[40:42], 0) // ExtFlags
	binary.LittleEndian.PutUint16(sector[42:44], 0) // FSVer
	binary.LittleEndian.PutUint32(
		sector[44:48],
		fat32RootCluster,
	)
	binary.LittleEndian.PutUint16(
		sector[48:50],
		fat32FSInfoSector,
	)
	binary.LittleEndian.PutUint16(
		sector[50:52],
		fat32BackupBoot,
	)

	sector[64] = 0x80
	sector[65] = 0
	sector[66] = 0x29
	binary.LittleEndian.PutUint32(sector[67:71], 0xEFBEADDE)
	copy(sector[71:82], []byte("GO_SD_CARD "))
	copy(sector[82:90], []byte("FAT32   "))

	sector[510] = 0x55
	sector[511] = 0xAA
}

func writeFSInfo(sector []byte) {
	binary.LittleEndian.PutUint32(sector[0:4], 0x41615252)
	binary.LittleEndian.PutUint32(sector[484:488], 0x61417272)

	// Unknown free-cluster count.
	binary.LittleEndian.PutUint32(sector[488:492], 0xFFFFFFFF)

	// Cluster 2 belongs to the root directory, so cluster 3 is the first
	// possible free cluster.
	binary.LittleEndian.PutUint32(sector[492:496], 3)

	binary.LittleEndian.PutUint32(sector[508:512], 0xAA550000)
}

func initializeFAT(img []byte, startSector uint32) {
	offset := uint64(startSector) * fat32SectorSize

	binary.LittleEndian.PutUint32(
		img[offset:offset+4],
		0x0FFFFFF8,
	)
	binary.LittleEndian.PutUint32(
		img[offset+4:offset+8],
		0xFFFFFFFF,
	)
	binary.LittleEndian.PutUint32(
		img[offset+8:offset+12],
		0x0FFFFFFF,
	)
}
