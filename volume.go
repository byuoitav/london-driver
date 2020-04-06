package london

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/byuoitav/connpool"
)

const (
	volumeScaleFactor = 65536
)

// GetVolumeByBlock returns the volume [0, 100] of the given block.
func (d *DSP) GetVolumeByBlock(ctx context.Context, block string) (int, error) {
	subscribe, err := buildSubscribeCommand(methodSubscribePercent, stateGain, block, minSubscribeInterval)
	if err != nil {
		return 0, fmt.Errorf("unable to build subscribe command: %w", err)
	}

	unsubscribe, err := buildUnsubscribeCommand(methodUnsubscribePercent, stateGain, block)
	if err != nil {
		return 0, fmt.Errorf("unable to build unsubscribe command: %w", err)
	}

	var resp []byte

	err = d.pool.Do(ctx, func(conn connpool.Conn) error {
		d.infof("Getting volume on %v", block)
		d.debugf("Writing subscribe command: 0x%x", subscribe)

		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))

		n, err := conn.Write(subscribe)
		switch {
		case err != nil:
			return fmt.Errorf("unable to write subscribe command: %w", err)
		case n != len(subscribe):
			return fmt.Errorf("unable to write subscribe command: wrote %v/%v bytes", n, len(subscribe))
		}

		resp, err = conn.ReadUntil(asciiETX, 3*time.Second)
		if err != nil {
			return fmt.Errorf("unable to read response: %w", err)
		}

		d.debugf("Got response: 0x%x", resp)
		d.debugf("Writing unsubscribe command: 0x%x", unsubscribe)

		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))

		n, err = conn.Write(unsubscribe)
		switch {
		case err != nil:
			return fmt.Errorf("unable to write unsubscribe command: %w", err)
		case n != len(unsubscribe):
			return fmt.Errorf("unable to write unsubscribe command: wrote %v/%v bytes", n, len(subscribe))
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	resp, err = decode(resp)
	if err != nil {
		return 0, fmt.Errorf("unable to decode response: %w", err)
	}

	data := resp[len(resp)-4:]
	vol := binary.BigEndian.Uint32(data)

	vol = vol / volumeScaleFactor
	vol++

	d.infof("Volume on %v is %v", block, int(vol))

	return int(vol), nil
}

// SetVolumeByBlock sets the volume on the given block. Volume must be in the range [0, 100].
func (d *DSP) SetVolumeByBlock(ctx context.Context, block string, volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be in range [0, 100]")
	}

	volume *= volumeScaleFactor
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(volume))

	cmd, err := buildCommand(methodSetPercent, stateGain, block, data)
	if err != nil {
		return fmt.Errorf("unable to build command: %w", err)
	}

	err = d.pool.Do(ctx, func(conn connpool.Conn) error {
		d.infof("Setting volume on %v to %v", block, volume)
		d.debugf("Writing command: 0x%x", cmd)

		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))

		n, err := conn.Write(cmd)
		switch {
		case err != nil:
			return fmt.Errorf("unable to write command: %w", err)
		case n != len(cmd):
			return fmt.Errorf("unable to write command: wrote %v/%v bytes", n, len(cmd))
		}

		return nil
	})
	if err != nil {
		return err
	}

	d.infof("Volume on %v successfully set", block)

	return nil
}
