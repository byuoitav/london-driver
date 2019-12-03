package london

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
)

const LEN_NODE = 2
const LEN_ADDR = 5

//builds command, node, virtual device, and object
func (d *DSP) GetCommandAddress(ctx context.Context, commandByte byte) ([]byte, error) {

	log.Printf("Addressing command %X to address %s...", commandByte, d.Address)
	command := []byte{commandByte}

	nodeBytes := make([]byte, 2)
	command = append(command, nodeBytes...)

	command = append(command, VIRTUAL_DEVICE)

	return command, nil
}

func (d *DSP) BuildRawMuteCommand(ctx context.Context, input, status string) ([]byte, error) {

	log.Printf("Building raw mute command for input: %s at address: %s", input, d.Address)

	command, err := d.GetCommandAddress(ctx, DI_SETSV)
	if err != nil {
		errorMessage := "Could not address command: " + err.Error()
		log.Printf(errorMessage)
		return []byte{}, errors.New(errorMessage)
	}

	gainBlock, err := hex.DecodeString(input)
	if err != nil {
		errorMessage := "Could not decode input string: " + err.Error()
		log.Printf(errorMessage)
		return []byte{}, errors.New(errorMessage)
	}

	command = append(command, gainBlock...)
	command = append(command, stateVariables["mute"]...)
	command = append(command, muteStates[status]...)

	checksum := GetChecksumByte(command)
	command = append(command, checksum)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	return command, nil
}

func (d *DSP) BuildRawVolumeCommand(ctx context.Context, input string, volume int) ([]byte, error) {
	log.Printf("Building raw volume command for input: %s on address: %s", input, d.Address)

	command, err := d.GetCommandAddress(ctx, DI_SETSVPERCENT)
	if err != nil {
		errorMessage := "Could not address command: " + err.Error()
		log.Printf(errorMessage)
		return []byte{}, errors.New(errorMessage)
	}

	gainBlock, err := hex.DecodeString(input)
	if err != nil {
		errorMessage := "Could not decode input string: " + err.Error()
		log.Printf(errorMessage)
		return []byte{}, errors.New(errorMessage)
	}

	command = append(command, gainBlock...)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	command = append(command, stateVariables["gain"]...)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	log.Printf("Calculating parameter for volume %s", volume)
	if volume > 100 || volume < 0 {
		return []byte{}, errors.New("Invalid volume request")
	}

	volume *= 65536
	log.Printf("toSend: %v", volume)

	hexValue := make([]byte, 4)
	binary.BigEndian.PutUint32(hexValue, uint32(volume))

	command = append(command, hexValue...)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	checksum := GetChecksumByte(command)

	command = append(command, checksum)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	return command, nil
}
