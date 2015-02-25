/*
* Archon Login Server
* Copyright (C) 2014 Andrew Rodman
*
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this program.  If not, see <http://www.gnu.org/licenses/>.
* ---------------------------------------------------------------------
*
* CHARACTER server logic.
 */
package server

import (
	"database/sql"
	"fmt"
	"io"
	"libarchon/util"
	"os"
	"sync"
)

var baseKeyConfig = [420]byte{
	0x00, 0x00, 0x00, 0x00, 0x26, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x13, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x50, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x59, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x5e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5d, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5f, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5d, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x5c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5f, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x56, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5e, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x43, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x45, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x46, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x47, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x4a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x4b, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x2b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x2d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2e, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x30, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x32, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0xff, 0xff,
	0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x08, 0x00,
	0x01, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00,
	0x00, 0x02, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
	0x01, 0x00, 0x00, 0x00,
}

var baseSymbolChats = [1248]byte{
	0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00, 0x48, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00,
	0x6f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x0d, 0x00, 0xff, 0xff, 0xff, 0xff, 0x05, 0x18, 0x1d, 0x00, 0x05, 0x28, 0x1d, 0x01,
	0x36, 0x20, 0x2a, 0x00, 0x3c, 0x00, 0x32, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00,
	0x47, 0x00, 0x6f, 0x00, 0x6f, 0x00, 0x64, 0x00, 0x2d, 0x00, 0x62, 0x00, 0x79, 0x00, 0x65, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x74, 0x00, 0x00, 0x00, 0x76, 0x04, 0x0c, 0x00, 0xff, 0xff, 0xff, 0xff,
	0x06, 0x15, 0x14, 0x00, 0x06, 0x2b, 0x14, 0x01, 0x05, 0x18, 0x1f, 0x00, 0x05, 0x28, 0x1f, 0x01,
	0x36, 0x20, 0x2a, 0x00, 0x3c, 0x00, 0x32, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00, 0x48, 0x00, 0x75, 0x00, 0x72, 0x00, 0x72, 0x00,
	0x61, 0x00, 0x68, 0x00, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00,
	0x62, 0x03, 0x62, 0x03, 0xff, 0xff, 0xff, 0xff, 0x09, 0x16, 0x1b, 0x00, 0x09, 0x2b, 0x1b, 0x01,
	0x37, 0x20, 0x2c, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00,
	0x43, 0x00, 0x72, 0x00, 0x79, 0x00, 0x69, 0x00, 0x6e, 0x00, 0x67, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x74, 0x00, 0x00, 0x00, 0x4f, 0x07, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0x06, 0x15, 0x14, 0x00, 0x06, 0x2b, 0x14, 0x01, 0x05, 0x18, 0x1f, 0x00, 0x05, 0x28, 0x1f, 0x01,
	0x21, 0x20, 0x2e, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00, 0x49, 0x00, 0x27, 0x00, 0x6d, 0x00, 0x20, 0x00,
	0x61, 0x00, 0x6e, 0x00, 0x67, 0x00, 0x72, 0x00, 0x79, 0x00, 0x21, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5c, 0x00, 0x00, 0x00,
	0x16, 0x01, 0x01, 0x00, 0xff, 0xff, 0xff, 0xff, 0x0b, 0x18, 0x1b, 0x01, 0x0b, 0x28, 0x1b, 0x00,
	0x33, 0x20, 0x2a, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x09, 0x00, 0x45, 0x00,
	0x48, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x70, 0x00, 0x20, 0x00, 0x6d, 0x00, 0x65, 0x00, 0x21, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xec, 0x00, 0x00, 0x00, 0x5e, 0x06, 0x38, 0x01, 0xff, 0xff, 0xff, 0xff,
	0x02, 0x17, 0x1b, 0x01, 0x02, 0x2a, 0x1b, 0x00, 0x31, 0x20, 0x2c, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x02,
	0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02, 0xff, 0x00, 0x00, 0x02,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
	0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
}

// Struct used by Character Info packet.
type CharacterData struct {
	Experience     uint32
	Level          uint32
	GuildcardStr   [16]byte
	Unknown        [2]uint32
	NameColor      uint32
	Model          uint8
	Padding        [15]byte
	NameColorChksm uint32
	SectionId      byte
	Class          byte
	V2flags        byte
	Version        byte
	V1Flags        uint32
	Costume        uint16
	Skin           uint16
	Head           uint16
	HairRed        uint16
	HairGreen      uint16
	HairBlue       uint16
	PropX          float32
	PropY          float32
	Name           [16]uint16
	Playtime       uint32
}

var charConnections *util.ConnectionList = util.NewClientList()

func handleCharLogin(client *LoginClient) error {
	_, err := VerifyAccount(client)
	if err != nil {
		LogMsg(err.Error(), LogTypeInfo, LogPriorityLow)
		return err
	}

	SendSecurity(client, BBLoginErrorNone, client.guildcard)
	return nil
}

// Handle the options request - load key config and other option data from the datebase
// or provide defaults for new accounts.
func handleKeyConfig(client *LoginClient) error {
	optionData := make([]byte, 420)
	archondb := GetConfig().Database()

	row := archondb.QueryRow(
		"SELECT key_config from player_options where guildcard = ?", client.guildcard)
	err := row.Scan(&optionData)
	if err == sql.ErrNoRows {
		// We don't have any saved key config - give them the defaults.
		copy(optionData[:420], baseKeyConfig[:])
		_, err = archondb.Exec("INSERT INTO player_options "+
			"(guildcard, key_config) VALUES (?, ?)", client.guildcard, optionData[:420])
	}
	if err != nil {
		errMsg := fmt.Sprintf("SQL Error: %s", err.Error())
		LogMsg(errMsg, LogTypeError, LogPriorityCritical)
		return &util.ServerError{Message: errMsg}
	}
	SendOptions(client, optionData)
	return nil
}

// Handle the character preview. Will either return information about a character given
// a particular slot in a 0xE5 response or indicate an empty slot via 0xE4.
func handleCharacterSelect(client *LoginClient) error {
	archondb := GetConfig().Database()
	var pkt CharPreviewRequestPacket
	util.StructFromBytes(client.recvData[:8], &pkt)

	var charData CharacterData
	err := archondb.QueryRow("SELECT character_data from characters "+
		"where guildcard = ? and slot_num = ?", client.guildcard, pkt.Slot).Scan(&charData)
	if err == sql.ErrNoRows {
		// We don't have a character for this slot - send the E4 ack.
		SendCharPreviewNone(client, pkt.Slot)
		return nil
	} else if err != nil {
		errMsg := fmt.Sprintf("SQL Error: %s", err.Error())
		LogMsg(errMsg, LogTypeError, LogPriorityCritical)
		return &util.ServerError{Message: errMsg}
	} else {
		// We've got a match - send the character preview.
		// TODO: Send E5
	}

	return nil
}

// Process packets sent to the CHARACTER port by sending them off to another
// handler or by taking some brief action.
func processCharacterPacket(client *LoginClient) error {
	var pktHeader BBPktHeader
	util.StructFromBytes(client.recvData[:BBHeaderSize], &pktHeader)

	if GetConfig().DebugMode {
		fmt.Printf("Got %v bytes from client:\n", pktHeader.Size)
		util.PrintPayload(client.recvData, int(pktHeader.Size))
		fmt.Println()
	}

	var err error = nil
	switch pktHeader.Type {
	case LoginType:
		err = handleCharLogin(client)
	case DisconnectType:
		// Just wait until we recv 0 from the client to d/c.
		break
	case OptionsRequestType:
		handleKeyConfig(client)
	case CharPreviewReqType:
		handleCharacterSelect(client)
	case ChecksumType:
		// Everybody else seems to ignore this, so...
		SendChecksumAck(client, 1)
	default:
		msg := fmt.Sprintf("Received unknown packet %x from %s", pktHeader.Type, client.ipAddr)
		LogMsg(msg, LogTypeInfo, LogPriorityMedium)
	}
	return err
}

// Handle communication with a particular client until the connection is closed or an
// error is encountered.
func handleCharacterClient(client *LoginClient) {
	defer func() {
		if err := recover(); err != nil {
			errMsg := fmt.Sprintf("Error in client communication: %s: %s\n", client.ipAddr, err)
			LogMsg(errMsg, LogTypeError, LogPriorityHigh)
		}
		client.conn.Close()
		charConnections.RemoveClient(client)
		LogMsg("Disconnected CHARACTER client "+client.ipAddr, LogTypeInfo, LogPriorityMedium)
	}()

	LogMsg("Accepted CHARACTER connection from "+client.ipAddr, LogTypeInfo, LogPriorityMedium)
	// We're running inside a goroutine at this point, so we can block on this connection
	// and not interfere with any other clients.
	for {
		// Wait for the packet header.
		for client.recvSize < BBHeaderSize {
			bytes, err := client.conn.Read(client.recvData[client.recvSize:])
			if bytes == 0 || err == io.EOF {
				// The client disconnected, we're done.
				client.conn.Close()
				return
			} else if err != nil {
				// Socket error, nothing we can do now
				LogMsg("Socket Error ("+client.ipAddr+") "+err.Error(),
					LogTypeWarning, LogPriorityMedium)
				return
			}

			client.recvSize += bytes
			if client.recvSize >= BBHeaderSize {
				// We have our header; decrypt it.
				client.clientCrypt.Decrypt(client.recvData[:BBHeaderSize], BBHeaderSize)
				client.packetSize, err = util.GetPacketSize(client.recvData[:2])
				if err != nil {
					// Something is seriously wrong if this causes an error. Bail.
					panic(err.Error())
				}
			}
		}

		// Wait until we have the entire packet.
		for client.recvSize < int(client.packetSize) {
			bytes, err := client.conn.Read(client.recvData[client.recvSize:])
			if err != nil {
				panic(err.Error())
			}
			client.recvSize += bytes
		}

		// We have the whole thing; decrypt the rest of it if needed and pass it along.
		if client.packetSize > BBHeaderSize {
			client.clientCrypt.Decrypt(
				client.recvData[BBHeaderSize:client.packetSize],
				uint32(client.packetSize-BBHeaderSize))
		}
		if err := processCharacterPacket(client); err != nil {
			break
		}

		// Alternatively, we could set the slice to to nil here and make() a new one in order
		// to allow the garbage collector to handle cleanup, but I expect that would have a
		// noticable impact on performance. Instead, we're going to clear it manually.
		util.ZeroSlice(client.recvData, client.recvSize)
		client.recvSize = 0
		client.packetSize = 0
	}
}

// Main worker thread for the CHARACTER portion of the server.
func StartCharacter(wg *sync.WaitGroup) {
	loginConfig := GetConfig()
	socket, err := util.OpenSocket(loginConfig.Hostname, loginConfig.CharacterPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Printf("Waiting for CHARACTER connections on %s:%s...\n\n",
		loginConfig.Hostname, loginConfig.CharacterPort)

	for {
		connection, err := socket.AcceptTCP()
		if err != nil {
			LogMsg("Failed to accept connection: "+err.Error(), LogTypeError, LogPriorityHigh)
			continue
		}
		client, err := NewClient(connection)
		if err != nil {
			continue
		}
		charConnections.AddClient(client)
		go handleCharacterClient(client)
	}
	wg.Done()
}
