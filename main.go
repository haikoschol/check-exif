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
	"fmt"
	"github.com/dsoprea/go-exif/v2"
	exifcommon "github.com/dsoprea/go-exif/v2/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	//cwd, err := os.Getwd()
	//if err != nil {
	//	log.Fatalf("os.Getcwd(): %v\n", err)
	//}
	//
	//entries, err := os.ReadDir(cwd)
	//if err != nil {
	//	log.Fatalf("os.ReadDir(): %v\n", err)
	//}
	//
	//for _, entry := range entries {
	//	if !isSupportedMedia(entry.Name()) {
	//		continue
	//	}
	//
	//	p := path.Join(cwd, entry.Name())
	//	mediaFile, err := os.Open(p)
	//	if err != nil {
	//		log.Printf("os.Open(%s): %v\n", p, err)
	//		continue
	//	}
	//	defer mediaFile.Close()
	//
	//	exifData, err := exif.Decode(mediaFile)
	//	if err != nil {
	//		log.Printf("exif.Decode(%s): %v\n", p, err)
	//		mediaFile.Close()
	//		continue
	//	}
	//
	//	lat, lon, err := exifData.LatLong()
	//	if err != nil {
	//		log.Printf("%s is missing GPS coordinates: %v\n", p, err)
	//	}
	//	_ = lat
	//	_ = lon
	//
	//	_, err = exifData.DateTime()
	//	if err != nil {
	//		log.Printf("%s is missing timestamp: %v\n", p, err)
	//	}
	//}

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
		if err = setExif(p); err != nil {
			log.Printf("setExif(): %v\n", err)
		}
	}

}

func isSupportedMedia(filename string) bool {
	filename = strings.ToLower(filename)
	return strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg")
}

func setExif(filepath string) error {
	t, err := time.Parse(time.RFC3339, "2023-03-17T12:00:00Z")
	if err != nil {
		return fmt.Errorf("time.Parse(): %w", err)
	}

	sl, err := parseJpeg(filepath)
	if err != nil {
		return fmt.Errorf("parseJpeg(): %w", err)
	}

	rootIb, err := getOrCreateRootIfdBuilder(sl)
	if err != nil {
		return fmt.Errorf("getOrCreateRootIfdBuilder(): %w", err)
	}

	if err = setDateTimeOriginal(sl, rootIb, t.UTC()); err != nil {
		return fmt.Errorf("setDateTimeOriginal(): %w", err)
	}

	if err = writeJpeg(sl, filepath); err != nil {
		return fmt.Errorf("writeJpeg(): %w", err)
	}

	rawExif, err := exif.SearchFileAndExtractExif(filepath)
	if err != nil {
		return fmt.Errorf("exif.SearchFileAndExtractExif(): %w", err)
	}

	latDegrees := exifcommon.Rational{Numerator: 8, Denominator: 1}
	latMinutes := exifcommon.Rational{Numerator: 28, Denominator: 1}
	latSeconds := exifcommon.Rational{Numerator: 44, Denominator: 1}

	lonDegrees := exifcommon.Rational{Numerator: 115, Denominator: 1}
	lonMinutes := exifcommon.Rational{Numerator: 26, Denominator: 1}
	lonSeconds := exifcommon.Rational{Numerator: 17, Denominator: 1}

	lat, err := exif.NewGpsDegreesFromRationals("S", []exifcommon.Rational{latDegrees, latMinutes, latSeconds})
	if err != nil {
		return fmt.Errorf("exif.NewGpsDegreesFromRationals(lat): %w", err)
	}

	lon, err := exif.NewGpsDegreesFromRationals("E", []exifcommon.Rational{lonDegrees, lonMinutes, lonSeconds})
	if err != nil {
		return fmt.Errorf("exif.NewGpsDegreesFromRationals(lon): %w", err)
	}

	if err = setLatLon(sl, rawExif, lat, lon); err != nil {
		return fmt.Errorf("setLatLon(): %w", err)
	}

	if err = writeJpeg(sl, filepath); err != nil {
		return fmt.Errorf("writeJpeg(): %w", err)
	}

	return nil
}

func parseJpeg(filepath string) (*jpegstructure.SegmentList, error) {
	jmp := jpegstructure.NewJpegMediaParser()

	intfc, err := jmp.ParseFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("ParseFile(): %w", err)
	}

	return intfc.(*jpegstructure.SegmentList), nil
}

func writeJpeg(sl *jpegstructure.SegmentList, filepath string) error {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("os.OpenFile(): %w", err)
	}
	defer f.Close()

	return sl.Write(f)
}

func getOrCreateRootIfdBuilder(sl *jpegstructure.SegmentList) (*exif.IfdBuilder, error) {
	rootIb, err := sl.ConstructExifBuilder()
	if err == nil {
		return rootIb, nil
	}

	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	if err := exif.LoadStandardTags(ti); err != nil {
		return nil, fmt.Errorf("exif.LoadStandardTags(): %w", err)
	}

	rootIb = exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)
	if err = rootIb.AddStandardWithName("ProcessingSoftware", "github.com/haikoschol/check-exif"); err != nil {
		return nil, fmt.Errorf("AddStandardWithName(ProcessingSoftware): %w", err)
	}

	return rootIb, nil
}

func setDateTimeOriginal(sl *jpegstructure.SegmentList, rootIb *exif.IfdBuilder, t time.Time) error {
	childIb, err := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0")
	if err != nil {
		return fmt.Errorf("GetOrCreateIbFromRootIb(): %w", err)
	}

	updatedTimestampPhrase := exif.ExifFullTimestampString(t)

	err = childIb.SetStandardWithName("DateTimeOriginal", updatedTimestampPhrase)
	if err != nil {
		return fmt.Errorf("SetStandardWithName(DateTimeOriginal): %w", err)
	}

	err = sl.SetExif(rootIb)
	if err != nil {
		return fmt.Errorf("SetExif(): %w", err)
	}
	return nil
}

func setLatLon(
	sl *jpegstructure.SegmentList,
	rawExif []byte,
	lat exif.GpsDegrees,
	lon exif.GpsDegrees,
) error {
	im := exif.NewIfdMapping()
	if err := exif.LoadStandardIfds(im); err != nil {
		return fmt.Errorf("exif.LoadStandardIfds(): %w", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		return fmt.Errorf("exif.Collect(): %w", err)
	}

	rootIfd := index.RootIfd
	rootIb := exif.NewIfdBuilderFromExistingChain(rootIfd)

	gpsIb := exif.NewIfdBuilder(im, ti, exifcommon.IfdGpsInfoStandardIfdIdentity, exifcommon.EncodeDefaultByteOrder)

	if err = rootIb.AddChildIb(gpsIb); err != nil {
		return fmt.Errorf("AddChildIb(): %w", err)
	}

	err = gpsIb.SetStandardWithName("GPSLatitude", lat.Raw())
	if err != nil {
		return fmt.Errorf("SetStandardWithName(GPSLatitude): %w", err)
	}

	err = gpsIb.SetStandardWithName("GPSLatitudeRef", string(lat.Orientation))
	if err != nil {
		return fmt.Errorf("SetStandardWithName(GPSLatitudeRef): %w", err)
	}

	err = gpsIb.SetStandardWithName("GPSLongitude", lon.Raw())
	if err != nil {
		return fmt.Errorf("SetStandardWithName(GPSLongitude): %w", err)
	}

	err = gpsIb.SetStandardWithName("GPSLongitudeRef", string(lon.Orientation))
	if err != nil {
		return fmt.Errorf("SetStandardWithName(GPSLongitudeRef): %w", err)
	}

	err = sl.SetExif(rootIb)
	if err != nil {
		return fmt.Errorf("SetExif(): %w", err)
	}
	return nil
}
