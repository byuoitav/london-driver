// package london

// import (
// 	"context"
// 	"log"
// 	"net"
// 	"time"

// 	"github.com/byuoitav/common/pooled"
// 	"github.com/fatih/color"
// )

// //TIMEOUT .
// const TIMEOUT = 5
// const PORT = "1023"

// var pool = pooled.NewMap(45*time.Second, 400*time.Millisecond, getConnection)

// //ReadWrite .
// type ReadWrite int

// //ReadWrite .
// const (
// 	Read ReadWrite = 1 + iota
// 	Write
// )

// //GetConnection .
// func (d *DSP) getConnection(ctx context.Context) (pooled.Conn, error) {
// 	log.Printf("%s", color.HiCyanString("[connection] getting connection to address on device %s", d.Address))

// 	conn, err := net.DialTimeout("tcp", d.Address+":"+PORT, 10*time.Second)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// read the NOKEY line
// 	pconn := pooled.Wrap(conn)

// 	return pconn, nil
// }
