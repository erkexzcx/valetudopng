package decoder

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
)

func Decode(payload []byte) ([]byte, error) {
	var err error
	if isPNG(payload) {
		payload, err = extractZtxtValetudoMapPngChunk(payload)
		if err != nil {
			err = errors.New("failed to extract Ztxt: " + err.Error())
			return nil, err
		}
	}

	if isCompressed(payload) {
		payload, err = inflateSync(payload)
		if err != nil {
			err = errors.New("failed to decompress: " + err.Error())
			return nil, err
		}
	}

	return payload, nil
}

func isPNG(data []byte) bool {
	return len(data) >= 8 &&
		data[0] == 0x89 &&
		data[1] == 0x50 &&
		data[2] == 0x4E &&
		data[3] == 0x47 &&
		data[4] == 0x0D &&
		data[5] == 0x0A &&
		data[6] == 0x1A &&
		data[7] == 0x0A
}

func extractZtxtValetudoMapPngChunk(data []byte) ([]byte, error) {
	ended := false
	idx := 8

	for idx < len(data) {
		// Read the length of the current chunk,
		// which is stored as a Uint32.
		length := binary.BigEndian.Uint32(data[idx : idx+4])
		idx += 4

		// Chunk includes name/type for CRC check (see below).
		chunk := make([]byte, length+4)
		copy(chunk, data[idx:idx+4])
		idx += 4

		// Get the name in ASCII for identification.
		name := string(chunk[:4])

		// The IEND header marks the end of the file,
		// so on discovering it break out of the loop.
		if name == "IEND" {
			ended = true
			break
		}

		// Read the contents of the chunk out of the main buffer.
		copy(chunk[4:], data[idx:idx+int(length)])
		idx += int(length)

		// Skip the CRC32.
		idx += 4

		// The chunk data is now copied to remove the 4 preceding
		// bytes used for the chunk name/type.
		chunkData := chunk[4:]

		if name == "zTXt" {
			i := 0
			keyword := ""

			for chunkData[i] != 0 && i < 79 {
				keyword += string(chunkData[i])
				i++
			}

			if keyword != "ValetudoMap" {
				continue
			}

			return chunkData[i+2:], nil
		}
	}

	if !ended {
		return nil, errors.New(".png file ended prematurely: no IEND header was found")
	}

	return nil, errors.New("no ValetudoMap chunk found in the PNG")
}

func isCompressed(data []byte) bool {
	return data[0x00] == 0x78
}

func inflateSync(data []byte) ([]byte, error) {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return out, nil
}
