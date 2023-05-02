// Copyright (C) 2023 Haiko Schol
// SPDX-License-Identifier: GPL-3.0-or-later

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	"path"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getcwd(): %v\n", err)
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatalf("os.ReadDir(): %v\n", err)
	}

	for _, entry := range entries {
		if !isSupportedMedia(entry.Name()) {
			continue
		}

		p := path.Join(cwd, entry.Name())
		mediaFile, err := os.Open(p)
		if err != nil {
			log.Printf("os.Open(%s): %v\n", p, err)
			continue
		}
		defer mediaFile.Close()

		exifData, err := exif.Decode(mediaFile)
		if err != nil {
			log.Printf("exif.Decode(%s): %v\n", p, err)
			mediaFile.Close()
			continue
		}

		lat, lon, err := exifData.LatLong()
		if err != nil {
			log.Printf("%s is missing GPS coordinates: %v\n", p, err)
		}
		_ = lat
		_ = lon

		_, err = exifData.DateTime()
		if err != nil {
			log.Printf("%s is missing timestamp: %v\n", p, err)
		}
	}
}

func isSupportedMedia(filename string) bool {
	filename = strings.ToLower(filename)
	return strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg")
}
