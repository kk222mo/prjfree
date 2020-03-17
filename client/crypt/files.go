package crypt

import (
	"bytes"
	"io/ioutil"
	"math"
	"prjfree/client/models"
)

func FileToBlocks(path string) []models.Block {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	return DataToBlocks(data)
}

func DataToBlocks(data []byte) []models.Block {
	var blocks []models.Block
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	data = bytes.ReplaceAll(data, []byte{'%', '%'}, []byte{'\n'})
	block_count := int(math.Ceil(float64(len(data)) / models.BLOCK_SIZE))
	blocks = make([]models.Block, block_count)
	for i := 0; i < block_count; i++ {
		b := models.Block{
			Data: data[i*models.BLOCK_SIZE : int(math.Min(float64(len(data)), float64(i+1)*models.BLOCK_SIZE))],
		}
		blocks[i] = b
	}
	return blocks
}
