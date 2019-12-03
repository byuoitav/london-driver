package london

type RawDICommand struct {
	Address string `json:"address"`
	Port    string `json:"port"`
	Command string `json:"command"`
}

type RawDIResponse struct {
	Response []string `json:"response"`
}

//standardize file based on these values
var gainBlocks = map[string][]byte{

	"mic1":   {0x00, 0x01, 0x0a},
	"mic2":   {0x00, 0x01, 0x0b},
	"mic3":   {0x00, 0x01, 0x0c},
	"mic4":   {0x00, 0x01, 0x0d},
	"media1": {0x00, 0x01, 0x0e},
	"media2": {0x00, 0x01, 0x0f},
}

//standards in London documentation
var stateVariables = map[string][]byte{

	"gain":     {0x00, 0x00},
	"mute":     {0x00, 0x01},
	"polarity": {0x00, 0x02},
}

var muteStates = map[string][]byte{

	"true":  {0x00, 0x00, 0x00, 0x01},
	"false": {0x00, 0x00, 0x00, 0x00},
}

var test = map[bool][]byte{

	true:  {0x00, 0x00, 0x00, 0x01},
	false: {0x00, 0x00, 0x00, 0x00},
}

//VIRTUAL_DEVICE byte should be the same for all cases!
var VIRTUAL_DEVICE = byte(0x03)

// var PORT = "1023"

var RATE = []byte{0x00, 0x00, 0x00, 0x32} //represents 50 ms, the shortest interval

var ACK = byte(0x06)
var ETX = byte(0x03)
var STX = byte(0x02)

var ENCODE = map[string]int{
	"STX":    0x02,
	"ETX":    0x03,
	"ACK":    0x06,
	"NAK":    0x15,
	"escape": 0x1b,
}

var DECODE = map[string]int{
	"STX":    0x1b82,
	"ETX":    0x1b83,
	"ACK":    0x1b86,
	"NAK":    0x1b95,
	"escape": 0x1b9b,
}

var substitutions = map[int][]byte{

	0x02:   {0x1b, 0x82},
	0x03:   {0x1b, 0x83},
	0x06:   {0x1b, 0x86},
	0x15:   {0x1b, 0x95},
	0x1b:   {0x1b, 0x9b},
	0x1b82: {0x02},
	0x1b83: {0x03},
	0x1b86: {0x06},
	0x1b95: {0x15},
	0x1b9b: {0x1b},
}

type Method int

const (
	Set Method = 1 + iota
	SetPercent
	Subscribe
	Unsubscribe
	SubscribePercent
	UnsubscribePercent
)

const (
	DI_SETSV                = 0x88
	DI_SETSVPERCENT         = 0x8d
	DI_SUBSCRIBESV          = 0x89
	DI_SUBSCRIBESVPERCENT   = 0x8e
	DI_UNSUBSCRIBESV        = 0x8a
	DI_UNSUBSCRIBESVPERCENT = 0x8f
)

type State string

const (
	GAIN State = "gain"
	MUTE State = "mute"
)
