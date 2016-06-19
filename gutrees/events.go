package gutrees

import "github.com/influx6/gu/guevents"

// Event defines the functions required for a event object.
type Event interface {
	Event() string
	EventID() string
	SetEventID(string)

	Apply(Markup)

	Meta() guevents.EventMeta
	Tree() Markup
	SetTree(Markup)

	Fire(guevents.Event)

	Clone() Event
	StopPropagation() Event
	StopImmediatePropagation() Event
	PreventDefault() Event
}

// EventHandler provides a custom event handler which allows access to the
// markup producing the event.
type EventHandler func(guevents.Event, Markup)

// EventObject provide a meta registry for helps in registering events for dom markups
// which is translated to the nodes themselves
type EventObject struct {
	meta *guevents.EventMetable
	fx   EventHandler
	tree Markup
}

// NewEvent returns a event object that allows registering events to eventlisteners
func NewEvent(eventType, eventSelector string, fx EventHandler) *EventObject {
	ex := EventObject{
		fx: fx,
		meta: &guevents.EventMetable{
			EventType:   eventType,
			EventTarget: eventSelector,
		},
	}

	return &ex
}

// StopImmediatePropagation will return itself and set StopPropagation to true
func (e *EventObject) StopImmediatePropagation() Event {
	e.meta.ShouldStopImmediatePropagation = true
	return e
}

// StopPropagation will return itself and set StopPropagation to true
func (e *EventObject) StopPropagation() Event {
	e.meta.ShouldStopPropagation = true
	return e
}

// Meta returns the meta associated with the giving event.
func (e *EventObject) Meta() guevents.EventMeta {
	return e.meta
}

// PreventDefault will return itself and set PreventDefault to true
func (e *EventObject) PreventDefault() Event {
	e.meta.ShouldPreventDefault = true
	return e
}

// Fire calls the giving event watcher with the provided guevents.Event.
func (e *EventObject) Fire(evs guevents.Event) {
	if e.fx != nil {
		e.fx(evs, e.tree)
	}
}

// Event returns the type of  the event.
func (e *EventObject) Event() string {
	return e.meta.EventType
}

// EventID returns the current event target for the event.
func (e *EventObject) EventID() string {
	return e.meta.EventTarget
}

// Tree returns the current tree used for attachment by the event.
func (e *EventObject) Tree() Markup {
	return e.tree
}

// SetTree sets the current tree for which the events bind to.
func (e *EventObject) SetTree(m Markup) {
	e.tree = m
}

// SetEventID sets the ID for the event which it attaches to.
func (e *EventObject) SetEventID(id string) {
	e.meta.EventTarget = id
}

//Clone replicates the style into a unique instance
func (e *EventObject) Clone() Event {
	ce := EventObject{
		fx: e.fx,
		meta: &guevents.EventMetable{
			EventType:   e.meta.EventType,
			EventTarget: e.meta.EventTarget,
		},
	}

	return &ce
}

// Apply adds the event into the elements events lists
func (e *EventObject) Apply(ex Markup) {
	EventApplier.Apply(ex, e)
}

//==============================================================================

// EventApplier defines a package level event applier for events.
var EventApplier eventy

type eventy struct{}

// MarkupEventProvider defines an interface for MarkupEvents providers that
// allows structures to register events for themselves.
type MarkupEventProvider interface {
	MarkupState
	EventID() string
	AddEvent(Event)
}

// Apply adds the event into the elements events lists
func (eventy) Apply(ex Markup, ev Event) {
	if em, ok := ex.(MarkupEventProvider); ok {
		if em.AllowEvents() {
			if ev.EventID() == "" {
				ev.SetEventID(ex.EventID())
			}

			ev.SetTree(ex)
			em.AddEvent(ev)
		}
	}
}

//==============================================================================
