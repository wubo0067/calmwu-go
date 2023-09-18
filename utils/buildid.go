/*
 * @Author: CALM.WU
 * @Date: 2023-09-18 11:15:05
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-09-18 15:07:10
 */

package utils

import (
	"bytes"
	"debug/elf"
	"encoding/hex"

	"github.com/pkg/errors"
)

type BuildIDType int

const (
	__gnuBuildIDSection = ".note.gnu.build-id"
	__goBuildIDSection  = ".note.go.buildid"
)

const (
	BuildIDTypeUnknown BuildIDType = iota
	BuildIDTypeGNU
	BuildIDTypeGO
)

type BuildID struct {
	Type BuildIDType
	ID   string
}

var (
	ErrNoBuildIDSection = errors.New("build ID section not found")
)

func GetBuildID(f *elf.File) (*BuildID, error) {
	var (
		buildID = &BuildID{
			Type: BuildIDTypeUnknown,
		}
		data []byte
		err  error
	)

	buildIDSection := f.Section(__gnuBuildIDSection)
	if buildIDSection == nil {
		buildIDSection = f.Section(__goBuildIDSection)
		if buildIDSection == nil {
			return nil, ErrNoBuildIDSection
		} else {
			buildID.Type = BuildIDTypeGO
		}
	} else {
		buildID.Type = BuildIDTypeGNU
	}

	switch buildID.Type {
	case BuildIDTypeGNU:
		data, err = buildIDSection.Data()
		if err != nil {
			return buildID, errors.Wrapf(err, "read %s.", __gnuBuildIDSection)
		}
		if len(data) < 16 {
			return buildID, errors.Wrapf(err, "%s is too small", __gnuBuildIDSection)
		}
		if !bytes.Equal([]byte("GNU"), data[12:15]) {
			return buildID, errors.Wrapf(err, "%s is not a GNU build-id", __gnuBuildIDSection)
		}
		rawBuildID := data[16:]
		if len(rawBuildID) != 20 && len(rawBuildID) != 8 { // 8 is xxhash, for example in Container-Optimized OS
			return buildID, errors.Wrapf(err, "%s has wrong size", __gnuBuildIDSection)
		}
		buildID.ID = hex.EncodeToString(rawBuildID)
	case BuildIDTypeGO:
		data, err = buildIDSection.Data()
		if err != nil {
			return buildID, errors.Wrapf(err, "read %s.", __goBuildIDSection)
		}
		if len(data) < 17 {
			return buildID, errors.Wrapf(err, "%s is too small", __goBuildIDSection)
		}
		data = data[16 : len(data)-1]
		if len(data) < 40 || bytes.Count(data, []byte("/")) < 2 {
			return buildID, errors.Wrapf(err, "wrong %s", __goBuildIDSection)
		}
		buildID.ID = Bytes2String(data)
		if buildID.ID == "redacted" {
			return buildID, errors.Wrapf(err, "blacklisted %s", __goBuildIDSection)
		}
	}

	return buildID, nil
}
