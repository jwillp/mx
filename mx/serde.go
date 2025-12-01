package mx

import (
	"encoding/json"
	"fmt"
	"github.com/morebec/misas/misas"
	"reflect"
)

var (
	CommandRegistry = newMessageRegistry[misas.CommandTypeName, misas.Command]()
	EventRegistry   = newMessageRegistry[misas.EventTypeName, misas.Event]()
	QueryRegistry   = newMessageRegistry[misas.QueryTypeName, misas.Query]()
)

type MessageRegistry[TN ~string, T any] struct {
	messages map[TN]reflect.Type
}

func newMessageRegistry[TN ~string, T any]() *MessageRegistry[TN, T] {
	return &MessageRegistry[TN, T]{
		messages: make(map[TN]reflect.Type),
	}
}

func (m *MessageRegistry[TN, T]) Register(tn TN, prototype T) {
	t := reflect.TypeOf(prototype)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // store underlying type
	}
	m.messages[tn] = t
}

func (m *MessageRegistry[TN, T]) Clear() { m.messages = make(map[TN]reflect.Type, len(m.messages)) }

func (m *MessageRegistry[TN, T]) UnmarshalFromJSON(tn TN, js []byte) (T, error) {
	var zero T
	typ, err := m.resolve(tn)
	if err != nil {
		return zero, err
	}

	// Make a pointer to the type
	ptr := reflect.New(typ)
	if err := json.Unmarshal(js, ptr.Interface()); err != nil {
		return zero, fmt.Errorf("failed to unmarshal message %q: %w", tn, err)
	}

	// Deref pointer
	val := ptr.Elem().Interface()

	tVal, ok := val.(T)
	if !ok {
		panic(misas.ErrBadLogic.WithMessage(fmt.Sprintf("unmarshaled type %q does not implement target interface", typ.Name())))
	}

	return tVal, nil
}

func (m *MessageRegistry[TN, T]) resolve(tn TN) (reflect.Type, error) {
	typ, found := m.messages[tn]
	if !found {
		return nil, fmt.Errorf("unresolved message %q", tn)
	}
	return typ, nil
}

//type EventStoreDeserializerDecorator struct {
//	misas.EventStore
//}
//
//func NewEventStoreDeserializerDecorator(eventStore misas.EventStore) *EventStoreDeserializerDecorator {
//	if eventStore == nil {
//		panic(misas.ErrBadLogic.WithMessage("event store cannot be nil"))
//	}
//	return &EventStoreDeserializerDecorator{EventStore: eventStore}
//}
//
//func (d EventStoreDeserializerDecorator) ReadFromStream(ctx context.Context, streamID misas.EventStreamID, options misas.ReadFromEventStreamOptions) (misas.EventStreamSlice, error) {
//	stream, err := d.EventStore.ReadFromStream(ctx, streamID, options)
//	if err != nil {
//		return misas.EventStreamSlice{}, err
//	}
//
//	for i, record := range stream.Events {
//		jsonEvent, ok := record.Data.(misas.JSONEvent)
//		if !ok {
//			continue
//		}
//		event, err := EventRegistry.UnmarshalFromJSON(record.TypeName, jsonEvent.Data())
//		if err != nil {
//			return misas.EventStreamSlice{}, fmt.Errorf("failed to unmarshal event data for event %q: %w", record.TypeName, err)
//		}
//		stream.Events[i].Data = event
//	}
//
//	return stream, nil
//}
