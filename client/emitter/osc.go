package emitter

import (
	"fmt"
	"math"
	"sort"
	"time"

	"alda.io/client/model"
	"github.com/daveyarwood/go-osc/osc"
)

// OSCEmitter sends OSC messages to a player process.
type OSCEmitter struct {
	Port int
}

func systemPlayMsg() *osc.Message {
	return osc.NewMessage("/system/play")
}

func systemStopMsg() *osc.Message {
	return osc.NewMessage("/system/stop")
}

func systemShutdownMsg(offset int32) *osc.Message {
	msg := osc.NewMessage("/system/shutdown")
	msg.Append(offset)
	return msg
}

func midiPatchMsg(track int32, offset int32, patch int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/patch", track))
	msg.Append(offset)
	msg.Append(patch)
	return msg
}

func midiPercussionMsg(track int32, offset int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/percussion", track))
	msg.Append(offset)
	return msg
}

func midiNoteMsg(
	track int32, offset int32, note int32, duration int32, audibleDuration int32,
	velocity int32,
) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/note", track))
	msg.Append(offset)
	msg.Append(note)
	msg.Append(duration)
	msg.Append(audibleDuration)
	msg.Append(velocity)
	return msg
}

func midiVolumeMsg(track int32, offset int32, volume int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/volume", track))
	msg.Append(offset)
	msg.Append(volume)
	return msg
}

func midiPanningMsg(track int32, offset int32, panning int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/panning", track))
	msg.Append(offset)
	msg.Append(panning)
	return msg
}

func oscClient(port int) *osc.Client {
	return osc.NewClient("localhost", int(port), osc.ClientProtocol(osc.TCP))
}

// EmitStopMessage sends a "stop" message to a player process.
func (oe OSCEmitter) EmitStopMessage() error {
	return oscClient(oe.Port).Send(systemStopMsg())
}

// EmitShutdownMessage sends a "shutdown" message to a player process.
func (oe OSCEmitter) EmitShutdownMessage(offset int32) error {
	return oscClient(oe.Port).Send(systemShutdownMsg(offset))
}

// EmitScore implements Emitter.EmitScore by sending OSC messages to instruct a
// player process how to perform the score.
func (oe OSCEmitter) EmitScore(
	score *model.Score, opts ...EmissionOption,
) error {
	ctx := &EmissionContext{}
	for _, opt := range opts {
		opt(ctx)
	}

	startOffset := 0.0
	endOffset := math.MaxFloat64

	if ctx.from != "" {
		offset, err := score.InterpretOffsetReference(ctx.from)
		if err != nil {
			return err
		}

		startOffset = offset
	}

	if ctx.to != "" {
		offset, err := score.InterpretOffsetReference(ctx.to)
		if err != nil {
			return err
		}

		endOffset = offset
	}

	bundle := osc.NewBundle(time.Now())

	// In order to support features like:
	//
	// * Avoiding scheduling more volume and panning control change messages than
	//   we have to (see below).
	//
	// * Playing just a slice of a score, e.g. `alda play --from 0:05 --to 0:10`
	//
	// ...we sort the events in the score by offset and schedule them in
	// chronological order.
	sort.Slice(score.Events, func(i, j int) bool {
		return score.Events[i].EventOffset() < score.Events[j].EventOffset()
	})

	// In Alda's model, the track volume and panning are an attribute of each
	// individual note. However, in MIDI, these attributes are set persistently on
	// a channel via a control change message.
	//
	// To make this work, as we're scheduling the events of the score in
	// chonological order, we keep track of the volume and panning attributes for
	// each track, so that we can send volume and panning control changes only
	// when necessary (when the values change).
	currentVolume := map[int32]float64{}
	currentPanning := map[int32]float64{}

	tracks := score.Tracks()

	for part, trackNumber := range tracks {
		currentVolume[trackNumber] = -1
		currentPanning[trackNumber] = -1

		// We currently only have MIDI instruments. This might change in the future,
		// which is why Instrument is an interface instead of a plain struct. For
		// now, we're operating under the assumption that all instruments are MIDI
		// instruments.
		stockInstrument := part.StockInstrument.(model.MidiInstrument)

		patchNumber := stockInstrument.PatchNumber
		bundle.Append(midiPatchMsg(trackNumber, 0, patchNumber))

		if stockInstrument.IsPercussion {
			bundle.Append(midiPercussionMsg(trackNumber, 0))
		}
	}

	for _, event := range score.Events {
		eventOffset := event.EventOffset()

		// Filter out events before the `--from` time marking / marker, when
		// supplied.
		if eventOffset < startOffset {
			continue
		}

		// Filter out events after the `--to` time marking / marker, when supplied.
		if eventOffset >= endOffset {
			break
		}

		switch event.(type) {
		case model.NoteEvent:
			noteEvent := event.(model.NoteEvent)
			track := tracks[noteEvent.Part]

			// We subtract `startOffset` from the offset so that when the `--from`
			// option is used (e.g. `--from 0:30`), we will shift all of the events
			// back by that amount so that playback starts as if those events were at
			// the beginning of the score. Otherwise, `--from 0:30` would result in
			// you having to wait 30 seconds before you hear anything.
			//
			// By default, `startOffset` is 0, so the usual scenario is that the event
			// offsets are not adjusted.
			offset := int32(math.Round(noteEvent.Offset - startOffset))

			if noteEvent.TrackVolume != currentVolume[track] {
				currentVolume[track] = noteEvent.TrackVolume

				bundle.Append(
					midiVolumeMsg(
						track,
						offset,
						int32(math.Round(noteEvent.TrackVolume*127)),
					),
				)
			}

			if noteEvent.Panning != currentPanning[track] {
				currentPanning[track] = noteEvent.Panning

				bundle.Append(
					midiPanningMsg(
						track,
						offset,
						int32(math.Round(noteEvent.Panning*127)),
					),
				)
			}

			bundle.Append(midiNoteMsg(
				track,
				offset,
				noteEvent.MidiNote,
				int32(math.Round(noteEvent.Duration)),
				int32(math.Round(noteEvent.AudibleDuration)),
				int32(math.Round(noteEvent.Volume*127)),
			))
		default:
			return fmt.Errorf("unsupported event: %#v", event)
		}
	}

	bundle.Append(systemPlayMsg())

	return oscClient(oe.Port).Send(bundle)
}