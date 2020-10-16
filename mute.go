package london

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byuoitav/connpool"
)

// GetMutedByBlock returns true if the given block is muted.
func (d *DSP) GetMutes(ctx context.Context, blocks []string) (map[string]bool, error) {
	toReturn := make(map[string]bool)

	for _, block := range blocks {
		subscribe, err := buildSubscribeCommand(methodSubscribe, stateMute, block, minSubscribeInterval)
		if err != nil {
			return toReturn, fmt.Errorf("unable to build subscribe command: %w", err)
		}

		unsubscribe, err := buildUnsubscribeCommand(methodUnsubscribe, stateMute, block)
		if err != nil {
			return toReturn, fmt.Errorf("unable to build unsubscribe command: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var resp []byte

		err = d.pool.Do(ctx, func(conn connpool.Conn) error {
			d.infof("Getting mute on %v", block)
			d.debugf("Writing subscribe command: 0x%x", subscribe)

			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))

			n, err := conn.Write(subscribe)
			switch {
			case err != nil:
				return fmt.Errorf("unable to write subscribe command: %w", err)
			case n != len(subscribe):
				return fmt.Errorf("unable to write subscribe command: wrote %v/%v bytes", n, len(subscribe))
			}

			deadline, ok := ctx.Deadline()
			if !ok {
				return fmt.Errorf("no deadline set")
			}

			resp, err = conn.ReadUntil(asciiETX, deadline)
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
			return toReturn, err
		}

		resp, err = decode(resp)
		if err != nil {
			return toReturn, fmt.Errorf("unable to decode response: %w", err)
		}

		data := resp[len(resp)-1:]

		switch {
		case bytes.Equal(data, []byte{0}):
			d.infof("Mute on %v is %v", block, false)
			toReturn[block] = false
		case bytes.Equal(data, []byte{1}):
			d.infof("Mute on %v is %v", block, true)
			toReturn[block] = true
		default:
			return toReturn, errors.New("bad data in response from DSP")
		}
	}

	return toReturn, nil
}

// SetMutedByBlock sets the mute state on the given block.
func (d *DSP) SetMute(ctx context.Context, block string, mute bool) error {
	data := []byte{0x00, 0x00, 0x00, 0x00}
	if mute {
		data[3] = 0x01
	}

	cmd, err := buildCommand(methodSet, stateMute, block, data)
	if err != nil {
		return fmt.Errorf("unable to build command: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = d.pool.Do(ctx, func(conn connpool.Conn) error {
		d.infof("Setting mute on %v to %v", block, mute)
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

	d.infof("Mute on %v successfully set", block)

	return nil
}
