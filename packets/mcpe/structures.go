package mcpe

import (
	"../raknet"
	"io"
)

type BlockCoordinates struct {
	X, Y, Z int32
}

type PlayerLocation struct {
	X, Y, Z float32
}

func (bc BlockCoordinates) Write(writer io.Writer) (err error) {
	err = raknet.WriteInt32(writer, bc.X)
	if err != nil {
		return
	}
	err = raknet.WriteInt32(writer, bc.Y)
	if err != nil {
		return
	}
	err = raknet.WriteInt32(writer, bc.Z)
	return
}

func (pl PlayerLocation) Write(writer io.Writer) (err error) {
	err = WriteFloat32(writer, pl.X)
	if err != nil {
		return
	}
	err = WriteFloat32(writer, pl.Y)
	if err != nil {
		return
	}
	err = WriteFloat32(writer, pl.Z)
	return
}

func NewBlockCoordinates(reader io.Reader) (bc BlockCoordinates, err error) {
	x, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}
	y, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}
	z, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}

	bc.X = x
	bc.Y = y
	bc.Z = z
	return
}

func NewPlayerLocation(reader io.Reader) (pl PlayerLocation, err error) {
	x, err := ReadFloat32(reader)
	if err != nil {
		return
	}
	y, err := ReadFloat32(reader)
	if err != nil {
		return
	}
	z, err := ReadFloat32(reader)
	if err != nil {
		return
	}

	pl.X = x
	pl.Y = y
	pl.Z = z
	return
}
