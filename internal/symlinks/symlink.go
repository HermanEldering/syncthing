// Copyright (C) 2014 The Syncthing Authors.
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program. If not, see <http://www.gnu.org/licenses/>.

package symlinks

import (
	"os"
	"runtime"

	"github.com/syncthing/syncthing/internal/osutil"
	"github.com/syncthing/syncthing/internal/protocol"
)

var (
	Supported = true
)

func Read(path string) (string, uint32, error) {
	var mode uint32
	stat, err := os.Stat(path)
	if err != nil {
		mode = protocol.FlagSymlinkMissingTarget
	} else if stat.IsDir() {
		mode = protocol.FlagDirectory
	}
	path, err = os.Readlink(path)

	return osutil.NormalizedFilename(path), mode, err
}

func IsSymlink(path string) (bool, error) {
	lstat, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return lstat.Mode()&os.ModeSymlink != 0, nil
}

func Create(target, source string, flags uint32) error {
	return os.Symlink(osutil.NativeFilename(target), source)
}

func ChangeType(path string, flags uint32) error {
	if runtime.GOOS != "windows" {
		// This is a Windows-only concept.
		return nil
	}

	target, cflags, err := Read(path)
	if err != nil {
		return err
	}

	// If it's the same type, nothing to do.
	if cflags&protocol.SymlinkTypeMask == flags&protocol.SymlinkTypeMask {
		return nil
	}

	// If the actual type is unknown, but the new type is file, nothing to do
	if cflags&protocol.FlagSymlinkMissingTarget != 0 && flags&protocol.FlagDirectory == 0 {
		return nil
	}

	return osutil.InWritableDir(func(path string) error {
		// It should be a symlink as well hence no need to change permissions
		// on the file.
		os.Remove(path)
		return Create(target, path, flags)
	}, path)
}
