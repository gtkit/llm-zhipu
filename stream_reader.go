package zhipu

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gtkit/llm-zhipu/utils"
)

var (
	// errorPrefix = []byte(`data: {"error":`).
	headerID    = []byte("id:")
	headerData  = []byte("data:")
	headerEvent = []byte("event:")
	headerMeta  = []byte("meta:")
)

type StreamReaderInterface[T streamable] interface {
	Recv() (response T, err error)
	Close()
	processLines() (T, error)
	unmarshalError() (errResp *ErrorResponse)
}

type streamable interface {
	GlmChatCompletionStreamResponse
}

type streamReader[T streamable] struct {
	isFinished bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler
}

var _ StreamReaderInterface[GlmChatCompletionStreamResponse] = (*streamReader[GlmChatCompletionStreamResponse])(nil)

type Event struct {
	ID    []byte
	Data  []byte
	Event []byte
	Meta  []byte
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Event{}
	},
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

func (stream *streamReader[T]) processLines() (T, error) {
	var (
		hasErrorPrefix bool
		response       T
	)

	event, ok := pool.Get().(*Event)
	if !ok {
		return *new(T), nil
	}

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')

		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return *new(T), fmt.Errorf("error, %w", respErr.Error)
			}
			if errors.Is(readErr, context.Canceled) {
				return *new(T), readErr
			}
			return *new(T), fmt.Errorf("stream read error, %w", readErr)
		}

		if bytes.Equal(rawLine, []byte("\n")) {
			meta := &GlmMeta{}
			if len(event.Meta) > 0 {
				if err := json.Unmarshal(event.Meta, meta); err != nil {
					log.Println("---Meta Unmarshal error:", err)
				}
			}

			response = T{
				ID:    string(event.ID),
				Event: string(event.Event),
				Choices: []ChatCompletionStreamChoice{
					{
						Delta: ChatCompletionStreamChoiceDelta{
							Content: string(event.Data),
						},
					},
				},
				Meta: *meta,
			}

			putEvent(event)

			return response, nil
		}

		e, _ := processEvent(rawLine)

		if e.Event != nil {
			if bytes.Equal(e.Event, []byte("finish")) {
				stream.isFinished = true
			}
			event.Event = e.Event
		}
		if e.ID != nil {
			event.ID = e.ID
		}
		if e.Data != nil {
			event.Data = append(event.Data, e.Data...)
		}
		if e.Meta != nil {
			event.Meta = e.Meta
		}
	}
}

func putEvent(e *Event) {
	e.ID = nil
	e.Event = nil
	e.Data = nil
	pool.Put(e)
}

func (stream *streamReader[T]) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}

func processEvent(msg []byte) (event *Event, err error) {
	var e Event

	if len(msg) < 1 {
		return nil, errors.New("event message was empty")
	}

	switch {
	case bytes.HasPrefix(msg, headerID):
		e.ID = append([]byte(nil), trimHeader(len(headerID), msg)...)
	case bytes.HasPrefix(msg, headerData):
		e.Data = append(e.Data, trimHeader(len(headerData), msg)...)
		if bytes.Equal(msg, []byte("data:   \n")) {
			e.Data = append(e.Data, byte('\n'))
		}
	// The spec says that a line that simply contains the string "data" should be treated
	// as a data field with an empty body.
	case bytes.Equal(msg, bytes.TrimSuffix(headerData, []byte(":"))):
		e.Data = append(e.Data, byte('\n'))
	case bytes.HasPrefix(msg, headerEvent):
		e.Event = append([]byte(nil), trimHeader(len(headerEvent), msg)...)
	case bytes.HasPrefix(msg, headerMeta):
		e.Meta = append([]byte(nil), trimHeader(len(headerMeta), msg)...)
	default:
		// Ignore any garbage that doesn't match what we're looking for.
	}

	// Trim the last "\n" per the spec.
	if len(e.Data) > 1 {
		e.Data = bytes.TrimSuffix(e.Data, []byte("\n"))
	}

	return &e, err
}

func trimHeader(size int, data []byte) []byte {
	if data == nil || len(data) < size {
		return data
	}
	data = data[size:]
	// Remove optional leading whitespace
	if len(data) > 0 && data[0] == 32 {
		data = data[1:]
	}
	return data
}
