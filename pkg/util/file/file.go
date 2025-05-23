// Copyright 2021 - 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"context"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
)

// ReadFile read all data from file
func ReadFile(fs fileservice.ReplaceableFileService, file string) ([]byte, error) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Minute, moerr.CauseReadFile)
	defer cancel()

	vec := &fileservice.IOVector{
		FilePath: file,
		Entries: []fileservice.IOEntry{
			{
				Offset: 0,
				Size:   -1,
			},
		},
	}
	if err := fs.Read(ctx, vec); err != nil {
		err = moerr.AttachCause(ctx, err)
		if moerr.IsMoErrCode(err, moerr.ErrFileNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return vec.Entries[0].Data, nil
}

// WriteFile write data to file
func WriteFile(fs fileservice.ReplaceableFileService, file string, data []byte) error {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Minute, moerr.CauseWriteFile)
	defer cancel()

	vec := fileservice.IOVector{
		FilePath: file,
		Entries: []fileservice.IOEntry{
			{
				Offset: 0,
				Size:   int64(len(data)),
				Data:   data,
			},
		},
	}
	err := fs.Replace(ctx, vec)
	return moerr.AttachCause(ctx, err)
}
