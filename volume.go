package london

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/byuoitav/common/pooled"
	"github.com/fatih/color"
)

//GetVolume .
func (d *DSP) GetVolumeByBlock(ctx context.Context, block string) (int, error) {

	subscribe, err := d.BuildCommand(ctx, block, "volume", []byte{}, SubscribePercent)
	if err != nil {
		msg := fmt.Sprintf("unable to build subscribe command %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return 0, errors.New(msg)
	}

	unsubscribe, err := d.BuildCommand(ctx, block, "volume", []byte{}, UnsubscribePercent)
	if err != nil {
		msg := fmt.Sprintf("unable to build unsubscribe command %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return 0, errors.New(msg)
	}
	var response []byte
	work := func(conn pooled.Conn) error {
		response, err = d.GetStatus(ctx, subscribe, unsubscribe, conn)
		if err != nil {
			msg := fmt.Sprintf("Could not execute commands: %s", err.Error())
			log.Printf("%s", color.HiRedString("[error] %s", msg))
			return errors.New(msg)
		}
		return nil
	}

	err = pool.Do(d.Address, work)
	if err != nil {
		return 0, err
	}

	response, err = Unwrap(response)
	if err != nil {
		errorMessage := "Could not unwrap message: " + err.Error()
		log.Printf(errorMessage)
		return 0, errors.New(errorMessage)
	}

	response, err = MakeSubstitutions(response, DECODE)
	if err != nil {
		errorMessage := "Could not substitute reserved bytes: " + err.Error()
		log.Printf(errorMessage)
		return 0, errors.New(errorMessage)
	}

	response, err = Validate(response)
	if err != nil {
		errorMessage := "Invalid message: " + err.Error()
		log.Printf(errorMessage)
		return 0, errors.New(errorMessage)
	}

	volume, err := d.ParseVolumeStatus(ctx, response)
	if err != nil {
		errorMessage := "Could not parse response: " + err.Error()
		log.Printf(errorMessage)
		return 0, errors.New(errorMessage)
	}

	return volume, nil

}

//SetVolume gets an volume level (int) and sets it on device
func (d *DSP) SetVolumeByBlock(ctx context.Context, block string, volume int) error {

	command, err := d.BuildRawVolumeCommand(ctx, block, volume)
	if err != nil {
		return err
	}

	command, err = MakeSubstitutions(command, ENCODE)
	if err != nil {
		return err
	}

	command, err = Wrap(command)
	if err != nil {
		return err
	}

	work := func(conn pooled.Conn) error {
		n, err := conn.Write(command)
		switch {
		case err != nil:
			return fmt.Errorf("unable to send command: %s", err)
		case n != len(command):
			return fmt.Errorf("unable to send command: wrote %v/%v bytes", n, len(command))
		}
		return nil
	}

	err = pool.Do(d.Address, work)
	if err != nil {
		return err
	}

	return nil
}

//GetMute .
func (d *DSP) GetMutedByBlock(ctx context.Context, block string) (bool, error) {

	log.Printf("%s", color.HiMagentaString("[status] getting mute status of channel %X from device at address %s", block, d.Address))

	subscribe, err := d.BuildCommand(ctx, block, "mute", []byte{}, Subscribe)
	if err != nil {
		msg := fmt.Sprintf("unable to build subscribe command %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return false, errors.New(msg)
	}

	unsubscribe, err := d.BuildCommand(ctx, block, "mute", []byte{}, Unsubscribe)
	if err != nil {
		msg := fmt.Sprintf("unable to build unsubscribe command %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return false, errors.New(msg)
	}
	var response []byte
	work := func(conn pooled.Conn) error {
		response, err = d.GetStatus(ctx, subscribe, unsubscribe, conn)
		if err != nil {
			errorMessage := "could not execute commands: " + err.Error()
			log.Printf(errorMessage)
			return errors.New(errorMessage)
		}
		return nil
	}

	err = pool.Do(d.Address, work)
	if err != nil {
		return false, err
	}

	response, err = Unwrap(response)
	if err != nil {
		errorMessage := "Could not unwrap message: " + err.Error()
		log.Printf(errorMessage)
		return false, errors.New(errorMessage)
	}

	response, err = MakeSubstitutions(response, DECODE)
	if err != nil {
		errorMessage := "Could not substitute reserved bytes: " + err.Error()
		log.Printf(errorMessage)
		return false, errors.New(errorMessage)
	}

	response, err = Validate(response)
	if err != nil {
		errorMessage := "Invalid message: " + err.Error()
		log.Printf(errorMessage)
		return false, errors.New(errorMessage)
	}

	state, err := d.ParseMuteStatus(ctx, response)
	if err != nil {
		errorMessage := "Could not parse response: " + err.Error()
		log.Printf(errorMessage)
		return false, errors.New(errorMessage)
	}

	log.Printf("%s", color.HiMagentaString("[status] successfully retrieved status"))

	return state, nil

}

func (d *DSP) SetMutedByBlock(ctx context.Context, block string, muted bool) error {
	var mute string
	if muted == true {
		mute = "true"
	} else {
		mute = "false"
	}

	command, err := d.BuildRawMuteCommand(ctx, block, mute)
	if err != nil {
		return err
	}

	command, err = MakeSubstitutions(command, ENCODE)
	if err != nil {
		return err
	}

	command, err = Wrap(command)
	if err != nil {
		return err
	}

	work := func(conn pooled.Conn) error {
		n, err := conn.Write(command)
		switch {
		case err != nil:
			return fmt.Errorf("unable to send command: %s", err)
		case n != len(command):
			return fmt.Errorf("unable to send command: wrote %v/%v bytes", n, len(command))
		}
		return nil
	}

	err = pool.Do(d.Address, work)
	if err != nil {
		return err
	}

	return nil
}

//BuildCommand .
func (d *DSP) BuildCommand(ctx context.Context, input, status string, data []byte, method Method) ([]byte, error) {

	log.Printf("[command] building command...")

	command, err := d.BuildRawCommand(ctx, input, status, data, method)
	if err != nil {
		msg := fmt.Sprintf("could not build subscribe command: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	command, err = MakeSubstitutions(command, ENCODE)
	if err != nil {
		msg := fmt.Sprintf("Could not substitute reserved bytes: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	command, err = Wrap(command)
	if err != nil {
		msg := fmt.Sprintf("Could not wrap message: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	return command, nil
}

//BuildRawCommand .
func (d *DSP) BuildRawCommand(ctx context.Context, input, state string, data []byte, method Method) ([]byte, error) {

	log.Printf("Building subscription message for %s on input %s at address %s", state, input, d.Address)

	var base byte
	switch method {
	case Set:
		base = DI_SETSV
	case SetPercent:
		base = DI_SETSVPERCENT
	case Subscribe:
		base = DI_SUBSCRIBESV
	case Unsubscribe:
		base = DI_UNSUBSCRIBESV
	case SubscribePercent:
		base = DI_SUBSCRIBESVPERCENT
	case UnsubscribePercent:
		base = DI_UNSUBSCRIBESVPERCENT
	}

	command, err := d.GetCommandAddress(ctx, base)
	if err != nil {
		msg := fmt.Sprintf("could not address command: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	gainBlock, err := hex.DecodeString(input)
	if err != nil {
		msg := fmt.Sprintf("Could not decode input string: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	command = append(command, gainBlock...)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	command = append(command, stateVariables[state]...)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	if method == Set || method == SetPercent {
		command = append(command, data...)
	} else if method == Subscribe || method == SubscribePercent {
		command = append(command, RATE...)
	} else { // it's an unsubscribe command and the rate is zero. I hope this works
		zero := make([]byte, len(RATE))
		command = append(command, zero...)
	}

	log.Printf("Command string: %s", hex.EncodeToString(command))

	checksum := GetChecksumByte(command)
	command = append(command, checksum)
	log.Printf("Command string: %s", hex.EncodeToString(command))

	return command, nil
}

//GetStatus .
func (d *DSP) GetStatus(ctx context.Context, subscribe, unsubscribe []byte, pconn pooled.Conn) ([]byte, error) {

	log.Printf("[status] handling status command: %s...", color.HiMagentaString("%X", subscribe))

	log.Printf("[status] writing status command...")

	_, err := pconn.Write(subscribe)
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to subscribe: %s", err)
	}

	response, err := pconn.ReadUntil(ETX, 3*time.Second)
	log.Printf("[status] reading status response...")
	if err != nil {
		msg := fmt.Sprintf("device not responding: %s", err.Error())
		log.Printf("%s", color.HiRedString("[error] %s", msg))
		return []byte{}, errors.New(msg)
	}

	log.Printf("[status] response: %s", color.HiBlueString("%x", response))
	log.Printf("[status] sending unsubscribe command: %s...", color.HiBlueString("%x", unsubscribe))

	_, err = pconn.Write(unsubscribe)
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to unsubscribe: %s", err)
	}

	return response, nil
}

//@pre: checksum byte removed
func (d *DSP) ParseVolumeStatus(ctx context.Context, message []byte) (int, error) {

	log.Printf("Parsing mute status message: %X", message)

	//get data - always 4 bytes
	data := message[len(message)-4:]
	log.Printf("data: %X", data)
	log.Printf("len(data): %v", len(data))

	//turn data into number between 0 and 100
	const SCALE_FACTOR = 65536
	var rawValue int32
	_ = binary.Read(bytes.NewReader(data), binary.BigEndian, &rawValue)
	log.Printf("rawValue %v", rawValue)

	trueValue := rawValue / SCALE_FACTOR

	trueValue++ //not sure why it comes up with the wrong number

	return int(trueValue), nil
}

//@pre: checksum byte removed
func (d *DSP) ParseMuteStatus(ctx context.Context, message []byte) (bool, error) {

	log.Printf("Parsing mute status message: %X", message)

	//mute status determined with last byte
	data := message[len(message)-1:]
	log.Printf("data: %X", data)
	if bytes.EqualFold(data, []byte{0}) {
		return false, nil
	} else if bytes.EqualFold(data, []byte{1}) {
		return true, nil
	} else { //bad data
		return false, errors.New("Bad data in status message")
	}
}
