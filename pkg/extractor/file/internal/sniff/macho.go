/*
 * Copyright (c) 2018-2019 vChain, Inc. All Rights Reserved.
 * This software is released under GPL3.
 * The full license information can be found under:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 */

package sniff

import (
	"debug/macho"
	"os"
	"strings"
)

const Platform_MachO = "Mach"

func MachO(file *os.File) (*Data, error) {
	f, err := macho.NewFile(file)
	if err != nil {
		return nil, err
	}

	cpu := strings.TrimPrefix(f.Cpu.String(), "Cpu")

	d := &Data{
		Type:     f.Type.String(),
		Platform: Platform_MachO,
		Arch:     cpu,
		X64:      strings.HasSuffix(cpu, "64"),
	}
	return d, nil
}
