package london

import (
	"bytes"
	"errors"
	"log"
)

//prepends STX byte and appends ETX byte
//@pre: slice does not contain STX nor ETX
//@post: slice begins with STX and ends with ETX bytes
func Wrap(message []byte) ([]byte, error) {

	log.Printf("Wrapping message %X", message)

	if bytes.Contains(message, []byte{STX}) || bytes.Contains(message, []byte{ETX}) {
		return []byte{}, errors.New("Message contains erroneous STX or ETX byte")
	}

	stx := []byte{STX}
	message = append(stx, message...)
	message = append(message, ETX)
	return message, nil
}

//removes STX and ETX bytes
//@pre: slice begins with STX and ends with ETX
//@post: slice does not contain STX nor ETX bytes
func Unwrap(message []byte) ([]byte, error) {

	log.Printf("Unwrapping message %X", message)

	firstByte := []byte{message[0]}
	lastByte := []byte{message[len(message)-1]}

	if !bytes.Equal([]byte{STX}, firstByte) {
		return []byte{}, errors.New("Message does not begin with STX byte")
	} else if !bytes.Equal([]byte{ETX}, lastByte) {
		return []byte{}, errors.New("Message does not end with ETX byte")
	}

	message = bytes.TrimPrefix(message, []byte{STX})
	message = bytes.TrimSuffix(message, []byte{ETX})

	if bytes.Contains(message, []byte{STX}) || bytes.Contains(message, []byte{ETX}) {
		return []byte{}, errors.New("Message contains erroneous STX or ETX byte")
	}

	return message, nil
}

//@pre: STX and ETX removed, subsitutions made for reserved bytes
//@post: checksum removed
//generates a checksum byte and compares it to the checksum supplied a checksum that does not match returns an error
func Validate(message []byte) ([]byte, error) {

	log.Printf("Validating  message %X", message)

	checksum := GetChecksumByte(message[:len(message)-1])
	if checksum != message[len(message)-1] {
		return []byte{}, errors.New("checksums do not match")
	}

	message = bytes.TrimSuffix(message, []byte{checksum})

	return message, nil
}

//generates a checksum byte according to the exclusive or of all bytes in the message
func GetChecksumByte(message []byte) byte {

	log.Printf("Generating checksum byte for message %X...", message)

	checksum := message[0] ^ message[1]

	for i := 2; i < len(message); i++ {
		checksum = checksum ^ message[i]
	}

	log.Printf("checksum: %X", checksum)
	return checksum
}

func MakeSubstitutions(command []byte, toCheck map[string]int) ([]byte, error) {

	log.Printf("Making substitutions for message %X...", command)

	//always address escape byte first
	escapeInt := toCheck["escape"]
	var toReplace []byte
	toReplace = append(toReplace, byte(escapeInt))

	if len(substitutions[escapeInt]) == 1 {
		//get the second bit
		newEscapeInt := escapeInt >> 8
		toReplace = append([]byte{byte(newEscapeInt)}, toReplace...)
	}

	newCommand := bytes.Replace(command, toReplace, substitutions[escapeInt], -1)

	for key, value := range toCheck {

		if key == "escape" {
			continue
		}

		var iHateYou []byte
		iHateYou = append(iHateYou, byte(value))

		if len(substitutions[value]) == 1 {
			//get the second bit
			newEscapeInt := value >> 8
			iHateYou = append([]byte{byte(newEscapeInt)}, iHateYou...)
		}

		newCommand = bytes.Replace(newCommand, iHateYou, substitutions[value], -1)

	}

	return newCommand, nil
}
